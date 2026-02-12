// Package collector 提供市场数据采集功能
package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"CryptoSentinel/internal/calculator"
	"CryptoSentinel/internal/model"
)

// binanceTickerResponse Binance ticker价格响应
type binanceTickerResponse struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

// CollectorV2 升级版数据采集器，支持多指标采集
type CollectorV2 struct {
	client    *http.Client
	proxyAddr string
	userAgent string
}

// NewCollectorV2 创建升级版数据采集器
func NewCollectorV2() *CollectorV2 {
	return &CollectorV2{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: "CryptoSentinel/1.0",
	}
}

// NewCollectorV2WithProxy 创建带代理的数据采集器
func NewCollectorV2WithProxy(proxyAddr string) *CollectorV2 {
	proxyURL, _ := url.Parse("http://" + proxyAddr)

	return &CollectorV2{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		},
		proxyAddr: proxyAddr,
		userAgent: "CryptoSentinel/1.0",
	}
}

// FetchAllIndicators 获取所有市场指标
// leverage: 当前账户杠杆率（由外部传入）
func (c *CollectorV2) FetchAllIndicators(leverage float64) (*model.MarketIndicators, error) {
	indicators := &model.MarketIndicators{
		Timestamp:       time.Now(),
		AccountLeverage: leverage,
		Source:          "Binance",
	}

	// 1. 获取AHR999指数
	if err := c.fetchAHR999(indicators); err != nil {
		return nil, fmt.Errorf("获取AHR999失败: %w", err)
	}

	// 2. 获取BTC和ETH价格（Mock或从其他API获取）
	if err := c.fetchPrices(indicators); err != nil {
		// 价格获取失败不阻断流程，使用Mock数据
		indicators.CurrentPriceBTC = 0
		indicators.CurrentPriceETH = 0
	}

	// 3. 获取MVRV-Z Score（Mock数据，预留接口）
	indicators.MVRVZScore = c.fetchMVRVZScore()

	// 4. 计算MA乘数状态（Mock数据，预留接口）
	indicators.MaMultiplierState = c.calculateMaMultiplierState(indicators.CurrentPriceBTC)

	// 5. 获取Pi周期状态（Mock数据，预留接口）
	indicators.PiCycleCross = c.fetchPiCycleStatus()

	// 6. 获取ETH回归带位置（Mock数据，预留接口）
	indicators.EthRegressionState = c.calculateEthRegressionState(indicators.CurrentPriceETH)

	return indicators, nil
}

// fetchAHR999 通过Binance K线数据自主计算AHR999指数
func (c *CollectorV2) fetchAHR999(indicators *model.MarketIndicators) error {
	var calc *calculator.AHR999Calculator
	if c.proxyAddr != "" {
		calc = calculator.NewAHR999CalculatorWithProxy(c.proxyAddr)
	} else {
		calc = calculator.NewAHR999Calculator()
	}

	result, err := calc.Calculate()
	if err != nil {
		return fmt.Errorf("计算AHR999失败: %w", err)
	}

	indicators.AHR999 = result.AHR999
	indicators.CurrentPriceBTC = result.CurrentPrice
	return nil
}

// fetchPrices 获取ETH价格（BTC价格已从AHR999计算中获得）
func (c *CollectorV2) fetchPrices(indicators *model.MarketIndicators) error {
	apiURL := "https://api.binance.com/api/v3/ticker/price?symbol=ETHUSDT"
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("创建ETH价格请求失败: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("获取ETH价格失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取ETH价格响应失败: %w", err)
	}

	var ticker binanceTickerResponse
	if err := json.Unmarshal(body, &ticker); err != nil {
		return fmt.Errorf("解析ETH价格响应失败: %w", err)
	}

	price, err := strconv.ParseFloat(ticker.Price, 64)
	if err != nil {
		return fmt.Errorf("解析ETH价格值失败: %w", err)
	}

	indicators.CurrentPriceETH = price
	return nil
}

// fetchMVRVZScore 获取MVRV-Z Score
func (c *CollectorV2) fetchMVRVZScore() float64 {
	// TODO: 接入真实MVRV-Z API（如Glassnode等）
	// 目前使用Mock数据：正常范围内的值
	return 2.5
}

// calculateMaMultiplierState 计算MA乘数状态
func (c *CollectorV2) calculateMaMultiplierState(btcPrice float64) model.MaMultiplierState {
	// TODO: 接入真实数据或计算730日均线
	// 简单Mock逻辑：
	// - 价格 < 20000: 熊市底部
	// - 价格 > 150000: 疯牛顶部
	// - 其他: 正常
	if btcPrice > 0 {
		if btcPrice < 20000 {
			return model.MaStateBearBottom
		}
		if btcPrice > 150000 {
			return model.MaStateBullTop
		}
	}
	return model.MaStateNormal
}

// fetchPiCycleStatus 获取Pi周期状态
func (c *CollectorV2) fetchPiCycleStatus() bool {
	// TODO: 接入真实Pi周期数据
	// 目前返回false（未死叉）
	return false
}

// calculateEthRegressionState 计算ETH回归带位置
func (c *CollectorV2) calculateEthRegressionState(ethPrice float64) model.EthRegressionState {
	// TODO: 接入真实回归带数据
	// 简单Mock逻辑：
	// - 价格 < 2000: 低估区
	// - 价格 > 5000: 高估区
	// - 其他: 中间区
	if ethPrice > 0 {
		if ethPrice < 2000 {
			return model.EthRegLower
		}
		if ethPrice > 5000 {
			return model.EthRegUpper
		}
	}
	return model.EthRegMiddle
}

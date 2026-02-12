// Package notifier 提供消息通知和格式化功能
package notifier

import (
	"fmt"
	"time"

	"CryptoSentinel/internal/model"
)

// FormatReportV2 生成简洁的周报格式
func FormatReportV2(indicators *model.MarketIndicators, signal *model.TradeSignal) string {
	date := time.Now().Format("2006-01-02")

	// 宏观定调
	macroTone := getMacroTone(signal)

	// AHR999 分析
	ahr999Section := formatAHR999Section(indicators.AHR999)

	// MVRV-Z 分析
	mvrvSection := formatMVRVSection(indicators.MVRVZScore)

	// ETH 分析
	ethSection := formatETHSection(indicators.EthRegressionState)

	// 安全检查
	safetySection := formatSafetySection(indicators)

	// 实时价格
	priceSection := formatPriceSection(indicators)

	// 执行建议
	actionSection := formatActionSection(signal)

	// 组装报告
	report := fmt.Sprintf(`🛡️ **CryptoSentinel %s**

📊 **宏观定调: %s**

%s

%s

%s

%s

%s

---------------------
%s`,
		date,
		macroTone,
		priceSection,
		ahr999Section,
		mvrvSection,
		ethSection,
		safetySection,
		actionSection,
	)

	return report
}

// formatPriceSection 格式化实时价格部分
func formatPriceSection(indicators *model.MarketIndicators) string {
	return fmt.Sprintf("**💲 实时价格**\n• BTC: `$%.2f`\n• ETH: `$%.2f`",
		indicators.CurrentPriceBTC, indicators.CurrentPriceETH)
}

// getMacroTone 获取宏观定调
func getMacroTone(signal *model.TradeSignal) string {
	if signal.IsHalted {
		if signal.ActionBTC == model.ActionSellAlert {
			return "🔴 逃顶警报"
		}
		return "⚠️ 风控熔断"
	}

	switch signal.ActionBTC {
	case model.ActionStrongBuy:
		return "🟢 贪婪抄底"
	case model.ActionDCABuy:
		return "🟢 适合定投"
	case model.ActionHold, model.ActionHoldCaution:
		return "🟡 持有观望"
	case model.ActionSell:
		return "🔴 逐步离场"
	default:
		return "🟡 中性"
	}
}

// formatAHR999Section 格式化AHR999部分
func formatAHR999Section(ahr999 float64) string {
	var emoji, status, distance, comment string

	if ahr999 < 0.45 {
		// 抄底区
		emoji = "🟢"
		status = "抄底区"
		// 距离定投区的百分比
		pct := (0.45 - ahr999) / 0.45 * 100
		distance = fmt.Sprintf("已进入抄底，距定投区 %.0f%% 📈", pct)
		comment = "绝佳机会，重仓买入"
	} else if ahr999 < 1.20 {
		// 定投区
		emoji = "🟢"
		status = "定投区"
		// 距离抄底区的百分比
		pct := (ahr999 - 0.45) / ahr999 * 100
		distance = fmt.Sprintf("距 [抄底区 0.45] 还有 %.0f%% 📉", pct)
		comment = "价格划算，坚持定投"
	} else if ahr999 < 5.00 {
		// 持有区
		emoji = "🟡"
		status = "持有区"
		pct := (ahr999 - 1.20) / ahr999 * 100
		distance = fmt.Sprintf("距 [定投区 1.20] 已涨 %.0f%% 📈", pct)
		comment = "暂停买入，持币待涨"
	} else {
		// 逃顶区
		emoji = "🔴"
		status = "逃顶区"
		pct := (ahr999 - 5.00) / ahr999 * 100
		distance = fmt.Sprintf("已超逃顶线 %.0f%% 🚨", pct)
		comment = "分批卖出，锁定利润"
	}

	return fmt.Sprintf(`**1. 囤币指标 (AHR999)**
• 数值: `+"`%.2f`"+` %s
• 状态: **%s**
• 距离: %s
_(点评: %s)_`, ahr999, emoji, status, distance, comment)
}

// formatMVRVSection 格式化MVRV-Z部分
func formatMVRVSection(zScore float64) string {
	var emoji, status, distance string

	if zScore < 0 {
		emoji = "🟢"
		status = "极度低估"
		distance = "已跌破 0 轴，历史大底区域"
	} else if zScore < 1 {
		emoji = "❄️"
		status = "底部区间"
		pct := zScore / 1 * 100
		distance = fmt.Sprintf("距 0 轴还有 %.0f%%，接近大底", 100-pct)
	} else if zScore < 3 {
		emoji = "🟡"
		status = "中性区间"
		distance = "市场温和，可正常操作"
	} else if zScore < 6 {
		emoji = "🟠"
		status = "偏热区间"
		pct := (zScore - 3) / 3 * 100
		distance = fmt.Sprintf("距 [过热 6.0] 还有 %.0f%%", 100-pct)
	} else {
		emoji = "🔴"
		status = "极度过热"
		distance = "市场狂热，谨慎追高"
	}

	return fmt.Sprintf(`**2. 市场冷热 (MVRV-Z)**
• 数值: `+"`%.2f`"+` %s
• 状态: **%s**
• 距离: %s`, zScore, emoji, status, distance)
}

// formatETHSection 格式化ETH部分
func formatETHSection(state model.EthRegressionState) string {
	var emoji, status, strategy string

	switch state {
	case model.EthRegLower:
		emoji = "🟢"
		status = "低估区"
		strategy = "可加大 ETH 配置比例"
	case model.EthRegMiddle:
		emoji = "🟡"
		status = "中性"
		strategy = "不主动出击，跟随 BTC 配比"
	case model.EthRegUpper:
		emoji = "🔴"
		status = "高估区"
		strategy = "减少 ETH，换成 BTC 或 U"
	default:
		emoji = "⚪️"
		status = "未知"
		strategy = "数据不足，保持观望"
	}

	return fmt.Sprintf(`**3. 以太坊 (ETH)**
• 状态: %s **%s**
• 策略: %s`, emoji, status, strategy)
}

// formatSafetySection 格式化安全检查部分
func formatSafetySection(indicators *model.MarketIndicators) string {
	// 杠杆状态
	leverageStatus := "✅"
	if indicators.AccountLeverage > 1.5 {
		leverageStatus = "❌ 危险"
	} else if indicators.AccountLeverage > 1.2 {
		leverageStatus = "⚠️ 警戒"
	}

	// 逃顶信号
	escapeStatus := "⚪️ 暂无风险"
	if indicators.PiCycleCross {
		escapeStatus = "🔴 Pi周期死叉"
	} else if indicators.MaMultiplierState == model.MaStateBullTop {
		escapeStatus = "🔴 突破两年红线"
	}

	return fmt.Sprintf(`**4. 安全检查**
• 杠杆: %.1fx %s (安全线 < 1.5x)
• 逃顶: %s`, indicators.AccountLeverage, leverageStatus, escapeStatus)
}

// formatActionSection 格式化执行建议部分
func formatActionSection(signal *model.TradeSignal) string {
	var action string

	switch signal.ActionBTC {
	case model.ActionHalt:
		action = "⛔️ 停止操作"
	case model.ActionSellAlert:
		action = "🚨 准备离场"
	case model.ActionStrongBuy:
		action = "💪 重仓买入 BTC"
	case model.ActionDCABuy:
		action = "📈 买入 BTC"
	case model.ActionHold, model.ActionHoldCaution:
		action = "✋ 持有等待"
	case model.ActionSell:
		action = "📉 分批卖出"
	default:
		action = "观望"
	}

	factorEmoji := "💰"
	if signal.AmountFactor >= 1.5 {
		factorEmoji = "💰💰"
	} else if signal.AmountFactor == 0 {
		factorEmoji = "🚫"
	}

	return fmt.Sprintf(`🚀 **本周执行: %s**
%s **资金系数: %.1f 倍**`, action, factorEmoji, signal.AmountFactor)
}

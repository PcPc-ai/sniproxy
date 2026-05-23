package whitelist

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Rule 白名单规则
type Rule struct {
	Domain string // 域名或域名后缀
	Type   string // "exact" 精确匹配, "suffix" 后缀匹配
}

// Whitelist 白名单管理器
type Whitelist struct {
	rules       []Rule
	enabled     bool
	defaultMode string // "allow" 或 "deny"
}

// NewWhitelist 创建新的白名单管理器
func NewWhitelist(enabled bool, defaultMode string) *Whitelist {
	return &Whitelist{
		rules:       make([]Rule, 0),
		enabled:     enabled,
		defaultMode: defaultMode,
	}
}

// LoadFromFile 从文件加载白名单规则
func (w *Whitelist) LoadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("打开白名单文件失败: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 解析规则: domain,type
		parts := strings.Split(line, ",")
		if len(parts) != 2 {
			return fmt.Errorf("第%d行格式错误: %s", lineNum, line)
		}

		domain := strings.TrimSpace(parts[0])
		ruleType := strings.TrimSpace(parts[1])

		// 移除域名末尾的点
		domain = strings.TrimSuffix(domain, ".")

		// 验证规则类型
		if ruleType != "exact" && ruleType != "suffix" {
			return fmt.Errorf("第%d行规则类型错误: %s (应为 exact 或 suffix)", lineNum, ruleType)
		}

		w.AddRule(domain, ruleType)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	return nil
}

// AddRule 添加白名单规则
func (w *Whitelist) AddRule(domain, ruleType string) {
	// 规范化域名
	domain = strings.ToLower(strings.TrimSpace(domain))
	domain = strings.TrimSuffix(domain, ".")

	rule := Rule{
		Domain: domain,
		Type:   ruleType,
	}
	w.rules = append(w.rules, rule)
}

// AddRules 批量添加白名单规则
func (w *Whitelist) AddRules(domains []string, ruleType string) {
	for _, domain := range domains {
		w.AddRule(domain, ruleType)
	}
}

// IsAllowed 检查域名是否被允许
func (w *Whitelist) IsAllowed(domain string) bool {
	if !w.enabled {
		return true
	}

	// 规范化域名
	domain = strings.ToLower(strings.TrimSpace(domain))
	domain = strings.TrimSuffix(domain, ".")

	// 检查所有规则
	for _, rule := range w.rules {
		if w.matchRule(domain, rule) {
			return true
		}
	}

	// 如果没有匹配的规则，根据默认模式决定
	return w.defaultMode == "allow"
}

// matchRule 检查域名是否匹配规则
func (w *Whitelist) matchRule(domain string, rule Rule) bool {
	switch rule.Type {
	case "exact":
		return domain == rule.Domain
	case "suffix":
		// 后缀匹配：要么完全相等，要么以 .domain 结尾
		if domain == rule.Domain {
			return true
		}
		return strings.HasSuffix(domain, "."+rule.Domain)
	default:
		return false
	}
}

// GetRulesCount 获取规则数量
func (w *Whitelist) GetRulesCount() int {
	return len(w.rules)
}

// IsEnabled 检查白名单是否启用
func (w *Whitelist) IsEnabled() bool {
	return w.enabled
}

// SetEnabled 设置白名单启用状态
func (w *Whitelist) SetEnabled(enabled bool) {
	w.enabled = enabled
}

// GetDefaultMode 获取默认模式
func (w *Whitelist) GetDefaultMode() string {
	return w.defaultMode
}

// SetDefaultMode 设置默认模式
func (w *Whitelist) SetDefaultMode(mode string) {
	if mode == "allow" || mode == "deny" {
		w.defaultMode = mode
	}
}

// ListRules 列出所有规则（用于调试）
func (w *Whitelist) ListRules() []Rule {
	rules := make([]Rule, len(w.rules))
	copy(rules, w.rules)
	return rules
}

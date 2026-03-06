package msgfmt

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
)

// OptionsRegex 匹配 <options>...</options> 块
var OptionsRegex = regexp.MustCompile(`(?s)<options>(.*?)</options>`)

// OptionRegex 匹配单个 <option> 标签
var OptionRegex = regexp.MustCompile(`<option\s+label="([^"]*)"(?:\s+description="([^"]*)")?\s*/?>`)

// OptionItem 表示一个选项
type OptionItem struct {
	Label        string
	Description string
}

// ParseOptions 从消息内容中解析选项
// 返回: 清理后的消息内容, 选项列表, 问题ID
func ParseOptions(content string) (string, []OptionItem, string) {
	matches := OptionsRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return content, nil, ""
	}

	var options []OptionItem
	for _, match := range matches {
		optionsBlock := match[1]
		optionMatches := OptionRegex.FindAllStringSubmatch(optionsBlock, -1)
		for _, om := range optionMatches {
			option := OptionItem{
				Label:        om[1],
				Description: om[2],
			}
			options = append(options, option)
		}
	}

	// 移除选项块后的内容
	cleanContent := OptionsRegex.ReplaceAllString(content, "")
	cleanContent = strings.TrimSpace(cleanContent)

	// 生成唯一的问题ID
	questionId := generateUUID()

	return cleanContent, options, questionId
}

// FormatAnswers 将用户选择的答案格式化为文本
func FormatAnswers(selectedIndices []int, options []OptionItem) string {
	if len(selectedIndices) == 0 || len(options) == 0 {
		return ""
	}

	var selectedLabels []string
	for _, idx := range selectedIndices {
		if idx >= 0 && idx < len(options) {
			selectedLabels = append(selectedLabels, options[idx].Label)
		}
	}

	return strings.Join(selectedLabels, ", ")
}

// generateUUID 生成一个简单的 UUID 字符串
func generateUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// 如果随机数生成失败, 使用时间戳作为备选
		return "fallback-" + hex.EncodeToString([]byte("00000000"))
	}
	return hex.EncodeToString(b)
}

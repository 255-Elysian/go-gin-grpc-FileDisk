package utils

import (
	"path/filepath"
	"regexp"
)

// Clean 保留原始扩展名 + 基础文件名，去掉特殊字符
func Clean(filename string) string {
	// 1. 获取基础名（防止路径穿越）
	base := filepath.Base(filename)

	// 2. 正则清理：只保留数字、字母、下划线、点、破折号
	reg := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	safe := reg.ReplaceAllString(base, "_")

	// 3. 防止太长（可选）
	if len(safe) > 128 {
		safe = safe[:128]
	}

	return safe
}

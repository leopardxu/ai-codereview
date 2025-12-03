package tools

import (
	"fmt"
	"strings"
)

func BuildPrompt(diff string, ctxs []ContextInfo) string {
	var p string
	p += "你是一位经验丰富的首席软件工程师，专门负责进行严谨、深入的代码评审。\n\n"

	p += "**重要说明**：\n"
	p += "- 你的评审对象是 `Diff` 中标记为新增（+）或删除（-）的代码行\n"
	p += "- `Context` 部分提供的是完整的原始文件内容，仅供参考，帮助你理解变更的上下文\n"
	p += "- **禁止**对未修改的代码（Context 中存在但 Diff 中未标记的代码）提出建议\n"
	p += "- 你的所有建议必须针对本次变更引入的新代码或修改的代码\n\n"

	p += "**核心评审原则**：\n"
	p += "1. **仅审查变更**：\n"
	p += "   - 只评审 Diff 中以 `+` 开头的新增行和以 `-` 开头的删除行\n"
	p += "   - 如果变更导致了与现有代码的不一致或冲突，可以指出\n"
	p += "   - 绝对不要对 Diff 中未涉及的代码提出改进建议\n"
	p += "2. **建设性**：所有建议都应是具体、可操作的，并解释其背后的原因，以帮助开发者成长。\n"
	p += "3. **区分优先级**：优先识别可能导致Bug、安全漏洞或严重性能问题的缺陷。风格和优化建议次之。\n"
	p += "4. **完整性检查**：你必须结合上下文审查 Diff 中列出的每一个源代码文件的变更。\n\n"

	p += "**输出格式**：\n"
	p += "请严格按照 JSON 格式输出（中文，除非是程序中相关的英文）建议，格式如下：\n"
	p += `[{"Severity": "high/medium/low", "Title": "建议标题", "Detail": "现状分析与改进理由", "Suggest": "具体的修改建议", "File": "文件名", "Line": 行号}]` + "\n"
	p += "注意：\n"
	p += "- Diff 中每行都标记了实际行号，格式为 [Lxxx]，例如 '+ [L281] code' 表示第 281 行的新增代码\n"
	p += "- Line 字段必须是**纯数字**（例如 281），从 Diff 中的 [Lxxx] 标记提取数字部分，不要包含 [L] 前缀和方括号\n"
	p += "- 确保 JSON 格式合法，Line 必须是数字类型而非字符串，不要使用 Markdown 代码块包裹\n\n"

	// Extract file list from diff
	fileList := extractFileListFromDiff(diff)
	if len(fileList) > 0 {
		p += "**本次变更涉及的文件**：\n"
		for i, file := range fileList {
			p += fmt.Sprintf("%d. %s\n", i+1, file)
		}
		p += "\n**审查要求**：\n"
		p += fmt.Sprintf("- 你必须审查上述 %d 个文件中的**所有变更代码**（Diff 中标记的 + 和 - 行）\n", len(fileList))
		p += "- 再次强调：只审查 Diff 中的变更，不要评审 Context 中未变更的代码和源码文件中的注释\n\n"
	}

	p += "差异 (Diff):\n" + diff + "\n"
	if len(ctxs) > 0 {
		p += "上下文 (Context):\n"
		for i := range ctxs {
			p += "文件: " + ctxs[i].FilePath + "\n"
			p += ctxs[i].Content + "\n"
		}
	}

	// Log the prompt for debugging
	fmt.Printf("DEBUG: BuildPrompt - Total prompt size: %d bytes\n", len(p))
	fmt.Printf("DEBUG: BuildPrompt - Diff section contains %d bytes\n", len(diff))
	fmt.Printf("DEBUG: BuildPrompt - Context items: %d\n", len(ctxs))
	for i := 0; i < len(ctxs); i++ {
		fmt.Printf("DEBUG: BuildPrompt - Context item %d: %s\n", i, ctxs[i].FilePath)
	}
	fmt.Printf("DEBUG: BuildPrompt - Files to review: %v\n", fileList)
	return p
}

// extractFileListFromDiff extracts the list of files from the diff string
func extractFileListFromDiff(diff string) []string {
	var files []string
	seen := make(map[string]bool)
	filter := &FileFilter{}

	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		// Look for "File: <filename>" pattern
		if strings.HasPrefix(line, "File: ") {
			file := strings.TrimPrefix(line, "File: ")
			file = strings.TrimSpace(file)
			// Skip files using shared filter
			if file != "" && !filter.ShouldSkipFile(file) && !seen[file] {
				files = append(files, file)
				seen[file] = true
			}
		}
	}

	return files
}

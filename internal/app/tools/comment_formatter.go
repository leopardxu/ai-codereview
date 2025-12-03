package tools

func FormatForGerrit(advs []map[string]interface{}) map[string]interface{} {
	comments := make(map[string][]map[string]interface{})
	for _, a := range advs {
		path, _ := a["file"].(string)
		if path == "" {
			continue
		}
		c := map[string]interface{}{
			"line":    a["line"],
			"message": a["message"],
		}
		comments[path] = append(comments[path], c)
	}
	msg := "生成" + itoa(len(advs)) + "条建议"
	return map[string]interface{}{"message": msg, "comments": comments}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

package tools

import (
	"eino-gerrit-review/internal/app/policies"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type GerritTool struct{}

func (t *GerritTool) client() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}

func (t *GerritTool) authHeader() string {
	u := os.Getenv("GERRIT_USER")
	p := os.Getenv("GERRIT_TOKEN")
	if u == "" || p == "" {
		return ""
	}
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(u+":"+p))
}

func (t *GerritTool) base() string { return strings.TrimRight(os.Getenv("GERRIT_BASE_URL"), "/") }

func stripXSSI(b []byte) []byte {
	s := string(b)
	if strings.HasPrefix(s, ")]}\n") || strings.HasPrefix(s, ")]}'\n") {
		return []byte(s[5:])
	}
	return b
}

func (t *GerritTool) GetOpenChanges(project, branch string, limit int) ([]map[string]interface{}, error) {
	if t.base() == "" {
		return []map[string]interface{}{
			{"id": "C123", "project": "linux", "branch": "main", "subject": "fix spinlock sleep"},
			{"id": "C456", "project": "android", "branch": "develop", "subject": "main thread sleep"},
		}, nil
	}
	q := url.QueryEscape("status:open+project:" + project + "+branch:" + branch)
	u := t.base() + "/a/changes/?q=" + q + "&n=" + fmt.Sprintf("%d", limit)
	req, _ := http.NewRequest("GET", u, nil)
	h := t.authHeader()
	if h != "" {
		req.Header.Set("Authorization", h)
	}
	resp, err := t.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, errors.New("gerrit changes error")
	}
	body, _ := io.ReadAll(resp.Body)
	body = stripXSSI(body)
	var arr []map[string]interface{}
	if err := json.Unmarshal(body, &arr); err != nil {
		return nil, err
	}
	return arr, nil
}

func (t *GerritTool) GetDiffs(changeNum, patchset string) ([]map[string]interface{}, error) {
	if t.base() == "" {
		return []map[string]interface{}{
			{"path": "kernel/lock.c", "lang": "c", "patch": "spin_lock(&lock);\nmsleep(20);\nspin_unlock(&lock);"},
			{"path": "app/src/main/java/com/example/MainActivity.java", "lang": "java", "patch": "public void onCreate(){\ntry{Thread.sleep(1000);}catch(Exception e){}\n}"},
		}, nil
	}
	filesURL := t.base() + "/a/changes/" + changeNum + "/revisions/" + patchset + "/files/"
	req, _ := http.NewRequest("GET", filesURL, nil)
	h := t.authHeader()
	if h != "" {
		req.Header.Set("Authorization", h)
	}
	resp, err := t.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, errors.New("gerrit files error")
	}
	body, _ := io.ReadAll(resp.Body)
	body = stripXSSI(body)
	var files map[string]map[string]interface{}
	if err := json.Unmarshal(body, &files); err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(files))
	filter := &FileFilter{}

	for p := range files {
		// Skip files that should not be reviewed
		if filter.ShouldSkipFile(p) {
			continue
		}

		du := t.base() + "/a/changes/" + changeNum + "/revisions/" + patchset + "/files/" + url.PathEscape(p) + "/diff"
		rq, _ := http.NewRequest("GET", du, nil)
		if h != "" {
			rq.Header.Set("Authorization", h)
		}
		rs, er := t.do(rq)
		if er != nil {
			continue
		}
		bb, _ := io.ReadAll(rs.Body)
		rs.Body.Close()
		bb = stripXSSI(bb)
		var diff map[string]interface{}
		if json.Unmarshal(bb, &diff) == nil {
			patch := ""
			// Track line numbers: Gerrit diff has skip, ab, a, b fields
			// We need to track the line number in the new file (side B)
			lineB := 0 // Line number in the new file (after changes)

			if v, ok := diff["content"].([]interface{}); ok {
				for _, seg := range v {
					if m, is := seg.(map[string]interface{}); is {
						// Handle skip: lines that are omitted from context
						if skip, ok := m["skip"].(float64); ok {
							lineB += int(skip)
							continue
						}

						// Gerrit diff format:
						// "a": lines in A (deleted) - don't increment lineB
						// "b": lines in B (added) - increment lineB
						// "ab": lines in both (common) - increment lineB

						if a, ok := m["a"].([]interface{}); ok {
							// Deleted lines - show them but don't increment lineB
							for _, line := range a {
								if s, ok := line.(string); ok {
									patch += "- " + s + "\n"
								}
							}
						}
						if b, ok := m["b"].([]interface{}); ok {
							// Added lines - show with line numbers
							for _, line := range b {
								if s, ok := line.(string); ok {
									lineB++
									patch += fmt.Sprintf("+ [L%d] %s\n", lineB, s)
								}
							}
						}
						if ab, ok := m["ab"].([]interface{}); ok {
							// Context lines - show with line numbers
							for _, line := range ab {
								if s, ok := line.(string); ok {
									lineB++
									patch += fmt.Sprintf("  [L%d] %s\n", lineB, s)
								}
							}
						}
					}
				}
			}
			out = append(out, map[string]interface{}{"path": p, "lang": detectLang(p), "patch": patch})
		}
	}
	return out, nil
}

func (t *GerritTool) GetFileContent(changeNum, revision, file string) (string, error) {
	if t.base() == "" {
		if file == "kernel/lock.c" {
			return "#include <linux/sched.h>\nvoid f(){spin_lock(&lock); msleep(20); spin_unlock(&lock);} ", nil
		}
		if file == "app/src/main/java/com/example/MainActivity.java" {
			return "class MainActivity{ void onCreate(){ try{ Thread.sleep(1000);}catch(Exception e){} } }", nil
		}
		return "", nil
	}
	u := t.base() + "/a/changes/" + changeNum + "/revisions/" + revision + "/files/" + url.PathEscape(file) + "/content"
	req, _ := http.NewRequest("GET", u, nil)
	h := t.authHeader()
	if h != "" {
		req.Header.Set("Authorization", h)
	}
	resp, err := t.do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", errors.New("gerrit content error")
	}
	body, _ := io.ReadAll(resp.Body)
	body = stripXSSI(body)
	dec, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		return string(body), nil
	}
	return string(dec), nil
}

// GetFileContentFromParent retrieves the file content from the parent (base) revision
// This is used to get the original file before changes, not the modified version
func (t *GerritTool) GetFileContentFromParent(changeNum, revision, file string) (string, error) {
	if t.base() == "" {
		// Mock data for testing - return original content before changes
		if file == "kernel/lock.c" {
			return "#include <linux/sched.h>\nvoid f(){spin_lock(&lock); msleep(20); spin_unlock(&lock);} ", nil
		}
		if file == "app/src/main/java/com/example/MainActivity.java" {
			return "class MainActivity{ void onCreate(){ try{ Thread.sleep(1000);}catch(Exception e){} } }", nil
		}
		return "", nil
	}
	// Use ?parent=1 query parameter to get the file content from the parent revision (base)
	u := t.base() + "/a/changes/" + changeNum + "/revisions/" + revision + "/files/" + url.PathEscape(file) + "/content?parent=1"
	req, _ := http.NewRequest("GET", u, nil)
	h := t.authHeader()
	if h != "" {
		req.Header.Set("Authorization", h)
	}
	resp, err := t.do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", errors.New("gerrit content error")
	}
	body, _ := io.ReadAll(resp.Body)
	body = stripXSSI(body)
	dec, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		return string(body), nil
	}
	return string(dec), nil
}

func (t *GerritTool) PostReview(changeNum, revision string, payload map[string]interface{}) (*http.Response, error) {
	if t.base() == "" {
		return nil, nil
	}
	u := t.base() + "/a/changes/" + changeNum + "/revisions/" + revision + "/review"
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", u, io.NopCloser(strings.NewReader(string(b))))
	req.Header.Set("Content-Type", "application/json")
	h := t.authHeader()
	if h != "" {
		req.Header.Set("Authorization", h)
	}
	resp, err := t.do(req)
	if err != nil {
		return nil, err
	}
	// Read and log response body for debugging
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	// Re-create body for caller
	resp.Body = io.NopCloser(strings.NewReader(string(body)))

	if resp.StatusCode >= 400 {
		return resp, fmt.Errorf("gerrit post review failed: %s, body: %s", resp.Status, string(body))
	}
	return resp, nil
}

func detectLang(p string) string {
	if strings.HasSuffix(p, ".c") || strings.HasSuffix(p, ".h") {
		return "c"
	}
	if strings.HasSuffix(p, ".cpp") || strings.HasSuffix(p, ".hpp") {
		return "cpp"
	}
	if strings.HasSuffix(p, ".java") {
		return "java"
	}
	if strings.HasSuffix(p, ".kt") {
		return "kotlin"
	}
	return "text"
}

var gerritLimiter = policies.NewRateLimiter(5)

func (t *GerritTool) do(req *http.Request) (*http.Response, error) {
	var last error
	for i := 0; i < 3; i++ {
		gerritLimiter.Acquire()
		resp, err := t.client().Do(req)
		if err == nil && resp != nil && resp.StatusCode < 500 {
			return resp, nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		last = err
		time.Sleep(time.Duration(200*(i+1)) * time.Millisecond)
	}
	if last != nil {
		return nil, last
	}
	return nil, errors.New("request failed")
}

//

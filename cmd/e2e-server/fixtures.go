package main

// multiFileDiff is a realistic multi-file diff fixture for E2E browser tests.
// It includes modified, added, and deleted files to exercise all UI states.
const multiFileDiff = `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,7 @@
 package main

-func main() {}
+import "fmt"
+
+func main() {
+	fmt.Println("hello world")
+}
diff --git a/utils.go b/utils.go
new file mode 100644
index 0000000..1234567
--- /dev/null
+++ b/utils.go
@@ -0,0 +1,5 @@
+package main
+
+func helper() string {
+	return "helper"
+}
diff --git a/old.go b/old.go
deleted file mode 100644
index 1234567..0000000
--- a/old.go
+++ /dev/null
@@ -1,3 +0,0 @@
-package main
-
-func deprecated() {}
`

// mockChatResponse is a NDJSON Claude stream for chat messages.
// assistant event must use "message":{"content":[...]} to match StreamEvent struct.
const mockChatResponse = `{"type":"system","subtype":"init","session_id":"e2e-mock-session"}
{"type":"assistant","message":{"content":[{"type":"text","text":"This is a mock response from Claude for E2E testing."}]}}
{"type":"result","subtype":"success","result":"This is a mock response from Claude for E2E testing.","session_id":"e2e-mock-session","stop_reason":"end_turn"}
`

// mockReviewResponse is a NDJSON Claude stream for review requests.
const mockReviewResponse = `{"type":"system","subtype":"init","session_id":"e2e-review-session"}
{"type":"assistant","message":{"content":[{"type":"text","text":"Code review: The changes look good overall. Consider adding error handling in main()."}]}}
{"type":"result","subtype":"success","result":"Code review: The changes look good overall. Consider adding error handling in main().","session_id":"e2e-review-session","stop_reason":"end_turn"}
`

package sandbox

import (
	"OJ-API/utils"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// === 結構定義 ===

// Generate System Message when error occur
// Format reference by gTest JSON file
func NewErrorResult(errType JudgeResult, errName string, errMsg string) string {
	now := time.Now().UTC().Format(time.RFC3339)

	failure := Failure{
		Failure: errMsg,
		Type:    "",
	}

	testCase := TestCase{
		Name:      errName,
		File:      "System Message",
		Line:      0,
		Status:    "failed",
		Result:    "ERROR",
		Timestamp: now,
		Time:      "0s",
		Classname: "System",
		Failures:  []Failure{failure},
	}

	testSuite := TestSuite{
		Name:      string(errType),
		MaxScore:  100,
		GetScore:  0,
		Tests:     1,
		Failures:  1,
		Disabled:  0,
		Errors:    0,
		Timestamp: now,
		Time:      "0s",
		TestSuite: []TestCase{testCase},
	}

	all := AllTests{
		Tests:      1,
		Failures:   1,
		Disabled:   0,
		Errors:     0,
		Timestamp:  now,
		Time:       "0s",
		Name:       "AllTests",
		TestSuites: []TestSuite{testSuite},
	}

	jsonBytes, _ := json.MarshalIndent(all, "", "  ")
	return string(jsonBytes)
}

// --- 生成錯誤測資 ---
func generateErrorSuite(target, errName, errMsg string) TestSuite {
	now := time.Now().UTC().Format(time.RFC3339)
	tc := TestCase{
		Name:      errName,
		File:      "System Message",
		Line:      0,
		Status:    "failed",
		Result:    "ERROR",
		Timestamp: now,
		Time:      "0s",
		Classname: "System",
		Failures:  []Failure{{Failure: errMsg, Type: ""}},
	}

	return TestSuite{
		Name:      target,
		MaxScore:  0,
		GetScore:  0,
		Tests:     1,
		Failures:  1,
		Timestamp: now,
		Time:      "0s",
		TestSuite: []TestCase{tc},
	}
}

// --- 讀取 meta 結果 ---
func readMetaResult(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	status := strings.TrimSpace(string(data))
	return status, nil
}

// --- 主流程：整合 message.txt + score.txt + failed result ---
func MergeJudgeResults(baseDir string, finalResults []SandboxScoreResult) (AllTests, float64, error) {
	var all AllTests
	messagePath := filepath.Join(baseDir, "message.txt")
	scorePath := filepath.Join(baseDir, "score.txt")

	// 讀取 message.txt
	data, err := os.ReadFile(messagePath)
	if err != nil {
		all = AllTests{
			Tests:      0,
			Failures:   0,
			Name:       "AllTests",
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
			TestSuites: []TestSuite{},
		}
	} else {
		_ = json.Unmarshal(data, &all)
	}

	// 嘗試讀取分數（單一浮點數）
	var totalScore float64 = 0
	scoreData, err := os.ReadFile(scorePath)
	if err == nil {
		scoreStr := strings.TrimSpace(string(scoreData))
		if s, parseErr := strconv.ParseFloat(scoreStr, 64); parseErr == nil {
			totalScore = s
		}
	}

	// 已存在 suite 索引（避免重複）
	existing := make(map[string]bool)
	for _, ts := range all.TestSuites {
		existing[ts.Name] = true
	}

	// 只處理失敗結果
	for _, r := range finalResults {

		metaPath := filepath.Join(baseDir, fmt.Sprintf("meta_%s_result.txt", r.Target))
		status, err := readMetaResult(metaPath)

		if err != nil {
			utils.Debugf("[Merge] meta not found for %s", r.Target)
			status = string(RUNTIME_ERROR) // fallback
		}

		utils.Debugf("[Merge] target=%s status=%s", r.Target, status)

		// ✅ AC 跳過
		if string(ACCEPTED) == status {
			continue
		}

		if existing[r.Target] {
			continue
		}

		// 🔥 用 meta 的 status，不是 r.Status
		errSuite := generateErrorSuite(r.Target, status, status)

		all.TestSuites = append(all.TestSuites, errSuite)
		all.Failures++
		all.Tests++
	}

	return all, totalScore, nil
}

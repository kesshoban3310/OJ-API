package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Meta map[string]string

type ResourceConfig struct {
	Memory      int `json:"memory"`
	StackMemory int `json:"stack_memory"`
	Time        int `json:"time"`
	WallTime    int `json:"wall_time"`
	FileSize    int `json:"file_size"`
	Processes   int `json:"processes"`
	OpenFiles   int `json:"open_files"`
}

type JudgeResult string

const (
	ACCEPTED              JudgeResult = "ACCEPTED"
	WRONG_ANSWER          JudgeResult = "WRONG_ANSWER"
	RUNTIME_ERROR         JudgeResult = "RUNTIME_ERROR"
	TIME_LIMIT_EXCEEDED   JudgeResult = "TIME_LIMIT_EXCEEDED"
	MEMORY_LIMIT_EXCEEDED JudgeResult = "MEMORY_LIMIT_EXCEEDED"
)

func loadResourceConfig(path string) (*ResourceConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read resource.json: %w", err)
	}

	var rc ResourceConfig
	if err := json.Unmarshal(data, &rc); err != nil {
		return nil, fmt.Errorf("failed to parse resource.json: %w", err)
	}

	return &rc, nil
}

func parseMeta(path string) (Meta, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	meta := make(Meta)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		meta[parts[0]] = parts[1]
	}

	return meta, nil
}

func getMetaPath(jsonPath string) string {
	filename := filepath.Base(jsonPath) // ut_all.json
	target := strings.TrimSuffix(filename, ".json")

	metaFile := "meta_" + target + ".txt"

	return metaFile
}

func metaJudger(metaInfo Meta, resourceLimit *ResourceConfig) JudgeResult {
	status := metaInfo["status"]
	timeStr := metaInfo["time"]
	wallStr := metaInfo["time-wall"]
	memStr := metaInfo["max-rss"]
	exitStr := metaInfo["exitcode"]
	exitSigStr := metaInfo["exitsig"]

	// parse
	timeVal, _ := strconv.ParseFloat(timeStr, 64)
	wallVal, _ := strconv.ParseFloat(wallStr, 64)
	memVal, _ := strconv.Atoi(memStr)
	exitCode, _ := strconv.Atoi(exitStr)
	exitSig, _ := strconv.Atoi(exitSigStr)

	if exitSig != 0 {
		return RUNTIME_ERROR
	} else if status == "SG" {
		return RUNTIME_ERROR
	} else if status == "TO" {
		return TIME_LIMIT_EXCEEDED
	} else if exitCode == 0 {
		if resourceLimit.Memory > 0 && memVal > resourceLimit.Memory {
			return MEMORY_LIMIT_EXCEEDED
		}
		if resourceLimit.Time > 0 && timeVal > float64(resourceLimit.Time) {
			return TIME_LIMIT_EXCEEDED
		}
		if resourceLimit.WallTime > 0 && wallVal > float64(resourceLimit.WallTime) {
			return TIME_LIMIT_EXCEEDED
		}
	}

	// 🔥 5. AC
	return ACCEPTED
}

func metaMain(targetName string, testName string) (Meta, bool) {
	metaPath := getMetaPath(targetName)
	resourceConfig, err := loadResourceConfig(testName)

	ans, _ := parseMeta(metaPath)
	if err != nil {
		log.Fatalf("resource load error: %v", err)
	}

	jr := metaJudger(ans, resourceConfig)

	err = os.WriteFile(strings.TrimSuffix(metaPath, ".txt")+"_result.txt", []byte(jr), 0644)
	if err != nil {
		log.Fatalf("write file failed: %v", err)
	}

	if jr != ACCEPTED {
		return ans, false
	}

	return ans, true
}

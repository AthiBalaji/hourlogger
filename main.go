package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const configFile = "hourlogger.txt"
const metadataFileName = "metadata.json"

type TaskMeta struct {
	Task     string `json:"task"`
	Type     string `json:"type"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Duration int64  `json:"duration_seconds"`
	File     string `json:"file"`
}

// ---------------- CONFIG ----------------

func getBasePath() string {
	data, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Println("Config not found. Run 'plan setup' first.")
		os.Exit(1)
	}

	path := strings.TrimSpace(string(data))

	// Validate path (write test)
	testFile := filepath.Join(path, ".test")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		fmt.Println("Invalid or non-writable base path:", err)
		os.Exit(1)
	}
	os.Remove(testFile)

	return path
}

func setup() {
	fmt.Print("Enter base path: ")
	reader := bufio.NewReader(os.Stdin)

	path, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	path = strings.TrimSpace(path)

	// Validate immediately
	testFile := filepath.Join(path, ".test")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		fmt.Println("Invalid path or no permission:", err)
		return
	}
	os.Remove(testFile)

	err = os.WriteFile(configFile, []byte(path), 0644)
	if err != nil {
		fmt.Println("Error saving config:", err)
		return
	}

	fmt.Println("Setup complete.")
}

// ---------------- START TASK ----------------

func startTask() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Task name: ")
	task, _ := reader.ReadString('\n')

	fmt.Print("Type (topic): ")
	taskType, _ := reader.ReadString('\n')

	task = strings.TrimSpace(task)
	taskType = strings.TrimSpace(taskType)

	if task == "" {
		fmt.Println("Task name cannot be empty")
		return
	}

	startTime := time.Now()

	fmt.Println("Task started. Type your notes below.")
	fmt.Println("Type ':end' to finish.")

	var notes strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			return
		}

		if strings.TrimSpace(line) == ":end" {
			break
		}

		timestamp := time.Now().Format("02 Jan 2006 15:04:05 MST")
		notes.WriteString(fmt.Sprintf("[%s] %s", timestamp, line))
	}

	endTime := time.Now()

	saveTask(task, taskType, startTime, endTime, notes.String())
}

// ---------------- SAVE TASK ----------------

func saveTask(task, taskType string, startTime, endTime time.Time, notes string) {
	base := getBasePath()

	year := endTime.Format("2006")
	month := endTime.Format("January")

	dir := filepath.Join(base, year, month)

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}

	cleanTask := sanitize(task)

	filename := fmt.Sprintf("%s-%s.txt",
		cleanTask,
		endTime.Format("20060102-150405"),
	)

	fullPath := filepath.Join(dir, filename)

	file, err := os.Create(fullPath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("Task: %s\n", task))
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	file.WriteString(fmt.Sprintf("Type: %s\n", taskType))
	file.WriteString(fmt.Sprintf("Start: %s\n", startTime.Format("02 Jan 2006 15:04:05 MST")))
	file.WriteString(fmt.Sprintf("End: %s\n\n", endTime.Format("02 Jan 2006 15:04:05 MST")))
	file.WriteString("Notes:\n")
	file.WriteString(notes)

	fmt.Println("Saved to:", fullPath)

	duration := endTime.Sub(startTime)

	meta := TaskMeta{
		Task:     task,
		Type:     taskType,
		Start:    startTime.Format(time.RFC3339),
		End:      endTime.Format(time.RFC3339),
		Duration: int64(duration.Seconds()),
		File:     fullPath,
	}

	updateMetadata(base, meta)
}

// ---------------- METADATA ----------------

func updateMetadata(base string, entry TaskMeta) {
	metadataPath := filepath.Join(base, metadataFileName)

	var logs []TaskMeta

	data, err := os.ReadFile(metadataPath)
	if err == nil {
		json.Unmarshal(data, &logs)
	}

	logs = append(logs, entry)

	updated, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		fmt.Println("Error encoding metadata:", err)
		return
	}

	temp := metadataPath + ".tmp"

	err = os.WriteFile(temp, updated, 0644)
	if err != nil {
		fmt.Println("Error writing metadata:", err)
		return
	}

	err = os.Rename(temp, metadataPath)
	if err != nil {
		fmt.Println("Error replacing metadata file:", err)
	}
}

// ---------------- REPORT ----------------

func report(args []string) {
	base := getBasePath()
	metadataPath := filepath.Join(base, metadataFileName)

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		fmt.Println("No metadata found.")
		return
	}

	var logs []TaskMeta
	err = json.Unmarshal(data, &logs)
	if err != nil {
		fmt.Println("Corrupted metadata file.")
		return
	}

	now := time.Now()

	var startDate, endDate time.Time

	if len(args) == 0 {
		endDate = now
		startDate = now.AddDate(0, 0, -7)
	} else if len(args) == 1 {
		endDate, err = time.Parse("2006-01-02", args[0])
		if err != nil {
			fmt.Println("Invalid date format (YYYY-MM-DD)")
			return
		}
		startDate = time.Time{}
	} else {
		startDate, err = time.Parse("2006-01-02", args[0])
		if err != nil {
			fmt.Println("Invalid start date")
			return
		}

		endDate, err = time.Parse("2006-01-02", args[1])
		if err != nil {
			fmt.Println("Invalid end date")
			return
		}
	}

	totalSeconds := int64(0)
	taskCount := make(map[string]int)
	typeTime := make(map[string]int64)

	for _, log := range logs {
		t, err := time.Parse(time.RFC3339, log.End)
		if err != nil {
			continue
		}

		if (t.After(startDate) || t.Equal(startDate)) &&
			(t.Before(endDate) || t.Equal(endDate)) {

			totalSeconds += log.Duration
			taskCount[log.Task]++
			typeTime[log.Type] += log.Duration
		}
	}

	if totalSeconds == 0 {
		fmt.Println("No data found.")
		return
	}

	h := totalSeconds / 3600
	m := (totalSeconds % 3600) / 60

	fmt.Printf("\nTotal Time logged: %dh %dm\n\n", h, m)

	fmt.Println("Tasks by frequency:")
	for k, v := range taskCount {
		fmt.Printf("- %s: %d\n", k, v)
	}

	fmt.Println("\nTime by type of task:")
	for k, v := range typeTime {
		h := v / 3600
		m := (v % 3600) / 60
		fmt.Printf("- %s: %dh %dm\n", k, h, m)
	}
}

// ---------------- UTILS ----------------

func sanitize(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	return s
}

// ---------------- MAIN ----------------

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: plan <setup|start|report>")
		return
	}

	switch os.Args[1] {
	case "setup":
		setup()
	case "start":
		startTask()
	case "report":
		report(os.Args[2:])
	default:
		fmt.Println("Unknown command")
	}
}
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	LFilePath "path/filepath"
	"strings"
)

type LogEntry struct {
	Type            string `json:"Type"`
	Function        string `json:"Function"`
	ContextID       uint64 `json:"ContextID"`
	ContextParentID uint64 `json:"ContextParentID"`
}

type CallGraph interface {
	Constructor()
	StartFunction(int, LogEntry)
	EndFunction(LogEntry)
	PrintGraph()
}

type CallGraphText struct {
	graphs map[uint64][]string
}

func (cg *CallGraphText) Constructor() {
	cg.graphs = make(map[uint64][]string)
}

func (cg *CallGraphText) StartFunction(stackSize int, entry LogEntry) {
	graph := fmt.Sprintf("%s- %s", strings.Repeat("  ", stackSize), entry.Function)
	cg.graphs[entry.ContextID] = append(cg.graphs[entry.ContextID], graph)
}

func (cg *CallGraphText) EndFunction(entry LogEntry) {
	// cg.graphs[entry.ContextID] = entry
}

func (cg *CallGraphText) PrintGraph() {
	for graph := range cg.graphs {
		for _, line := range cg.graphs[graph] {
			fmt.Println(line)
		}
	}
}

// Load the log file
func load_log(filePath string) ([]LogEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []LogEntry

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if idx := strings.Index(line, "CALL_GRAPH:"); idx != -1 {
			jsonPart := strings.TrimSpace(line[idx+len("CALL_GRAPH:"):])

			var entry LogEntry
			if err := json.Unmarshal([]byte(jsonPart), &entry); err != nil {
				fmt.Printf("Error parsing JSON on line: %s\nError: %v\n", line, err)
				continue
			}

			entries = append(entries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

func main() {
	inFilePath := os.Args[1]

	// Get the absolute path of the input file
	filePath, err := LFilePath.Abs(inFilePath)
	if err != nil {
		fmt.Println("Error absolute path:", err)
		return
	}

	// Load the log file
	entries, err := load_log(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Create a map of stacks and graphs
	stacks := make(map[uint64][]string)
	// graphs := make(map[uint64][]string)
	graphs := CallGraphText{}
	graphs.Constructor()

	for _, entry := range entries {
		if entry.Type == "S" {
			stacks[entry.ContextID] = append(stacks[entry.ContextID], entry.Function)
			stack := stacks[entry.ContextID]
			graphs.StartFunction(len(stack)-1, entry)
		} else if entry.Type == "E" {
			stack := stacks[entry.ContextID]
			if len(stack) > 0 {
				stacks[entry.ContextID] = stack[:len(stack)-1]
			}
		}
	}

	// Print the graphs
	graphs.PrintGraph()
	// for graph := range graphs {
	// 	for _, line := range graphs[graph] {
	// 		fmt.Println(line)
	// 	}
	// }
}

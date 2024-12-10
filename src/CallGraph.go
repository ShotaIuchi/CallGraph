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
	Action          string `json:"Action"`
	ContextID       string `json:"ContextID"`
	ContextParentID string `json:"ContextParentID"`
	Message         string `json:"Message"`
	Timestamp       uint64 `json:"Timestamp"`
}

type CallGraph interface {
	Constructor()
	StartAction(int, LogEntry, LogEntry)
	EndAction(LogEntry, LogEntry)
	IFAction(LogEntry, LogEntry)
	PrintGraph()
}

type CallGraphText struct {
	graphs map[string][]string
}

func (cg *CallGraphText) Constructor() {
	cg.graphs = make(map[string][]string)
}

func (cg *CallGraphText) StartAction(depth int, pEntry LogEntry, cEntry LogEntry) {
	var time uint64
	if pEntry.Action != "" {
		time = cEntry.Timestamp - pEntry.Timestamp
	}
	graph := fmt.Sprintf("%s- %s: %s (T:%d)", strings.Repeat("  ", depth), cEntry.Action, cEntry.Message, time)
	cg.graphs[cEntry.ContextID] = append(cg.graphs[cEntry.ContextID], graph)
}

func (cg *CallGraphText) EndAction(pEntry LogEntry, cEntry LogEntry) {
}

func (cg *CallGraphText) IFAction(pEntry LogEntry, cEntry LogEntry) {
}

func (cg *CallGraphText) PrintGraph() {
	for graph := range cg.graphs {
		for _, line := range cg.graphs[graph] {
			fmt.Println(line)
		}
	}
}

type CallGraphPlantUML struct {
	graphs map[string][]string
}

func (cg *CallGraphPlantUML) Constructor() {
	cg.graphs = make(map[string][]string)
}

func (cg *CallGraphPlantUML) StartAction(depth int, pEntry LogEntry, cEntry LogEntry) {
	if pEntry.Action != "" {
		time := cEntry.Timestamp - pEntry.Timestamp
		graph := fmt.Sprintf("%s -> %s : %s (T:%d)", pEntry.Action, cEntry.Action, cEntry.Message, time)
		cg.graphs[cEntry.ContextID] = append(cg.graphs[cEntry.ContextID], graph)
	}
}

func (cg *CallGraphPlantUML) EndAction(pEntry LogEntry, cEntry LogEntry) {
	if pEntry.Action != "" {
		time := cEntry.Timestamp - pEntry.Timestamp
		graph := fmt.Sprintf("%s <- %s : %s (T:%d)", pEntry.Action, cEntry.Action, cEntry.Message, time)
		cg.graphs[cEntry.ContextID] = append(cg.graphs[cEntry.ContextID], graph)
	}
}

func (cg *CallGraphPlantUML) IFAction(pEntry LogEntry, cEntry LogEntry) {
	if pEntry.Action != "" {
		time := cEntry.Timestamp - pEntry.Timestamp
		graph := fmt.Sprintf("%s -> %s : %s (T:%d)\\n%s", pEntry.Action, pEntry.Action, cEntry.Action, time, cEntry.Message)
		cg.graphs[cEntry.ContextID] = append(cg.graphs[cEntry.ContextID], graph)
	}
}

func (cg *CallGraphPlantUML) PrintGraph() {
	fmt.Println("@startuml")
	for graph := range cg.graphs {
		for _, line := range cg.graphs[graph] {
			fmt.Println(line)
		}
	}
	fmt.Println("@enduml")
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
	stacks := make(map[string][]LogEntry)
	stacksBef := make(map[string]LogEntry)

	// Create a CallGraph object
	var graphs CallGraph = nil
	// graphs = &CallGraphText{}
	graphs = &CallGraphPlantUML{}
	graphs.Constructor()

	// Iterate over the entries
	for _, cEntry := range entries {
		if cEntry.Type == "ST" {
			var pEntry LogEntry = LogEntry{}
			if len(stacks[cEntry.ContextID]) > 0 {
				pEntry = stacks[cEntry.ContextID][len(stacks[cEntry.ContextID])-1]
			}
			stacks[cEntry.ContextID] = append(stacks[cEntry.ContextID], cEntry)
			stack := stacks[cEntry.ContextID]
			graphs.StartAction(len(stack)-1, pEntry, cEntry)
		} else if cEntry.Type == "ED" {
			stack := stacks[cEntry.ContextID]
			if len(stack) > 0 {
				cEntry := stack[len(stack)-1]
				stacks[cEntry.ContextID] = stack[:len(stack)-1]
				var pEntry LogEntry = LogEntry{}
				if len(stacks[cEntry.ContextID]) > 0 {
					pEntry = stacks[cEntry.ContextID][len(stacks[cEntry.ContextID])-1]
				}
				graphs.EndAction(pEntry, cEntry)
			}
		} else if cEntry.Type == "DO" {
			if stack, ok := stacksBef[cEntry.ContextID]; ok {
				graphs.IFAction(stack, cEntry)
			}
		}
		stacksBef[cEntry.ContextID] = cEntry
	}

	// Print the graphs
	graphs.PrintGraph()
}

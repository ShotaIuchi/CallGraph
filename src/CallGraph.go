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
	StartFunction(int, LogEntry, LogEntry)
	EndFunction(LogEntry, LogEntry)
	PrintGraph()
}

type CallGraphText struct {
	graphs map[uint64][]string
}

func (cg *CallGraphText) Constructor() {
	cg.graphs = make(map[uint64][]string)
}

func (cg *CallGraphText) StartFunction(depth int, pEntry LogEntry, cEntry LogEntry) {
	graph := fmt.Sprintf("%s- %s", strings.Repeat("  ", depth), cEntry.Function)
	cg.graphs[cEntry.ContextID] = append(cg.graphs[cEntry.ContextID], graph)
}

func (cg *CallGraphText) EndFunction(pEntry LogEntry, cEntry LogEntry) {
}

func (cg *CallGraphText) PrintGraph() {
	for graph := range cg.graphs {
		for _, line := range cg.graphs[graph] {
			fmt.Println(line)
		}
	}
}

type CallGraphPlantUML struct {
	graphs map[uint64][]string
}

func (cg *CallGraphPlantUML) Constructor() {
	cg.graphs = make(map[uint64][]string)
}

func (cg *CallGraphPlantUML) StartFunction(depth int, pEntry LogEntry, cEntry LogEntry) {
	if pEntry.Function != "" {
		graph := fmt.Sprintf("%s -> %s", pEntry.Function, cEntry.Function)
		cg.graphs[cEntry.ContextID] = append(cg.graphs[cEntry.ContextID], graph)
	}
}

func (cg *CallGraphPlantUML) EndFunction(pEntry LogEntry, cEntry LogEntry) {
	if pEntry.Function != "" {
		graph := fmt.Sprintf("%s <- %s", pEntry.Function, cEntry.Function)
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
	stacks := make(map[uint64][]LogEntry)

	// Create a CallGraph object
	var graphs CallGraph = nil
	// graphs = &CallGraphText{}
	graphs = &CallGraphPlantUML{}
	graphs.Constructor()

	// Iterate over the entries
	for _, cEntry := range entries {
		if cEntry.Type == "S" {
			var pEntry LogEntry = LogEntry{}
			if len(stacks[cEntry.ContextID]) > 0 {
				pEntry = stacks[cEntry.ContextID][len(stacks[cEntry.ContextID])-1]
			}
			stacks[cEntry.ContextID] = append(stacks[cEntry.ContextID], cEntry)
			stack := stacks[cEntry.ContextID]
			graphs.StartFunction(len(stack)-1, pEntry, cEntry)
		} else if cEntry.Type == "E" {
			stack := stacks[cEntry.ContextID]
			if len(stack) > 0 {
				cEntry := stack[len(stack)-1]
				stacks[cEntry.ContextID] = stack[:len(stack)-1]
				var pEntry LogEntry = LogEntry{}
				if len(stacks[cEntry.ContextID]) > 0 {
					pEntry = stacks[cEntry.ContextID][len(stacks[cEntry.ContextID])-1]
				}
				graphs.EndFunction(pEntry, cEntry)
			}
		}
	}

	// Print the graphs
	graphs.PrintGraph()
}

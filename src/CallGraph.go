package main

import (
	"fmt"
	"strings"
)

type LogEntry struct {
	Type            string
	Function        string
	ContextID       uint64
	ContextParentID uint64
}

func main() {
	entries := []LogEntry{
		{Type: "S", Function: "FuncA", ContextID: 1},
		{Type: "S", Function: "FuncA-1", ContextID: 1},
		{Type: "S", Function: "FuncA-1-1", ContextID: 2},
		{Type: "E", Function: "FuncA-1-1", ContextID: 2},
		{Type: "S", Function: "FuncA-1-2", ContextID: 1},
		{Type: "E", Function: "FuncA-1-2", ContextID: 1},
		{Type: "E", Function: "FuncA-1", ContextID: 1},
		{Type: "S", Function: "FuncA-2", ContextID: 1},
		{Type: "E", Function: "FuncA-2", ContextID: 1},
		{Type: "E", Function: "FuncA", ContextID: 1},
		{Type: "S", Function: "FuncB", ContextID: 1},
		{Type: "E", Function: "FuncB", ContextID: 1},
	}

	stacks := make(map[uint64][]string)
	graphs := make(map[uint64][]string)

	for _, entry := range entries {
		if entry.Type == "S" {
			stacks[entry.ContextID] = append(stacks[entry.ContextID], entry.Function)
			stack := stacks[entry.ContextID]
			graph := fmt.Sprintf("%s- %s", strings.Repeat("  ", len(stack)-1), entry.Function)
			graphs[entry.ContextID] = append(graphs[entry.ContextID], graph)
		} else if entry.Type == "E" {
			stack := stacks[entry.ContextID]
			if len(stack) > 0 {
				stacks[entry.ContextID] = stack[:len(stack)-1]
			}
		}
	}

	for graph := range graphs {
		for _, line := range graphs[graph] {
			fmt.Println(line)
		}
	}
}

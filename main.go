package main

import (
	"fmt"
	"os"

	"github.com/engigu/baihu-panel/cmd"
	"github.com/engigu/baihu-panel/internal/bootstrap"
	"github.com/engigu/baihu-panel/internal/constant"
)

func printHelp() {
	fmt.Println("Usage: baihu <command> [arguments]")
	fmt.Println("Available commands:")
	for _, info := range constant.Commands {
		fmt.Printf("  %-12s %s\n", info.Name, info.Description)
	}
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	commandName := os.Args[1]

	if commandName == "server" {
		bootstrap.New().Run()
		return
	}

	if handler, ok := cmd.Handlers[commandName]; ok {
		handler(os.Args[2:])
		return
	}

	fmt.Printf("Unknown command: %s\n", commandName)
	printHelp()
	os.Exit(1)
}

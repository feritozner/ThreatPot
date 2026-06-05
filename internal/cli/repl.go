package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"threatpot/internal/core"
)

var currentContext = "null"
var cmdHistory []string

func Start() {
	PrintBanner()
	core.InitPots()
	scanner := bufio.NewScanner(os.Stdin)

	for {

		printPrompt(currentContext)

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		cmdHistory = append(cmdHistory, input)
		parts := strings.Fields(input)
		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "help":
			printHelp()

		case "clear":
			clearScreen()

		case "banner":
			PrintBanner()

		case "history":
			PrintHistory(cmdHistory)

		case "create":
			createPot(parts)

		case "use":
			usePot(parts)

		case "back":
			currentContext = "null"

		case "logs":
			printLogs(currentContext, parts)

		case "watch":
			handleWatchCommand(parts)

		case "inspect":
			inspectRequest(currentContext, parts)

		case "route":
			routeCommand(currentContext, parts)

		case "show":
			showCommand(currentContext, parts)

		case "set":
			setOption(currentContext, parts)

		case "start":
			startPot(currentContext)

		case "stop":
			stopPot(currentContext)

		case "export":
			exportLogs(currentContext, parts)

		case "exit":
			fmt.Printf("%s[*] Shutting down ThreatPot...%s\n", ColorYellow, ColorReset)
			os.Exit(0)

		default:
			fmt.Printf("%s[-] Unknown command: %s%s\n", ColorRed, cmd, ColorReset)
		}
	}
}

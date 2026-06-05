package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"threatpot/internal/core"
	"threatpot/internal/webpot"
	"time"
)

func exportLogs(currentContext string, parts []string) {
	if currentContext == "null" {
		fmt.Printf("%s[-] Please select a pot first.%s\n", ColorRed, ColorReset)
		return
	}

	if len(parts) < 2 {
		fmt.Printf("%s[-] Usage: export all OR export <hit_id>%s\n", ColorRed, ColorReset)
		return
	}

	pot := core.Pots[currentContext]
	targetID := parts[1]

	if err := os.MkdirAll("dumps", os.ModePerm); err != nil {
		fmt.Printf("%s[-] Failed to create dumps directory: %v%s\n", ColorRed, err, ColorReset)
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	var filename string
	var contentBuilder strings.Builder

	pot.Mutex.Lock()
	if targetID == "all" {
		if len(pot.ReqDumps) == 0 {
			fmt.Printf("%s[-] No requests to export.%s\n", ColorRed, ColorReset)
			pot.Mutex.Unlock()
			return
		}
		filename = fmt.Sprintf("dumps/%s_all_%s.txt", pot.Name, timestamp)
		contentBuilder.WriteString(fmt.Sprintf("=== THREATPOT EXPORT: %s ===\n", pot.Name))
		contentBuilder.WriteString(fmt.Sprintf("Export Date: %s\n\n", time.Now().Format(time.RFC1123)))

		for _, req := range pot.ReqDumps {
			contentBuilder.WriteString(fmt.Sprintf("---------- HIT #%d ----------\n", req.ID))
			contentBuilder.WriteString(fmt.Sprintf("Time: %s | IP: %s\n\n", req.Time, req.IP))
			contentBuilder.WriteString(req.RawData)
			contentBuilder.WriteString("\n\n")
		}
	} else {
		reqID, err := strconv.Atoi(targetID)
		if err != nil {
			fmt.Printf("%s[-] Invalid HIT ID.%s\n", ColorRed, ColorReset)
			pot.Mutex.Unlock()
			return
		}

		found := false
		for _, req := range pot.ReqDumps {
			if req.ID == reqID {
				filename = fmt.Sprintf("dumps/%s_hit_%d_%s.txt", pot.Name, reqID, timestamp)
				contentBuilder.WriteString(fmt.Sprintf("=== THREATPOT EXPORT: %s (HIT #%d) ===\n", pot.Name, reqID))
				contentBuilder.WriteString(fmt.Sprintf("Time: %s | IP: %s\n\n", req.Time, req.IP))
				contentBuilder.WriteString(req.RawData)
				contentBuilder.WriteString("\n")
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("%s[-] No request found with HIT ID #%d%s\n", ColorRed, reqID, ColorReset)
			pot.Mutex.Unlock()
			return
		}
	}
	pot.Mutex.Unlock()

	err := os.WriteFile(filename, []byte(contentBuilder.String()), 0644)
	if err != nil {
		fmt.Printf("%s[-] Failed to write file: %v%s\n", ColorRed, err, ColorReset)
	} else {
		fmt.Printf("%s[+] Successfully exported to: %s%s\n", ColorGreen, filename, ColorReset)
	}
}

func stopPot(currentContext string) {
	if currentContext == "null" {
		fmt.Printf("%s[-] Please select a pot first.%s\n", ColorRed, ColorReset)
		return
	}
	pot := core.Pots[currentContext]
	if err := webpot.StopServer(pot); err != nil {
		fmt.Printf("%s[-] Failed to stop: %v%s\n", ColorRed, err, ColorReset)
	} else {
		fmt.Printf("[*] %s stopped.\n", pot.Name)
	}
}

func startPot(currentContext string) {
	if currentContext == "null" {
		fmt.Printf("%s[-] Please select a pot first.%s\n", ColorRed, ColorReset)
		return
	}
	pot := core.Pots[currentContext]
	if err := webpot.StartServer(pot); err != nil {
		fmt.Printf("%s[-] Failed to start: %v%s\n", ColorRed, err, ColorReset)
	} else {
		fmt.Printf("%s[+] %s started on %s:%d%s\n", ColorGreen, pot.Name, pot.BindIP, pot.Port, ColorReset)
	}
}

func setOption(currentContext string, parts []string) {
	if currentContext == "null" {
		fmt.Printf("%s[-] Please select a pot first.%s\n", ColorRed, ColorReset)
		return
	}
	if len(parts) < 3 {
		fmt.Printf("%s[-] Usage: set <option> <value>%s\n", ColorRed, ColorReset)
		return
	}
	pot := core.Pots[currentContext]
	option := strings.ToLower(parts[1])
	value := strings.Join(parts[2:], " ")

	if pot.IsRunning {
		fmt.Printf("%s[-] Cannot change settings while pot is running.%s\n", ColorRed, ColorReset)
		return
	}

	switch option {
	case "port":
		portNum, err := strconv.Atoi(value)
		if err != nil {
			fmt.Printf("%s[-] Invalid port.%s\n", ColorRed, ColorReset)
		} else {
			pot.Port = portNum
			fmt.Printf("[+] Port set to => %d\n", pot.Port)
		}
	case "bindip":
		pot.BindIP = value
		fmt.Printf("[+] BindIP set to => %s\n", pot.BindIP)
	case "serverheader":
		pot.ServerHeader = value
		fmt.Printf("[+] ServerHeader set to => %s\n", pot.ServerHeader)
	case "delay":
		delayNum, err := strconv.Atoi(value)
		if err != nil || delayNum < 0 {
			fmt.Printf("%s[-] Invalid delay. Must be a positive number in ms.%s\n", ColorRed, ColorReset)
		} else {
			pot.Delay = delayNum
			fmt.Printf("[+] Delay (Tarpit) set to => %d ms\n", pot.Delay)
		}
	case "description":
		pot.Description = value
		fmt.Printf("[+] Description set to => %s\n", pot.Description)
	default:
		fmt.Printf("%s[-] Unknown option: %s%s\n", ColorRed, option, ColorReset)
	}
}

func showCommand(currentContext string, parts []string) {
	if len(parts) > 1 {
		if parts[1] == "pots" {
			fmt.Printf("\n%sAvailable Pots:%s\n", ColorYellow, ColorReset)
			for name, p := range core.Pots {
				status := "Stopped"
				if p.IsRunning {
					status = "Running"
				}
				fmt.Printf("  - %-10s (Port: %-4d Status: %-7s) => %s\n", name, p.Port, status, p.Description)
			}
			fmt.Println()
		} else if parts[1] == "options" && currentContext != "" {
			p := core.Pots[currentContext]
			fmt.Printf("\n%sOptions for %s:%s\n", ColorYellow, p.Name, ColorReset)
			fmt.Printf("  Description   : %s\n", p.Description)
			fmt.Printf("  Port          : %d\n", p.Port)
			fmt.Printf("  BindIP        : %s\n", p.BindIP)
			fmt.Printf("  ServerHeader  : %s\n", p.ServerHeader)
			fmt.Printf("  Delay (ms)    : %d\n", p.Delay)
			fmt.Printf("  Routes Count  : %d (Type 'route list' to see details)\n\n", len(p.Routes))
		} else if parts[1] == "options" && currentContext == "null" {
			fmt.Printf("%s[-] Please select a pot first.%s\n", ColorRed, ColorReset)
		}
	} else {
		fmt.Printf("%s[-] Usage: show pots OR show options%s\n", ColorRed, ColorReset)
	}
}

func routeCommand(currentContext string, parts []string) {
	if currentContext == "null" {
		fmt.Printf("%s[-] Please select a pot first.%s\n", ColorRed, ColorReset)
		return
	}
	if len(parts) < 2 {
		fmt.Printf("%s[-] Usage: route <list|add|del> ...%s\n", ColorRed, ColorReset)
		return
	}

	pot := core.Pots[currentContext]
	action := strings.ToLower(parts[1])

	switch action {
	case "list":
		fmt.Printf("\n%sRoutes for %s:%s\n", ColorYellow, pot.Name, ColorReset)
		pot.Mutex.Lock()
		if len(pot.Routes) == 0 {
			fmt.Println("  No routes configured.")
		} else {
			for path, file := range pot.Routes {
				fmt.Printf("  %-20s -> %s\n", path, file)
			}
		}
		pot.Mutex.Unlock()
		fmt.Println()

	case "add":
		if len(parts) < 4 {
			fmt.Printf("%s[-] Usage: route add <endpoint> <file_path>%s\n", ColorRed, ColorReset)
			fmt.Printf("    Example: route add /vpn templates/login.html\n")
			return
		}
		endpoint := parts[2]
		filePath := parts[3]

		pot.Mutex.Lock()
		pot.Routes[endpoint] = filePath
		pot.Mutex.Unlock()
		fmt.Printf("%s[+] Route mapped: %s -> %s%s\n", ColorGreen, endpoint, filePath, ColorReset)

	case "del":
		if len(parts) < 3 {
			fmt.Printf("%s[-] Usage: route del <endpoint>%s\n", ColorRed, ColorReset)
			return
		}
		endpoint := parts[2]

		pot.Mutex.Lock()
		if _, exists := pot.Routes[endpoint]; exists {
			delete(pot.Routes, endpoint)
			fmt.Printf("%s[*] Route removed: %s%s\n", ColorYellow, endpoint, ColorReset)
		} else {
			fmt.Printf("%s[-] Route '%s' not found.%s\n", ColorRed, endpoint, ColorReset)
		}
		pot.Mutex.Unlock()

	default:
		fmt.Printf("%s[-] Unknown route action. Use: list, add, del%s\n", ColorRed, ColorReset)
	}
}

func inspectRequest(currentContext string, parts []string) {
	if currentContext == "null" {
		fmt.Printf("%s[-] Please select a pot first.%s\n", ColorRed, ColorReset)
		return
	}
	if len(parts) < 2 {
		fmt.Printf("%s[-] Usage: inspect <hit_id>%s\n", ColorRed, ColorReset)
		return
	}

	reqID, err := strconv.Atoi(parts[1])
	if err != nil {
		fmt.Printf("%s[-] Invalid HIT ID. Must be a number.%s\n", ColorRed, ColorReset)
		return
	}

	pot := core.Pots[currentContext]
	found := false

	pot.Mutex.Lock()
	for _, req := range pot.ReqDumps {
		if req.ID == reqID {
			fmt.Printf("\n%s========== RAW HTTP DUMP [HIT #%d] ==========%s\n", ColorCyan, req.ID, ColorReset)
			fmt.Printf("Time   : %s\n", req.Time)
			fmt.Printf("IP     : %s\n", req.IP)
			fmt.Println(strings.Repeat("-", 45))
			fmt.Printf("%s%s%s\n", ColorGreen, req.RawData, ColorReset)
			fmt.Printf("%s=============================================%s\n\n", ColorCyan, ColorReset)
			found = true
			break
		}
	}
	pot.Mutex.Unlock()

	if !found {
		fmt.Printf("%s[-] No request found with HIT ID #%d%s\n", ColorRed, reqID, ColorReset)
	}
}

func printLogs(potName string, parts []string) {
	if currentContext != "null" {
		printStaticLogs(currentContext)
	} else {
		if len(parts) > 1 {
			if parts[1] == "*" {
				for name := range core.Pots {
					printStaticLogs(name)
				}
			} else {
				printStaticLogs(parts[1])
			}
		} else {
			fmt.Printf("%s[-] Usage: logs <pot_name> OR logs *%s\n", ColorRed, ColorReset)
		}
	}
}

func usePot(parts []string) {
	if len(parts) < 2 {
		fmt.Printf("%s[-] Usage: use <pot_name>%s\n", ColorRed, ColorReset)
		return
	}
	potName := parts[1]
	if _, exists := core.Pots[potName]; exists {
		currentContext = potName
	} else {
		fmt.Printf("%s[-] Pot '%s' not found.%s\n", ColorRed, potName, ColorReset)
	}
}

func createPot(parts []string) {
	if len(parts) < 2 {
		fmt.Printf("%s[-] Usage: create <pot_name>%s\n", ColorRed, ColorReset)
		return
	}
	potName := parts[1]
	if _, exists := core.Pots[potName]; exists {
		fmt.Printf("%s[-] Pot '%s' already exists.%s\n", ColorRed, potName, ColorReset)
	} else {
		core.CreatePot(potName)
		fmt.Printf("%s[+] Pot '%s' created successfully!%s\n", ColorGreen, potName, ColorReset)
	}
}

func PrintHistory(cmdHistory []string) {
	fmt.Printf("\n%sCommand History:%s\n", ColorYellow, ColorReset)
	for i, h := range cmdHistory {
		fmt.Printf("  %d: %s\n", i+1, h)
	}
	fmt.Println()
}

func printPrompt(CurrentContext string) {
	promptStr := ColorGreen + "[" + ColorReset + currentContext + ColorGreen + "] threatpot:~$ " + ColorReset
	fmt.Printf(promptStr)
}

func printStaticLogs(potName string) {
	pot, exists := core.Pots[potName]
	if !exists {
		fmt.Printf("%s[-] Pot '%s' not found.%s\n", ColorRed, potName, ColorReset)
		return
	}

	fmt.Printf("\n%s--- Static Logs for %s ---%s\n", ColorYellow, potName, ColorReset)
	pot.Mutex.Lock()
	if len(pot.Logs) == 0 {
		fmt.Println("  No logs recorded yet.")
	} else {
		for _, l := range pot.Logs {
			fmt.Println(l)
		}
	}
	pot.Mutex.Unlock()
	fmt.Println()
}

func handleWatchCommand(parts []string) {
	var targets []*core.Pot

	if len(parts) == 1 {
		if currentContext != "" {
			targets = append(targets, core.Pots[currentContext])
		} else {
			fmt.Printf("%s[-] Usage: watch <pot_name> OR watch *%s\n", ColorRed, ColorReset)
			return
		}
	} else {
		if parts[1] == "*" {
			for _, p := range core.Pots {
				targets = append(targets, p)
			}
		} else {
			if p, exists := core.Pots[parts[1]]; exists {
				targets = append(targets, p)
			} else {
				fmt.Printf("%s[-] Pot '%s' not found.%s\n", ColorRed, parts[1], ColorReset)
				return
			}
		}
	}

	fmt.Printf("\n%s[*] Live Stream Started. Press ENTER to stop...%s\n\n", ColorGreen, ColorReset)

	for _, p := range targets {
		p.Mutex.Lock()
		for _, l := range p.Logs {
			fmt.Println(l)
		}
		p.Mutex.Unlock()
	}

	aggChan := make(chan string, 100)
	for _, p := range targets {
		ch := p.Subscribe()
		defer p.Unsubscribe(ch)
		go func(c chan string) {
			for l := range c {
				aggChan <- l
			}
		}(ch)
	}

	done := make(chan struct{})
	go func() {
		reader := bufio.NewReader(os.Stdin)
		reader.ReadBytes('\n')
		close(done)
	}()

WatchLoop:
	for {
		select {
		case l := <-aggChan:
			fmt.Println(l)
		case <-done:
			break WatchLoop
		}
	}

	fmt.Printf("\n%s[*] Stopped watching.%s\n", ColorYellow, ColorReset)
}

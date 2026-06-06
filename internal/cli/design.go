package cli

import "fmt"

const (
	ColorGreen   = "\033[32m"
	ColorRed     = "\033[31m"
	ColorYellow  = "\033[33m"
	ColorReset   = "\033[0m"
	ColorCyan    = "\033[36m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
)

func PrintBanner() {
	clearScreen()
	fmt.Printf(ColorGreen + ` 
 _____ _                    _   ____       _   
|_   _| |__  _ __ ___  __ _| |_|  _ \ ___ | |_ 
  | | | '_ \| '__/ _ \/ _' | __| |_) / _ \| __|
  | | | | | | | |  __/ (_| | |_|  __/ (_) | |_ 
  |_| |_| |_|_|  \___|\__,_|\__|_|   \___/ \__|` + "\n\n" + ColorReset)
	fmt.Printf(ColorYellow + " [*] Welcome to ThreatPot!" + ColorReset + " Type 'help' to help menu and type 'exit' to quit.\n")
	fmt.Printf("    ~~~~~~~~~~~~~~~~~~~~~~\n\n")
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func printHelp() {
	fmt.Printf("\n%sCommands:%s\n", ColorYellow, ColorReset)
	fmt.Println("  help          - Show this menu")
	fmt.Println("  banner        - Clear screen and show ThreatPot banner")
	fmt.Println("  clear         - Clear the terminal screen")
	fmt.Println("  create [name] - Create a new honeypot")
	fmt.Println("  show pots     - List all pots")
	fmt.Println("  show options  - Show settings for selected pot")
	fmt.Println("  use [pot]     - Select a pot (e.g. use pot_1)")
	fmt.Println("  set [k] [v]   - Set an option (e.g. set port 80)")
	fmt.Println("  route list    - View HTML mappings for this pot")
	fmt.Println("  route add     - Add route (e.g. route add /wp-login templates/login.html)")
	fmt.Println("  route del     - Delete route (e.g. route del /admin)")
	fmt.Println("  start         - Start the selected honeypot")
	fmt.Println("  stop          - Stop the selected honeypot")
	fmt.Println("  logs          - View static history logs (logs * for all)")
	fmt.Println("  watch         - LIVE view of logs (Press Enter to exit)")
	fmt.Println("  inspect [id]  - View request dump details (e.g. inspect 3)")
	fmt.Println("  history       - Show previously typed commands")
	fmt.Println("  export id|all - Export request(s) to dumps/ folder")
	fmt.Println("  back          - Return to main context")
	fmt.Println("  exit          - Exit the framework\n")
}

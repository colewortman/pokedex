package main
import (
	"fmt"
	"strings"
	"bufio"
	"os"
)

type cliCommand struct {
	name		string
	description string
	callback	func() error
}

var commandRegistry = map[string]cliCommand{
    "exit": {
        name:        "exit",
        description: "Exit the Pokedex",
        callback:    commandExit,
    },
}

func cleanInput(text string) []string {
	lowercase := strings.ToLower(text)
	words := strings.Fields(lowercase)

	return words
}

func commandExit() error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")

		if scanner.Scan() {
			args := cleanInput(scanner.Text())
			if len(args) == 0 {
				continue
			}

			command := args[0]
			
			if cmd, found := commandRegistry[command]; found {
				if err := cmd.callback(); err != nil {
					fmt.Println("Error:", err)
				}
			} else {
				fmt.Println("Unknown command")
			}
		}
	}
}
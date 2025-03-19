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

var commandRegistry map[string]cliCommand

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

func commandHelp() error {
	fmt.Println("Usage:\n\n")
	for _, cmd := range commandRegistry {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	return nil
}

func init() {
	commandRegistry = map[string]cliCommand{
		"help": {
			name: "help",
			description: "Displays a help message",
			callback: commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
	}
}

func main() {
	fmt.Println("Welcome to the Pokedex!")

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
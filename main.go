package main
import (
	"fmt"
	"strings"
	"bufio"
	"os"
)

func cleanInput(text string) []string {
	lowercase := strings.ToLower(text)
	words := strings.Fields(lowercase)

	return words
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		if scanner.Scan() {
			fmt.Println("Your command was:", cleanInput(scanner.Text())[0])
		}
	}
}
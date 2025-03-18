package main
import (
	"fmt"
	"strings"
)

func cleanInput(text string) []string {
	lowercase := strings.ToLower(text)
	words := strings.Fields(lowercase)

	return words
}

func main() {
	fmt.Println("Hello, World!")
}
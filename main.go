package main
import (
	"fmt"
	"strings"
	"bufio"
	"os"
	"io"
	"encoding/json"
	"net/http"
	"time"
	"github.com/colewortman/pokedex/pokecache"
	"math/rand"
)

type Config struct {
	NextURL string
	PrevURL string
	Cache *pokecache.Cache
	Pokedex map[string]Pokemon
}

type Pokemon struct {
	Name		string	`json:"name"`
	Height		int		`json:"height"`
	Weight		int		`json:"weight"`
	BaseExperience int  `json:"base_experience"`
	Stats		[]Stat	`json:"stats"`
	Types		[]Type	`json:"types"`
}

type Stat struct {
	Stat StatDetail `json:"stat"`
	Base int		`json:"base_stat"`
}

type StatDetail struct {
	Name string `json:"name"`
}

type Type struct {
	Type TypeDetail `json:"type"`
}

type TypeDetail struct {
	Name string `json:"name"`
}

type LocationResponse struct {
	Results []struct {
		Name string
	}
	Next	 string
	Previous string
}

type cliCommand struct {
	name		string
	description string
	callback	func(*Config, []string) error
}

var commandRegistry map[string]cliCommand

func cleanInput(text string) []string {
	lowercase := strings.ToLower(text)
	words := strings.Fields(lowercase)

	return words
}

func commandExit(cfg *Config, args []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *Config, args []string) error {
	fmt.Println("Usage:")
	for _, cmd := range commandRegistry {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	return nil
}

func fetchLocations(url string, cache *pokecache.Cache) (*LocationResponse, error) {

	if data, found := cache.Get(url); found {
		fmt.Println("Cache hit! Using cached data.")
		var cachedData LocationResponse
		if err := json.Unmarshal(data, &cachedData); err != nil {
			return nil, fmt.Errorf("failed to parse cached JSON: %v", err)
		}
		return &cachedData, nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var data LocationResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	cache.Add(url, body)

	return &data, nil
}

func commandMap(cfg *Config, args []string) error {
	url := cfg.NextURL
	if url == "" {
		url = "https://pokeapi.co/api/v2/location-area/"
	}

	data, err := fetchLocations(url, cfg.Cache)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	for _, location := range data.Results {
		fmt.Println(location.Name)
	}

	cfg.NextURL = data.Next
	cfg.PrevURL = data.Previous

	return nil
}

func commandMapBack(cfg *Config, args []string) error {
	if cfg.PrevURL == "" {
		fmt.Println("No previous locations available.")
		return nil
	}

	data, err := fetchLocations(cfg.PrevURL, cfg.Cache)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	for _, location := range data.Results {
		fmt.Println(location.Name)
	}

	cfg.NextURL = data.Next
	cfg.PrevURL = data.Previous

	return nil
}

func commandExplore(cfg *Config, args []string) error {
	if len(args) != 1 {
		fmt.Println("Usage: explore <location-area>")
		return nil
	}

	location := args[0]
	cacheKey := fmt.Sprintf("explore:%s", location)
	
	if data, found := cfg.Cache.Get(cacheKey); found {
		fmt.Println("Using cached data...")
		printPokemonFromData(data)
		return nil
	}

	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", location)
	fmt.Printf("Exploring %s...\n", location)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch location data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Location not found")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	cfg.Cache.Add(cacheKey, body)
	printPokemonFromData(body)

	return nil
}

func printPokemonFromData(data []byte) {
	var result struct {
		PokemonEncounters []struct {
			Pokemon struct {
				Name string
			} `json:"pokemon"`
		} `json:"pokemon_encounters"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		fmt.Println("Error parsing data:", err)
		return
	}

	fmt.Println("Found Pokémon:")
	for _, entry := range result.PokemonEncounters {
		fmt.Println(" -", entry.Pokemon.Name)
	}
}

func commandCatch(cfg *Config, args []string) error {
	if len(args) != 1 {
		fmt.Println("Usage: catch <pokemon>")
		return nil
	}

	pokemonName := args[0]
	fmt.Printf("Throwing a Pokeball at %s...\n", pokemonName)

	if _, exists := cfg.Pokedex[pokemonName]; exists {
		fmt.Printf("You already caught %s!\n", pokemonName)
		return nil
	}

	cacheKey := fmt.Sprintf("pokemon:%s", pokemonName)
	if data, found := cfg.Cache.Get(cacheKey); found {
		return attemptCatch(cfg, data, pokemonName)
	}

	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", pokemonName)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Failed to catch Pokemon data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Pokemon not found")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to read response: %v", err)
	}

	cfg.Cache.Add(cacheKey, body)
	return attemptCatch(cfg, body, pokemonName)
}

func attemptCatch(cfg *Config, data []byte, pokemonName string) error {
	var pokemon Pokemon
	if err := json.Unmarshal(data, &pokemon); err != nil {
		fmt.Println("Error parsing Pokémon data:", err)
		return nil
	}

	catchChance := 1.0 / (1.0 + float64(pokemon.BaseExperience)/100.0)
	rand.Seed(time.Now().UnixNano())
	if rand.Float64() < catchChance {
		fmt.Printf("%s was caught!\n", pokemonName)
		cfg.Pokedex[pokemonName] = pokemon
		fmt.Println("You may now inspect it with the inspect command.")
	} else {
		fmt.Printf("%s escaped!\n", pokemonName)
	}
	return nil
}

func commandInspect(cfg *Config, args []string) error {
	if len(args) != 1 {
		fmt.Println("Usage: inspect <pokemon>")
		return nil
	}

	pokemonName := args[0]

	pokemon, exists := cfg.Pokedex[pokemonName]
	if !exists {
		fmt.Println("You have not caught that Pokemon")
		return nil
	}

	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)

	fmt.Println("Stats:")
	for _, stat := range pokemon.Stats {
		fmt.Printf("\t-%s: %d\n", stat.Stat.Name, stat.Base)
	}

	fmt.Println("Types:")
	for _, t := range pokemon.Types {
		fmt.Printf("\t-%s\n", t.Type.Name)
	}

	return nil
}

func commandPokedex(cfg *Config, args []string) error {
	fmt.Println("Your Pokedex:")

	if len(cfg.Pokedex) == 0 {
		fmt.Println("You have not caught any Pokemon yet.")
		return nil
	}

	for _, pokemon := range cfg.Pokedex {
		fmt.Printf("\t-%s\n", pokemon.Name)
	}
	return nil
}

func init() {
	commandRegistry = map[string]cliCommand{
		"help": {
			name: 		 "help",
			description: "Displays a help message",
			callback: commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"map": {
			name: 		 "map",
			description: "Display the next 20 map locations",
			callback: commandMap,
		},
		"mapb": {
			name: 		 "mapb",
			description: "Display the previous 20 map locations",
			callback: commandMapBack,
		},
		"explore": {
			name: 		 "explore",
			description: "Display the pokemon in a given location",
			callback: commandExplore,
		},
		"catch": {
			name: 		 "catch",
			description: "Attempt to catch and store a given Pokemon in the Pokedex",
			callback: commandCatch,
		},
		"inspect": {
			name: 		 "inspect",
			description: "Display info for a given caught Pokemon",
			callback: commandInspect,
		},
		"pokedex": {
			name: 		 "pokedex",
			description: "Display your Pokedex",
			callback: commandPokedex,
		},
	}
}

func main() {
	fmt.Println("Welcome to the Pokedex!")

	var cache = pokecache.NewCache(5 * time.Minute)
	cfg := &Config{
		Cache: cache,
		Pokedex: make(map[string]Pokemon),
	}
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
				if err := cmd.callback(cfg, args[1:]); err != nil {
					fmt.Println("Error:", err)
				}
			} else {
				fmt.Println("Unknown command")
			}
		}
	}
}
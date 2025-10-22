package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("Welcome to your smart kitchen assistant!")

	// Prompt the user
	fmt.Print("Enter the ingredients you have (comma separated): ")

	// Read input from the user
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading input:", err)
		return
	}

	// Clean and split the input
	input = strings.TrimSpace(input)
	ingredients := strings.Split(input, ",")

	// Trim spaces around each ingredient
	for i := range ingredients {
		ingredients[i] = strings.TrimSpace(ingredients[i])
	}

	// Print ingredients
	fmt.Println("You have entered:", ingredients)
}

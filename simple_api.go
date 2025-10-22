package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const apiKey = "YOUR_API_KEY" 

const endpoint = "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key="

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

type RequestBody struct {
	Contents []Content `json:"contents"`
}

type Candidate struct {
	Content Content `json:"content"`
}

type ResponseBody struct {
	Candidates []Candidate `json:"candidates"`
}

func main() {
	prompt := "Suggest a healthy recipe with spinach and garlic."

	// Build request body
	reqBody := RequestBody{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
	}

	// Convert to JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	// Send request
	url := endpoint + apiKey
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Request failed:", err)
		return
	}
	defer resp.Body.Close()

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	// Parse response JSON
	var response ResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error decoding response JSON:", err)
		fmt.Println("Raw response:", string(body))
		return
	}

	// Print the response
	if len(response.Candidates) > 0 {
		fmt.Println("\n🍽️ Gemini says:")
		fmt.Println(response.Candidates[0].Content.Parts[0].Text)
	} else {
		fmt.Println("No response from Gemini.")
	}
}

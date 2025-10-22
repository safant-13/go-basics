package main

import (
	"bufio" // Added for reading user input
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings" // Added for trimming user input

	"github.com/joho/godotenv"
)

const (
	// Choose a suitable Gemini model for text generation
	TEXT_MODEL = "gemini-pro"
	// Choose an Imagen model for image generation
	// Note: Imagen models often require Vertex AI specific endpoints/configurations.
	// If you face issues with this, we might need to adjust or use a different model.
	IMAGE_MODEL = "imagen-4.0-generate-preview-06-06"
	// Replace with the desired output image file name
	OUTPUT_IMAGE_FILE = "generated_image.png"
)

// GenerativeContentRequest struct for text generation
type GenerativeContentRequest struct {
	Contents []Content `json:"contents"`
}

// Content struct represents a part of the prompt
type Content struct {
	Parts []Part `json:"parts"`
}

// Part struct represents text in the prompt
type Part struct {
	Text string `json:"text"`
}

// GenerativeContentResponse struct for text generation response
type GenerativeContentResponse struct {
	Candidates []Candidate `json:"candidates"`
}

// Candidate struct within the text generation response
type Candidate struct {
	Content Content `json:"content"`
}

// ImagenRequest struct for image generation
type ImagenRequest struct {
	Instances []ImagenInstance `json:"instances"`
	Parameters ImagenParameters `json:"parameters"`
}

// ImagenInstance struct
type ImagenInstance struct {
	Prompt string `json:"prompt"`
}

// ImagenParameters struct for image generation settings
type ImagenParameters struct {
	SampleCount  int    `json:"sampleCount"`
	SampleImageSize string `json:"sampleImageSize"` // e.g., "1024x1024"
}

// ImagenResponse struct for image generation response
type ImagenResponse struct {
	Predictions []ImagenPrediction `json:"predictions"`
}

// ImagenPrediction struct within the image generation response
type ImagenPrediction struct {
	BytesBase64Encoded string `json:"bytesBase64Encoded"`
	MimeType           string `json:"mimeType"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ Warning: could not load .env file. Proceeding with system env vars.")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("❌ GEMINI_API_KEY not found in environment")
	}

	// --- 1. Get user input for the question ---
	fmt.Print("What question or prompt would you like to ask (for text and image)? ")
	
	// FIX: Use bufio.NewReader(os.Stdin) to read from standard input
	reader := bufio.NewReader(os.Stdin)
	inputBytes, _ := reader.ReadString('\n')
	userPrompt := strings.TrimSpace(inputBytes) // Use strings.TrimSpace to remove newline and other whitespace

	if userPrompt == "" {
		log.Fatal("❌ No prompt provided.")
	}

	fmt.Printf("You asked: \"%s\"\n", userPrompt)

	// --- 2. Call the Gemini Text Model ---
	fmt.Println("\n--- Getting Text Answer from Gemini ---")
	textAPIURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", TEXT_MODEL, apiKey)

	textRequest := GenerativeContentRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: userPrompt},
				},
			},
		},
	}

	textRequestBody, err := json.Marshal(textRequest)
	if err != nil {
		log.Fatalf("Failed to marshal text request: %v", err)
	}

	textResp, err := http.Post(textAPIURL, "application/json", bytes.NewBuffer(textRequestBody))
	if err != nil {
		log.Fatalf("HTTP request to text model failed: %v", err)
	}
	defer textResp.Body.Close()

	textBodyBytes, err := io.ReadAll(textResp.Body)
	if err != nil {
		log.Fatalf("Reading text response body failed: %v", err)
	}

	if textResp.StatusCode != 200 {
		log.Fatalf("Text API error %d: %s", textResp.StatusCode, string(textBodyBytes))
	}

	var textResponse GenerativeContentResponse
	err = json.Unmarshal(textBodyBytes, &textResponse)
	if err != nil {
		log.Fatalf("Failed to parse text response JSON: %v\nRaw: %s", err, string(textBodyBytes))
	}

	var generatedText string
	if len(textResponse.Candidates) > 0 && len(textResponse.Candidates[0].Content.Parts) > 0 {
		generatedText = textResponse.Candidates[0].Content.Parts[0].Text
		fmt.Println("Gemini's Answer:", generatedText)
	} else {
		fmt.Println("No text answer generated.")
		generatedText = userPrompt // Fallback to user prompt for image if no text answer
	}

	// --- 3. Call the Imagen Image Generation Model ---
	fmt.Println("\n--- Generating Image with Imagen ---")
	
	imagePrompt := userPrompt
    if generatedText != "" && generatedText != userPrompt {
        imagePrompt = generatedText
        fmt.Printf("Using Gemini's answer as image prompt: \"%s\"\n", imagePrompt)
    } else {
        fmt.Printf("Using original prompt as image prompt: \"%s\"\n", imagePrompt)
    }

	// Using the generativelanguage.googleapis.com endpoint for Imagen as an assumption.
	// As discussed, this might need adjustment to a Vertex AI endpoint if it fails.
	imageAPIURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:predict?key=%s", IMAGE_MODEL, apiKey)


	imageRequest := ImagenRequest{
		Instances: []ImagenInstance{
			{Prompt: imagePrompt},
		},
		Parameters: ImagenParameters{
			SampleCount:  1,
			SampleImageSize: "1024x1024", // You can adjust this size
		},
	}

	imageRequestBody, err := json.Marshal(imageRequest)
	if err != nil {
		log.Fatalf("Failed to marshal image request: %v", err)
	}

	imageResp, err := http.Post(imageAPIURL, "application/json", bytes.NewBuffer(imageRequestBody))
	if err != nil {
		log.Fatalf("HTTP request to image model failed: %v", err)
	}
	defer imageResp.Body.Close()

	imageBodyBytes, err := io.ReadAll(imageResp.Body)
	if err != nil {
		log.Fatalf("Reading image response body failed: %v", err)
	}

	if imageResp.StatusCode != 200 {
		log.Fatalf("Image API error %d: %s", imageResp.StatusCode, string(imageBodyBytes))
	}

	var imagenResponse ImagenResponse
	err = json.Unmarshal(imageBodyBytes, &imagenResponse)
	if err != nil {
		log.Fatalf("Failed to parse image response JSON: %v\nRaw: %s", err, string(imageBodyBytes))
	}

	if len(imagenResponse.Predictions) > 0 {
		base64Image := imagenResponse.Predictions[0].BytesBase64Encoded
		decodedImage, err := base64.StdEncoding.DecodeString(base64Image)
		if err != nil {
			log.Fatalf("Failed to decode base64 image: %v", err)
		}

		err = os.WriteFile(OUTPUT_IMAGE_FILE, decodedImage, 0644)
		if err != nil {
			log.Fatalf("Failed to save image to file: %v", err)
		}
		fmt.Printf("✅ Image successfully saved to %s\n", OUTPUT_IMAGE_FILE)
	} else {
		fmt.Println("No image generated or found in the response.")
	}

	fmt.Println("\nProgram finished.")
}
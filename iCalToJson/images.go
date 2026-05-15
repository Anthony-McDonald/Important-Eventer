package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
)

// images maps image filenames to semantic tags used for embedding comparison.
var images = map[string]string{
	"car.jpg":         "car driving travel road trip transport taxi commute fuel petrol traffic highway journey",
	"celebration.jpg": "birthday celebration party success achievement congratulations cake balloons festivities",
	"food.jpg":        "food eating meal dining restaurant drink coffee breakfast lunch dinner takeaway bar pub cuisine",
	"health.jpg":      "health fitness gym exercise workout doctor dentist hospital medicine sleep recovery wellbeing",
	"person.jpg":      "person meeting visit guest arrival friend family social meetup conversation interaction",
	"plane.jpg":       "international flight airport long distance europe usa asia holiday vacation flying abroad",
	"shopping.jpg":    "shopping buying purchase retail store supermarket mall order delivery pickup spending goods",
	"theatre.jpg":     "theatre cinema movie film performance play show concert stage acting entertainment drama",
	"train.jpg":       "uk rail train railway station london manchester birmingham commute intercity eurostar short distance europe travel",
	"work.jpg":        "work job office business meeting laptop email commute productivity task deadline professional",
}

// EmbeddingResponse represents the response returned by the embedding API.
type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

// getEmbeddings is the real implementation that calls Ollama to generate embeddings.
func getEmbeddings(embeddingVectorURL string, text string) ([]float64, error) {
	body := map[string]string{
		"model":  "nomic-embed-text",
		"prompt": text,
	}

	b, _ := json.Marshal(body)

	resp, err := http.Post(
		fmt.Sprintf("%s/api/embeddings", embeddingVectorURL),
		"application/json",
		bytes.NewBuffer(b),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out EmbeddingResponse
	json.NewDecoder(resp.Body).Decode(&out)

	return out.Embedding, nil
}

// embedFn is a function variable that allows embedding logic to be replaced in tests.
// By default, it uses the real embed function.
var embedFn = getEmbeddings

// cosineSimilarity calculates similarity between two embedding vectors using dot product normalization.
func cosineSimilarity(a, b []float64) float64 {
	var dot, normA, normB float64

	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// chooseImage selects the most relevant image based on semantic similarity
// between the event description and image tag embeddings.
func chooseImage(embeddingVectorURL string, event string) string {
	eventVec, _ := embedFn(embeddingVectorURL, event)

	bestScore := -1.0
	bestImage := "default.png"

	for file, tags := range images {
		vec, _ := embedFn(embeddingVectorURL, tags)

		score := cosineSimilarity(eventVec, vec)

		if score > bestScore {
			bestScore = score
			bestImage = file
		}
	}

	return bestImage
}

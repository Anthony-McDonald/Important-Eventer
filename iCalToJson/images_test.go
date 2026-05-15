package main

import (
	"math"
	"testing"
)

// TestCosineSimilarity verifies correct vector similarity calculations.
func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a        []float64
		b        []float64
		expected float64
	}{
		{"identical vectors", []float64{1, 2, 3}, []float64{1, 2, 3}, 1},
		{"orthogonal vectors", []float64{1, 0}, []float64{0, 1}, 0},
		{"scaled same direction", []float64{1, 0}, []float64{2, 0}, 1},
		{"opposite direction", []float64{1, 0}, []float64{-1, 0}, -1},
	}

	for _, tC := range tests {
		t.Run(tC.name, func(t *testing.T) {
			got := cosineSimilarity(tC.a, tC.b)

			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Fatalf("invalid result: %v", got)
			}

			if math.Abs(got-tC.expected) > 1e-6 {
				t.Errorf("expected %v, got %v", tC.expected, got)
			}
		})
	}
}

// TestChooseImage verifies image selection logic using a mocked embedding function.
func TestChooseImage(t *testing.T) {
	// Override image dataset for deterministic behavior.
	images = map[string]string{
		"car.png":  "car road travel",
		"food.png": "food meal eating",
	}

	// Mock embedding behavior.
	embedFn = func(_ string, text string) ([]float64, error) {
		switch text {
		case "car accident":
			return []float64{1, 0}, nil
		case "car road travel":
			return []float64{1, 0}, nil
		case "food meal eating":
			return []float64{0, 1}, nil
		default:
			return []float64{0, 0}, nil
		}
	}

	t.Cleanup(func() {
		embedFn = getEmbeddings // restore real function after test
	})

	got := chooseImage("ignored", "car accident")

	if got != "car.png" {
		t.Errorf("expected car.png, got %s", got)
	}
}

package main

import (
	"testing"
)

func TestCleanFileName(t *testing.T) {
	cfg, _, err := loadConfig("config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "The.Aviator.1080p.avi",
			expected: "The Aviator",
		},
		{
			input:    "Shadow.2020.HD.mkv",
			expected: "Shadow 2020",
		},
		{
			input:    "Morgan.ORG.720p.mkv",
			expected: "Morgan",
		},
		{
			input:    "Inception (2010) [1080p] [BluRay] [Dual Audio] (LiNE).mp4",
			expected: "Inception (2010)",
		},
		{
			input:    "www.10xflix.com - Movie.Title.2023.HDRip.mkv",
			expected: "Movie Title 2023",
		},
		{
			input:    "Doctor.Strange.in.the.Multiverse.of.Madness.2022.1080p.WEB-DL.HinEng.mkv",
			expected: "Doctor Strange in the Multiverse of Madness 2022",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := cleanFileName(tt.input, cfg.BuzzWords, cfg.SeparatorWords)
			if got != tt.expected {
				t.Errorf("cleanFileName(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

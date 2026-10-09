package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// Config represents the structure of config.json
type Config struct {
	DirName           string   `json:"DIR_NAME"`
	BuzzWords         []string `json:"BUZZ_WORDS"`
	SeparatorWords    []string `json:"SEPARATOR_WORDS"`
	AllowedExtensions []string `json:"ALLOWED_EXTENSIONS"`
}

var (
	isWordRegex        = regexp.MustCompile(`^\w+$`)
	emptyBracketsRegex = regexp.MustCompile(`\(\s*\)|\[\s*\]|\{\s*\}`)
)

// loadConfig reads and validates the configuration file.
func loadConfig(configPath string) (*Config, map[string]bool, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read config file '%s': %w", configPath, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, nil, fmt.Errorf("failed to parse JSON in '%s': %w", configPath, err)
	}

	if strings.TrimSpace(cfg.DirName) == "" {
		return nil, nil, fmt.Errorf("DIR_NAME is not defined in the config file")
	}

	if len(cfg.AllowedExtensions) == 0 {
		return nil, nil, fmt.Errorf("ALLOWED_EXTENSIONS is not defined in the config file")
	}

	allowedExtMap := make(map[string]bool)
	for _, ext := range cfg.AllowedExtensions {
		cleanExt := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(ext), "."))
		if cleanExt != "" {
			allowedExtMap[cleanExt] = true
		}
	}

	return &cfg, allowedExtMap, nil
}

// cleanFileName cleans a movie file stem by removing buzzwords and replacing separators.
func cleanFileName(fileName string, buzzWords []string, separatorWords []string) string {
	ext := filepath.Ext(fileName)
	stem := strings.TrimSuffix(fileName, ext)

	// 1. Remove buzzwords (case-insensitive, word boundaries if alphanumeric)
	for _, word := range buzzWords {
		trimmedWord := strings.TrimSpace(word)
		if trimmedWord == "" {
			continue
		}

		var pattern string
		if isWordRegex.MatchString(trimmedWord) {
			pattern = `(?i)\b` + regexp.QuoteMeta(trimmedWord) + `\b`
		} else {
			pattern = `(?i)` + regexp.QuoteMeta(trimmedWord)
		}

		re, err := regexp.Compile(pattern)
		if err == nil {
			stem = re.ReplaceAllString(stem, " ")
		}
	}

	// 2. Replace separators with spaces
	for _, sep := range separatorWords {
		if sep != "" {
			stem = strings.ReplaceAll(stem, sep, " ")
		}
	}

	// 3. Clean up empty brackets/parentheses like (), [], or {}
	stem = emptyBracketsRegex.ReplaceAllString(stem, " ")

	// 4. Normalize multiple whitespace into a single space
	cleanName := strings.Join(strings.Fields(stem), " ")
	return strings.TrimSpace(cleanName)
}

// processDirectory iterates through files in dirPath and previews/applies renames.
func processDirectory(
	dirPath string,
	buzzWords []string,
	separatorWords []string,
	allowedExtensions map[string]bool,
	applyRename bool,
	autoConfirm bool,
) {
	fileInfo, err := os.Stat(dirPath)
	if err != nil || !fileInfo.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: Directory does not exist: %s\n", dirPath)
		return
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory '%s': %v\n", dirPath, err)
		return
	}

	// Filter allowed media files
	var matchedFiles []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(entry.Name()), "."))
		if allowedExtensions[ext] {
			matchedFiles = append(matchedFiles, entry)
		}
	}

	if len(matchedFiles) == 0 {
		fmt.Printf("No matching media files found in: %s\n", dirPath)
		return
	}

	// Sort files alphabetically
	sort.Slice(matchedFiles, func(i, j int) bool {
		return matchedFiles[i].Name() < matchedFiles[j].Name()
	})

	fmt.Printf("Found %d file(s) in %s\n%s\n", len(matchedFiles), dirPath, strings.Repeat("-", 70))

	reader := bufio.NewReader(os.Stdin)

	for _, entry := range matchedFiles {
		origName := entry.Name()
		origExt := filepath.Ext(origName)
		cleanedStem := cleanFileName(origName, buzzWords, separatorWords)

		if cleanedStem == "" {
			fmt.Printf("[SKIP] Cleaning resulted in empty name for: '%s'\n\n", origName)
			continue
		}

		newFileName := cleanedStem + strings.ToLower(origExt)
		if origName == newFileName {
			fmt.Printf("[UNCHANGED] '%s'\n\n", origName)
			continue
		}

		fmt.Printf("Original : %s\n", origName)
		fmt.Printf("Processed: %s\n", newFileName)

		if !applyRename {
			fmt.Printf("[DRY-RUN] Pass 'rename' or '--rename' to apply changes.\n\n")
			continue
		}

		targetPath := filepath.Join(dirPath, newFileName)
		sourcePath := filepath.Join(dirPath, origName)

		if _, err := os.Stat(targetPath); err == nil {
			fmt.Printf("[ERROR] Destination file already exists: '%s'. Skipping to avoid overwrite.\n\n", newFileName)
			continue
		}

		if !autoConfirm {
			fmt.Print("Rename this file? (yes/no): ")
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading input. Skipping.\n\n")
				continue
			}
			input = strings.TrimSpace(strings.ToLower(input))
			if input != "y" && input != "yes" {
				fmt.Printf("Skipped.\n\n")
				continue
			}
		}

		if err := os.Rename(sourcePath, targetPath); err != nil {
			fmt.Printf("[FAILED] Could not rename '%s': %v\n\n", origName, err)
		} else {
			fmt.Printf("[SUCCESS] Renamed '%s' -> '%s'\n\n", origName, newFileName)
		}
	}
}

func runCommand(configPath string, applyRename bool, autoConfirm bool) {
	cfg, allowedExtMap, err := loadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	processDirectory(
		cfg.DirName,
		cfg.BuzzWords,
		cfg.SeparatorWords,
		allowedExtMap,
		applyRename,
		autoConfirm,
	)
}

func main() {
	var configPath string
	var renameFlag bool
	var autoConfirm bool

	rootCmd := &cobra.Command{
		Use:   "movie-name-changer [command]",
		Short: "Clean and rename movie filenames by removing unwanted buzzwords",
		Long: `A CLI tool to clean buzzwords, tags, codecs, and metadata from movie filenames 
and safely rename them based on your configuration in config.json.`,
		Run: func(cmd *cobra.Command, args []string) {
			runCommand(configPath, renameFlag, autoConfirm)
		},
	}

	// Persistent flags (available to root and all subcommands)
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "config.json", "Path to configuration file")

	// Flags for root preview/rename
	rootCmd.Flags().BoolVarP(&renameFlag, "rename", "r", false, "Perform the actual renaming (default is preview/dry-run)")
	rootCmd.Flags().BoolVarP(&autoConfirm, "yes", "y", false, "Auto-confirm all renames without prompting")

	// 'preview' subcommand
	previewCmd := &cobra.Command{
		Use:   "preview",
		Short: "Preview filename changes without renaming any files (dry-run)",
		Run: func(cmd *cobra.Command, args []string) {
			runCommand(configPath, false, false)
		},
	}

	// 'rename' subcommand
	var renameCmdAutoConfirm bool
	renameCmd := &cobra.Command{
		Use:   "rename",
		Short: "Perform actual file renaming",
		Run: func(cmd *cobra.Command, args []string) {
			runCommand(configPath, true, renameCmdAutoConfirm || autoConfirm)
		},
	}
	renameCmd.Flags().BoolVarP(&renameCmdAutoConfirm, "yes", "y", false, "Auto-confirm all renames without prompting")

	rootCmd.AddCommand(previewCmd, renameCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

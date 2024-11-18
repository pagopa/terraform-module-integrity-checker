package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	modulesDir = ".terraform/modules"
	hashesFile = ".module_hashes.json"
	red        = "\033[0;31m"
	nc         = "\033[0m" // No Color
)

type ModuleHash struct {
	Module string `json:"module"`
	Hash   string `json:"hash"`
}

type HashStore struct {
	Hashes map[string]string `json:"hashes"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Missed action: init, apply, plan ect.")
		os.Exit(0)
	}

	action := os.Args[1]

	// Load existing hashes
	hashStore, err := loadHashes(hashesFile)
	if err != nil {
		fmt.Printf("Error loading hashes: %v\n", err)
		os.Exit(1)
	}

	// Download modules
	cmd := exec.Command("terraform", append([]string{action}, os.Args[2:]...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running terraform: %v\n", err)
		os.Exit(1)
	}

	if action == "init" {
		modules, err := filepath.Glob(filepath.Join(modulesDir, "*"))
		if err != nil {
			fmt.Printf("Error reading modules: %v\n", err)
			os.Exit(1)
		}

		for _, modulePath := range modules {
			//fmt.Println(modulePath)
			if info, err := os.Stat(modulePath); err == nil && info.IsDir() {
				moduleName := filepath.Base(modulePath)

				// Calculate hash of the downloaded module
				newHash, err := calculateHash(modulePath)
				if err != nil {
					fmt.Printf("Error calculating hash: %v\n", err)
					continue
				}

				// Compare with existing hash
				if oldHash, exists := hashStore.Hashes[moduleName]; exists {
					if oldHash != newHash {
						fmt.Printf("%sThe module %s has been changed!%s\n", red, moduleName, nc)
						os.Exit(-1)
					}
				} else {
					// Save new hash if no previous hash exists
					// fmt.Printf("Salvataggio del nuovo hash del modulo %s.\n", moduleName)
				}

				// Update the hash store
				hashStore.Hashes[moduleName] = newHash
			}
		}

		// Save updated hashes back to the JSON file
		if err := saveHashes(hashesFile, hashStore); err != nil {
			fmt.Printf("Error saving hashes: %v\n", err)
		}
	}
}

func calculateHash(path string) (string, error) {
	cmd := exec.Command("tar", "-cf", "-", path)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error creating tar: %v", err)
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, &out); err != nil {
		return "", fmt.Errorf("error calculating hash: %v", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func loadHashes(filePath string) (HashStore, error) {
	var store HashStore
	store.Hashes = make(map[string]string)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return store, nil // Return empty store if file does not exist
		}
		return store, err
	}

	if err := json.Unmarshal(data, &store); err != nil {
		return store, err
	}

	return store, nil
}

func saveHashes(filePath string, store HashStore) error {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

package main
// Scopo: implementazione CLI per verifica integrità moduli Terraform
import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	modulesDir          = ".terraform/modules"
	modulesMetadataFile = ".terraform/modules/modules.json"
	registryURL         = "registry.terraform.io"
	hashesFile          = ".module_hashes.json"
)

type ModulesMetadata struct {
	Modules []struct {
		Source string `json:"Source"`
		Key    string `json:"Key"`
	} `json:"Modules"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: tf <command> [args...]")
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "init":
		runTerraform(command, args...)
		if err := handleInit(); err != nil {
			fmt.Printf("Error during init: %v\n", err)
			os.Exit(1)
		}
	default:
		// Pass other commands to Terraform directly
		runTerraform(command, args...)
	}
}

func handleInit() error {
	// Check if modules directory exists
	if _, err := os.Stat(modulesDir); errors.Is(err, os.ErrNotExist) {
		fmt.Println("No modules directory found. Skipping module check.")
		return nil
	}

	// Check if modules metadata file exists
	if _, err := os.Stat(modulesMetadataFile); errors.Is(err, os.ErrNotExist) {
		fmt.Println("No modules metadata file found. Skipping module check.")
		return nil
	}

	// Read modules metadata
	metadataContent, err := ioutil.ReadFile(modulesMetadataFile)
	if err != nil {
		return fmt.Errorf("error reading modules metadata file: %w", err)
	}

	var metadata ModulesMetadata
	if err := json.Unmarshal(metadataContent, &metadata); err != nil {
		return fmt.Errorf("error parsing modules metadata file: %w", err)
	}

	// Filter registry modules
	registryModules := []string{}
	for _, module := range metadata.Modules {
		if strings.Contains(module.Source, registryURL) {
			registryModules = append(registryModules, module.Key)
		}
	}

	if len(registryModules) == 0 {
		fmt.Println("No Terraform modules from the registry were found. No lock file check needed.")
		return nil
	}

	// Check if lock file exists
	hashes := make(map[string]string)
	if _, err := os.Stat(hashesFile); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("%s is missing. Creating a new lock file...\n", hashesFile)
	} else {
		// Read lock file if it exists
		hashesContent, err := ioutil.ReadFile(hashesFile)
		if err != nil {
			return fmt.Errorf("error reading hashes file: %w", err)
		}
		if err := json.Unmarshal(hashesContent, &hashes); err != nil {
			return fmt.Errorf("error parsing hashes file: %w", err)
		}
	}

	// Validate module hashes and update lock file if needed
	updatedHashes := make(map[string]string)
	for _, moduleKey := range registryModules {
		modulePath := filepath.Join(modulesDir, moduleKey)
		if _, err := os.Stat(modulePath); errors.Is(err, os.ErrNotExist) {
			fmt.Printf("Module path %s not found.\n", modulePath)
			continue
		}

		newHash, err := calculateHash(modulePath)
		if err != nil {
			return fmt.Errorf("error calculating hash for module %s: %w", moduleKey, err)
		}

		moduleName := filepath.Base(modulePath)
		if previousHash, exists := hashes[moduleName]; exists {
			if previousHash != newHash {
				fmt.Printf("The module %s has changed! Exiting...\n", moduleName)
				return errors.New("hash consistency check failed")
			}
		} else {
			fmt.Printf("The hash for module %s is not present in the lock file. Adding it...\n", moduleName)
		}

		// Update hash in the map
		updatedHashes[moduleName] = newHash
	}

	// Write updated hashes to lock file
	if len(updatedHashes) > 0 {
		if err := writeLockFile(hashesFile, updatedHashes); err != nil {
			return fmt.Errorf("error writing lock file: %w", err)
		}
		fmt.Printf("%s has been successfully updated.\n", hashesFile)
	}

	// Run Terraform init
	runTerraform("init")
	return nil
}

// isHidden checks if a file or directory is hidden (starts with a dot)
func isHidden(name string) bool {
	return len(name) > 0 && name[0] == '.'
}

func calculateHash(path string) (string, error) {

	hasher := sha256.New()

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories
		if info.IsDir() && isHidden(info.Name()) {
			return filepath.SkipDir
		}

		// If it's a file, read its contents and update the hash
		if !info.IsDir() {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			hasher.Write(data)
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash), nil
}

func writeLockFile(filename string, hashes map[string]string) error {
	data, err := json.MarshalIndent(hashes, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

func runTerraform(command string, args ...string) {
	cmdArgs := append([]string{command}, args...)
	cmd := exec.Command("terraform", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Printf("Terraform %s failed: %v\n", command, err)
		os.Exit(1)
	}
}

# Terraform Module Integrity Checker

## Overview

This Go application is a command-line tool designed to work with Terraform. It helps manage Terraform modules by ensuring their integrity through hash verification. The tool performs actions such as `init`, `apply`, and `plan`, while also checking for changes in the downloaded modules.

## Features

- **Action Execution**: Supports executing Terraform commands like `init`, `apply`, and `plan`.
- **Module Integrity Check**: Calculates and stores SHA-256 hashes of downloaded Terraform modules to detect any changes.
- **Error Handling**: Provides informative error messages for various failure scenarios, such as loading hashes, running Terraform commands, and calculating hashes.

## Prerequisites

- Go (version 1.16 or later)
- Terraform (installed and available in the system PATH)
- Access to the filesystem for reading and writing module hashes

## Installation

1. Clone the repository:
   ```bash
   git clone terraform-module-integrity-checker
   cd terraform-module-integrity-checker
   ```

2. Build the application:
   ```bash
   make build
   ```

3. (Optional) Install the binary:
   ```bash
   make install
   ```

## Usage

Run the application with the desired Terraform action as the first argument, followed by any additional arguments required by the action. For example:

```bash
./tf init
```

### Supported Actions

- `init`: Initializes the Terraform working directory and downloads the required modules.
- `apply`: Applies the changes required to reach the desired state of the configuration.
- `plan`: Creates an execution plan, showing what actions Terraform will take.

### Example

To initialize a Terraform configuration and check for module integrity:

```bash
./tf init
```

## How It Works

1. **Loading Existing Hashes**: The application attempts to load existing module hashes from a JSON file (`.module_hashes.json`).
2. **Executing Terraform Command**: It runs the specified Terraform command using the `os/exec` package.
3. **Hash Calculation**: If the action is `init`, it calculates the SHA-256 hash of each downloaded module and compares it with the previously stored hash.
4. **Integrity Check**: If a module's hash has changed, the application notifies the user and exits with an error.
5. **Saving Hashes**: After processing, it saves the updated hashes back to the JSON file.

## Error Handling

The application provides error messages for various scenarios, including:

- Missing action argument
- Errors loading or saving hashes
- Errors running Terraform commands
- Errors calculating module hashes

# Testing Suite for GIST

## Overview

This project now includes a comprehensive testing suite with the following components:

## Tests Implemented

### Unit Tests

1. **Configuration Path Handling**
   - Tests `getConfigPath()` with default and custom environment variables

2. **Git Path Resolution**
   - Tests `getGitPath()` with various environment variable configurations

3. **Key-Value Parsing**
   - Tests `parseKeyValue()` with various formats including quoted values and leading dashes

4. **Configuration Loading**
   - Tests `loadConfig()` with realistic YAML content
   - Validates profile parsing and structure

5. **Configuration Saving**
   - Tests `saveConfig()` with proper YAML generation
   - Verifies correct file structure and content

6. **Profile Management**
   - Tests `findProfile()` for profile lookup
   - Tests `initConfig()` for initialization logic

7. **Interactive Commands**
   - Tests the logic of `commandAdd()` (input simulation)

## Testing Approach

The testing strategy focuses on:
- **Isolation**: Each function is tested independently
- **Coverage**: All major code paths are covered
- **Realistic Data**: Tests use realistic configuration structures
- **Edge Cases**: Handles malformed input gracefully

## Running Tests

### Locally

```bash
# Run all tests with verbose output
go test -v ./...

# Run specific test file
go test -v main_test.go main.go

# Run tests with coverage
go test -cover ./...
```

### GitHub Actions

The workflow in `.github/workflows/test.yml` automatically:
- Checks out the code
- Sets up Go environment
- Installs dependencies
- Runs all tests
- Runs vet and fmt checks
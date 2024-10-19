package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

// Check if the required command is available in the user's PATH
func checkCommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func ensureAirInstalled() {
	if !checkCommandExists("air") {
		log.Fatalf("Error: 'air' is not installed. Please install it by running:\n\n\tcurl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin\n")
	}
}

// createDirectoriesAndFiles creates the necessary project directories and initial files
func createDirectoriesAndFiles(basePath string) {
	// Define directories and their corresponding initial files
	dirs := map[string][]string{
		"cmd":            {"main.go"},
		"pkg/router":     {"router.go"},
		"pkg/middleware": {"middleware.go"},
		"pkg/handlers":   {"handlers.go"},
		"pkg/models":     {"models.go"},
		"pkg/db":         {"db.go"},
		"config":         {"config.yaml"},
	}

	for dir, files := range dirs {
		// Create the directory with the basePath
		dirPath := filepath.Join(basePath, dir)
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			log.Fatalf("Error creating directory %s: %v", dirPath, err)
		} else {
			fmt.Printf("Created directory: %s\n", dirPath)
		}

		// Create initial files in the directory
		for _, file := range files {
			filePath := filepath.Join(dirPath, file)
			f, err := os.Create(filePath)
			if err != nil {
				log.Fatalf("Error creating file %s: %v", filePath, err)
			} else {
				fmt.Printf("Created file: %s\n", filePath)
				defer f.Close()

				// Write initial content to the file from a template
				templateFileName := getTemplateFileName(file)
				writeContentFromTemplate(f, templateFileName)
			}
		}
	}

	// Create the Dockerfile and docker-compose.yaml in the base directory
	dockerFiles := []string{"Dockerfile", "docker-compose.yaml", "Makefile", ".air.toml"}
	for _, file := range dockerFiles {
		filePath := filepath.Join(basePath, file)
		f, err := os.Create(filePath)
		if err != nil {
			log.Fatalf("Error creating file %s: %v", filePath, err)
		} else {
			fmt.Printf("Created file: %s\n", filePath)
			defer f.Close()

			// Write initial content to the file from a template
			templateFileName := getTemplateFileName(file)
			writeContentFromTemplate(f, templateFileName)
		}
	}
}

// getTemplateFileName returns the template file path based on the Go file being created
func getTemplateFileName(fileName string) string {
	// Map the Go file name to a corresponding template file name
	templateMapping := map[string]string{
		"main.go":             "templates/main.txt",
		"router.go":           "templates/routes.txt",
		"middleware.go":       "templates/middleware.txt",
		"handlers.go":         "templates/handlers.txt",
		"models.go":           "templates/models.txt",
		"db.go":               "templates/db.txt",
		"Dockerfile":          "templates/dockers.txt",
		"docker-compose.yaml": "templates/docker.txt",
		"Makefile":            "templates/makerun.txt",
		".air.toml":           "templates/air.txt",
		"config.yaml":         "templates/config.txt",
	}
	return templateMapping[fileName]
}

// writeContentFromTemplate reads the content of the template file and writes it to the target file
func writeContentFromTemplate(f *os.File, templatePath string) {
	// Read the template file
	content, err := ioutil.ReadFile(templatePath)
	if err != nil {
		log.Fatalf("Error reading template file %s: %v", templatePath, err)
	}

	// Write the template content to the new Go file
	_, err = f.Write(content)
	if err != nil {
		log.Fatalf("Error writing content to file %s: %v", f.Name(), err)
	}
}

// runCommand runs a shell command in the given directory
func runCommand(dir, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Error running command: %v", err)
	}
	return err
}

// initializeGoMod initializes a Go module in the project directory
func initializeGoMod(basePath, projectName string) {
	fmt.Println("Initializing Go module...")

	// Run `go mod init <projectName>`
	err := runCommand(basePath, "go", "mod", "init", projectName)
	if err != nil {
		log.Fatalf("Error initializing Go module: %v", err)
	}

	// Run `go mod tidy` to ensure dependencies are added and cleaned
	fmt.Println("Tidying up Go module dependencies...")
	err = runCommand(basePath, "go", "mod", "tidy")
	if err != nil {
		log.Fatalf("Error tidying Go module: %v", err)
	}

	fmt.Println("Go module initialized successfully!")
}

// ExpandTilde expands the tilde (~) in the given path to the user's home directory
func ExpandTilde(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		// Get the current user
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		// Replace ~ with the user's home directory
		return filepath.Join(usr.HomeDir, path[2:]), nil
	}
	return path, nil
}
func main() {
	ensureAirInstalled()
	reader := bufio.NewReader(os.Stdin)

	// Ask for the base path to create the project
	fmt.Print("Enter the base path where you want to create the project: ")
	baseInput, _ := reader.ReadString('\n')
	baseInput = strings.TrimSpace(baseInput)

	if baseInput == "" {
		log.Fatal("Base path cannot be empty")
	}

	// Expand the tilde in the base path
	expandedBasePath, err := ExpandTilde(baseInput)
	if err != nil {
		log.Fatalf("Error expanding base path: %v", err)
	}

	// Ask for the project name
	fmt.Print("Enter the project name: ")
	projectName, _ := reader.ReadString('\n')
	projectName = strings.TrimSpace(projectName)

	if projectName == "" {
		log.Fatal("Project name cannot be empty")
	}

	// Create the full project path
	basePath := filepath.Join(expandedBasePath, projectName)
	if err := os.Mkdir(basePath, 0755); err != nil {
		log.Fatalf("Failed to create project base directory: %v", err)
	} else {
		fmt.Printf("Created project base directory: %s\n", basePath)
	}

	// Create directories and files
	createDirectoriesAndFiles(basePath)

	// Initialize Go module
	initializeGoMod(basePath, projectName)
}

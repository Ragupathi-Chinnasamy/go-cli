package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Go backend project",
	Long:  "This command sets up a new Go backend project with a go.mod, main.go, and .env file",
	Run: func(cmd *cobra.Command, args []string) {
		initializeProject()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

// Function to prompt the user for input and create necessary files
func initializeProject() {
	reader := bufio.NewReader(os.Stdin)

	// Prompt for the project name
	fmt.Print("Enter project name: ")
	projectName, _ := reader.ReadString('\n')
	projectName = strings.TrimSpace(projectName)

	// Prompt for the port number
	fmt.Print("Enter port number: ")
	port, _ := reader.ReadString('\n')
	port = strings.TrimSpace(port)

	// Prompt for the database connection string
	fmt.Print("Enter database connection string: ")
	dbConn, _ := reader.ReadString('\n')
	dbConn = strings.TrimSpace(dbConn)

	// Create the go.mod file
	createGoMod(projectName)

	// Create the main.go file
	createMainGo(projectName)

	// Create the .env file
	createEnvFile(port, dbConn)

	fmt.Println("Project initialized successfully!")
}

// Create go.mod file
func createGoMod(projectName string) {
	goModContent := fmt.Sprintf("module %s\n\ngo 1.20\n", projectName)

	err := os.WriteFile("go.mod", []byte(goModContent), 0644)
	if err != nil {
		fmt.Println("Error creating go.mod file:", err)
	}
}

// Create main.go file
func createMainGo(projectName string) {
	mainGoTemplate := `package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}

	fmt.Printf("Starting %s on port %s...\n", "{{.ProjectName}}", port)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from {{.ProjectName}}!")
	})
	err := http.ListenAndServe(":" + port, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
`

	tmpl, err := template.New("main.go").Parse(mainGoTemplate)
	if err != nil {
		fmt.Println("Error parsing main.go template:", err)
		return
	}

	file, err := os.Create("main.go")
	if err != nil {
		fmt.Println("Error creating main.go file:", err)
		return
	}
	defer file.Close()

	err = tmpl.Execute(file, struct {
		ProjectName string
	}{
		ProjectName: projectName,
	})

	if err != nil {
		fmt.Println("Error writing to main.go file:", err)
	}
}

// Create .env file
func createEnvFile(port string, dbConn string) {
	envContent := fmt.Sprintf("PORT=%s\nDATABASE_URL=%s\n", port, dbConn)

	err := os.WriteFile(".env", []byte(envContent), 0644)
	if err != nil {
		fmt.Println("Error creating .env file:", err)
	}
}

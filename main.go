package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type ProjectSpec struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Framework   string            `json:"framework"`
	Components  []string          `json:"components"`
	Files       map[string]string `json:"files"`
	Description string            `json:"description"`
}

type DevAgent struct {
	client *genai.Client
	model  *genai.GenerativeModel
	ctx    context.Context
}

func NewDevAgent(apiKey string) (*DevAgent, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	model := client.GenerativeModel("gemini-1.5-pro")
	// Set to lower temperature for more focused code generation
	model.SetTemperature(0.2)

	return &DevAgent{
		client: client,
		model:  model,
		ctx:    ctx,
	}, nil
}

func (a *DevAgent) GenerateProjectSpec(prompt string) (*ProjectSpec, error) {
	systemPrompt := `As an AI development agent, analyze the user's request and create a detailed project specification.
Think through this step by step:

1. Understand the core requirements
2. Identify the best framework and technologies
3. Break down the components needed
4. Plan the file structure
5. Create a comprehensive project specification

USER REQUEST: %s

Generate a JSON project specification that includes:
- Project name
- Project type (web, mobile, cli, etc.)
- Framework recommendation
- List of required components
- File structure (provide all the files and their descriptions required for production ready code)
- Project description

Respond only with valid JSON in the following structure:
{
  "name": "<project name>",
  "type": "<project type>",
  "framework": "<recommended framework>",
  "components": [
    "<component 1>",
    "<component 2>",
    ...
  ],
  "files": {
    "<file 1 path>": "<file 1 description and prompt to generate file and import chains>",
    "<file 2 path>": "<file 2 description and prompt to generate file and import chains>",
    ...
  },
  "description": "<project description>"
}
DO NOT respond with anything else. JUST JSON repsonse should start with` + "start with ```json and end with ```"

	fullPrompt := fmt.Sprintf(systemPrompt, prompt)

	resp, err := a.model.GenerateContent(a.ctx, genai.Text(fullPrompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate project spec: %v", err)
	}
	var spec ProjectSpec
	respContentStr := removeFirstAndLastLine(string(resp.Candidates[0].Content.Parts[0].(genai.Text)))
	fmt.Println(respContentStr)
	err = json.Unmarshal([]byte(respContentStr), &spec)
	if err != nil {
		return nil, fmt.Errorf("failed to parse project spec: %v", err)
	}

	return &spec, nil
}

func (a *DevAgent) GenerateCode(spec *ProjectSpec) error {
	fmt.Printf("🚀 Generating project: %s\n", spec.Name)
	fmt.Printf("📋 Type: %s using %s\n", spec.Type, spec.Framework)
	fmt.Println("📁 Generating files...")

	// Create project directory
	projectDir := spec.Name
	err := os.MkdirAll(projectDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create project directory: %v", err)
	}

	// Keep track of generated files and their content
	generatedFiles := make(map[string]string)

	// Sort files to ensure consistent generation order
	var filePaths []string
	for filePath := range spec.Files {
		filePaths = append(filePaths, filePath)
	}
	sort.Strings(filePaths)

	for _, filePath := range filePaths {
		description := spec.Files[filePath]
		fmt.Printf("⚙️  Generating %s...\n", filePath)

		// Build context from previously generated files
		var contextBuilder strings.Builder
		if len(generatedFiles) > 0 {
			contextBuilder.WriteString("\nPreviously generated files:\n")
			for prevPath, content := range generatedFiles {
				contextBuilder.WriteString(fmt.Sprintf("\n%s:\n```\n%s\n```\n", prevPath, content))
			}
		}

		codePrompt := fmt.Sprintf(`Generate the complete code for the file %s in the %s project.
Project Description: %s
File Purpose: %s

Requirements:
- Use %s framework
- Follow best practices
- Include necessary imports
- Add helpful comments
- Make sure the code is complete and functional
- Ensure compatibility with other project files
%s
Generate only the code, no explanations.`, filePath, spec.Name, spec.Description, description, spec.Framework, contextBuilder.String())

		resp, err := a.model.GenerateContent(a.ctx, genai.Text(codePrompt))
		if err != nil {
			return fmt.Errorf("failed to generate code for %s: %v", filePath, err)
		}

		fullPath := filepath.Join(projectDir, filePath)
		err = os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			return fmt.Errorf("failed to create directories for %s: %v", filePath, err)
		}

		fileContent := resp.Candidates[0].Content.Parts[0].(genai.Text)
		fileContentStr := removeFirstAndLastLine(string(fileContent))

		// Store generated content for context in subsequent generations
		generatedFiles[filePath] = fileContentStr

		err = os.WriteFile(fullPath, []byte(fileContentStr), 0644)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %v", filePath, err)
		}
	}

	// Generate README.md with context of all generated files
	var contextBuilder strings.Builder
	for filePath, content := range generatedFiles {
		contextBuilder.WriteString(fmt.Sprintf("\n%s:\n```\n%s\n```\n", filePath, content))
	}

	readmePrompt := fmt.Sprintf(`Generate a comprehensive README.md for the %s project.
Description: %s
Framework: %s
Components: %v

Project Structure:%s

Include:
1. Project overview
2. Setup instructions
3. Usage examples
4. Component descriptions
5. Dependencies
`, spec.Name, spec.Description, spec.Framework, spec.Components, contextBuilder.String())

	resp, err := a.model.GenerateContent(a.ctx, genai.Text(readmePrompt))
	if err != nil {
		return fmt.Errorf("failed to generate README: %v", err)
	}
	readmeContent := removeFirstAndLastLine(string(resp.Candidates[0].Content.Parts[0].(genai.Text)))
	err = os.WriteFile(filepath.Join(projectDir, "README.md"), []byte(readmeContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write README: %v", err)
	}

	fmt.Println("✨ Project generated successfully!")
	return nil
}

func (a *DevAgent) Close() {
	a.client.Close()
}

func main() {
	apiKey := flag.String("api-key", "", "Google API Key for Gemini")
	flag.Parse()

	if *apiKey == "" {
		*apiKey = os.Getenv("GOOGLE_API_KEY")
		if *apiKey == "" {
			fmt.Println("Please provide an API key via -api-key flag or GOOGLE_API_KEY environment variable")
			os.Exit(1)
		}
	}

	agent, err := NewDevAgent(*apiKey)
	if err != nil {
		fmt.Printf("Error initializing agent: %v\n", err)
		os.Exit(1)
	}
	defer agent.Close()

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("🧞 Genie - AI Project Generator (Type 'exit' to quit)")
	fmt.Println("-------------------------------------------")
	fmt.Println("I'm your project genie! Describe what you want to build and I'll make it happen.")
	fmt.Println("Example: 'Create a React dashboard with authentication, dark mode, and real-time charts'")
	fmt.Println("Let's get started!")
	fmt.Println()

	for {
		fmt.Print("Project description: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "exit" {
			break
		}

		if input == "" {
			continue
		}

		// Generate project specification
		spec, err := agent.GenerateProjectSpec(input)
		if err != nil {
			fmt.Printf("Error generating project specification: %v\n", err)
			continue
		}

		// Show specification and ask for confirmation
		specJSON, _ := json.MarshalIndent(spec, "", "  ")
		fmt.Println("\n📋 Project Specification:")
		fmt.Println(string(specJSON))
		fmt.Print("\nProceed with generation? (y/n): ")

		confirm, _ := reader.ReadString('\n')
		confirm = strings.TrimSpace(strings.ToLower(confirm))

		if confirm == "y" {
			err = agent.GenerateCode(spec)
			if err != nil {
				fmt.Printf("Error generating project: %v\n", err)
				continue
			}
		}

		fmt.Println()
	}
}

func removeFirstAndLastLine(input string) string {
	input = strings.ReplaceAll(input, "```", "")
	// Split the string into lines
	lines := strings.Split(input, "\n")

	// Check if there are at least 3 lines to remove the first and last
	if len(lines) <= 2 {
		return "" // Not enough lines to keep any
	}

	// Join the middle lines back into a string
	return strings.Join(lines[1:len(lines)-1], "\n")
}

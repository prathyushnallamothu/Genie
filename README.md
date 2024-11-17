# 🧞 Genie - AI Project Generator

Genie is a powerful CLI tool that uses Google's Gemini AI to generate complete project scaffolding from natural language descriptions. Simply describe the project you want to build, and Genie will create all the necessary files, structure, and boilerplate code for you.

## ✨ Features

- **Natural Language Project Generation**: Describe your project in plain English
- **Smart Project Structure**: Automatically determines the best framework and architecture
- **Complete Scaffolding**: Generates all necessary files, including configuration and documentation
- **Context-Aware**: Ensures consistency across generated files
- **Framework Agnostic**: Supports various project types and frameworks

## 🚀 Getting Started

### Prerequisites

- Go 1.23.2 or higher
- Google API Key for Gemini

### Installation

1. Clone the repository:
```bash
git clone https://github.com/prathyushnallamothu/genie.git
cd genie
```

2. Install dependencies:
```bash
go mod download
```

3. Set up your Google API Key:
```bash
export GOOGLE_API_KEY="your-api-key-here"
```

### Usage

1. Run Genie:
```bash
go run main.go
```

2. When prompted, describe your project. For example:
```
"Create a React dashboard with authentication, dark mode, and real-time charts"
```

3. Review the generated project specification and confirm to proceed
4. Find your generated project in a new directory named after your project

## 🛠️ Example Projects

You can generate various types of projects, such as:
- Web applications (React, Vue, Angular)
- Backend services (Node.js, Go, Python)
- Mobile apps (React Native, Flutter)
- CLI tools
- And more!

## 🔑 API Key Configuration

You can provide your Google API Key in two ways:
1. Environment variable: `GOOGLE_API_KEY`
2. Command line flag: `-api-key`

## 📝 Project Structure

```
genie/
├── main.go          # Main application code
├── go.mod          # Go module definition
├── README.md       # Project documentation
└── generated/      # Directory for generated projects
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## ⚠️ Disclaimer

This tool uses AI to generate code. While it strives for best practices and security, please review generated code before using in production.

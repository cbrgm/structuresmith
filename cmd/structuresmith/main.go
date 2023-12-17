package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/alecthomas/kong"
)

// Global variables for tool metadata.
var (
	Version   string              // Version of the tool.
	Revision  string              // Revision or Commit this binary was built from.
	GoVersion = runtime.Version() // GoVersion running this binary.
	StartTime = time.Now()        // StartTime has the time this was started.
)

// CLI struct defines the command line arguments.
var CLI struct {
	Validate struct {
		GlobalArgs
	} `cmd:"" help:"Validates the YAML configuration"`

	Diff struct {
		DiffArgs
	} `cmd:"" help:"Performs a dry-run to show file paths that would be created"`

	Render struct {
		DiffArgs
	} `cmd:"" help:"Writes the templated files to disk"`
}

// DiffArgs struct for diff related arguments.
type DiffArgs struct {
	GlobalArgs
	Project string `arg:"project" help:"The project in the config to render or diff"`
}

// GlobalArgs struct for global arguments.
type GlobalArgs struct {
	ConfigFile   string `name:"config" help:"Path to the YAML configuration file" type:"path" default:"anvil.yml"`
	OutputPath   string `name:"output" help:"Output path prefix for generated files" type:"path" default:"out"`
	TemplatesDir string `name:"templates" help:"Directory where template files are stored" type:"path" default:"templates"`
}

func main() {
	ctx := kong.Parse(&CLI,
		kong.Name("structuresmith"),
		kong.Description(
			fmt.Sprintf(
				"Automates the generation of project files.\n Version: %s %s\n BuildTime: %s\n %s\n",
				Revision,
				Version,
				StartTime.Format("2006-01-02"),
				GoVersion,
			),
		),
	)
	switch ctx.Command() {
	case "validate":
		executeValidateCommand(CLI.Validate.GlobalArgs)
	case "diff <project>":
		executeDiffCommand(CLI.Diff.DiffArgs)
	case "render <project>":
		executeRenderCommand(CLI.Render.DiffArgs)
	default:
		panic(ctx.Command())
	}
}

// executeValidateCommand handles the 'validate' command.
func executeValidateCommand(args GlobalArgs) {
	app := newStructuresmith(Options{
		ConfigFile:   args.ConfigFile,
		OutputDir:    args.OutputPath,
		TemplatesDir: args.TemplatesDir,
	})

	if _, err := app.loadAndValidateConfig(); err != nil {
		log.Fatalf("Configuration validation error: %v\n", err)
	}
}

// executeDiffCommand handles the 'diff' command.
func executeDiffCommand(args DiffArgs) {
	app := newStructuresmith(Options{
		ConfigFile:   args.ConfigFile,
		OutputDir:    args.OutputPath,
		TemplatesDir: args.TemplatesDir,
	})
	cfg, err := app.loadAndValidateConfig()
	if err != nil {
		log.Fatalf("Configuration validation error: %v\n", err)
	}

	if err := app.diff(args.Project, cfg); err != nil {
		log.Fatalf("Configuration diff error: %v\n", err)
	}
}

// executeRenderCommand handles the 'render' command.
func executeRenderCommand(args DiffArgs) {
	app := newStructuresmith(Options{
		ConfigFile:   args.ConfigFile,
		OutputDir:    args.OutputPath,
		TemplatesDir: args.TemplatesDir,
	})

	cfg, err := app.loadAndValidateConfig()
	if err != nil {
		log.Fatalf("Configuration validation error: %v\n", err)
	}

	if err := app.render(args.Project, cfg); err != nil {
		log.Fatalf("Configuration diff error: %v\n", err)
	}
}

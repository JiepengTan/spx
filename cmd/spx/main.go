package main

import (
	"embed"
	"fmt"
	"godot-ext/gdspx/cmd/gdspx/pkg/impl"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	_ "embed"
)

var (
	//go:embed template/engine/*
	engineFiles embed.FS

	//go:embed template/go.mod.txt
	go_mode_txt string

	//go:embed template/gitignore.txt
	gitignore string

	//go:embed template/main.go
	main_go string
)

func main() {
	impl.ReplaceTemplate(go_mode_txt, main_go, gitignore)
	impl.CheckPresetEnvironment()
	impl.TargetDir = "."
	if len(os.Args) > 2 {
		impl.TargetDir = os.Args[2]
	}
	if len(os.Args) <= 1 {
		showHelpInfo()
		return
	}
	switch os.Args[1] {
	case "help", "version":
		showHelpInfo()
		return
	case "clear":
		if impl.IsFileExist(impl.TargetDir + "/.godot") {
			clearProject(impl.TargetDir)
		} else {
			fmt.Println("Not a spx project skip")
		}
		return
	case "stopweb":
		impl.StopWebServer()
		return
	case "init":
		impl.PrepareGoEnv()
	}

	if !impl.IsFileExist(impl.TargetDir + "/go.mod") {
		impl.PrepareGoEnv()
	}

	if err := wrap(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func wrap() error {
	CopyEmbed(impl.TargetDir)
	// look for a go.mod file
	gd4spxPath, project, libPath, err := impl.SetupEnv()
	if err != nil {
		return err
	}

	switch os.Args[1] {
	case "init":
		return nil
	case "run", "editor", "export", "build":
		BuildDll(project, libPath)
	case "runweb", "buildweb", "exportweb":
		impl.BuildWasm(project)
	}

	switch os.Args[1] {
	case "run":
		return impl.RunGdspx(gd4spxPath, project, "")
	case "editor":
		return impl.RunGdspx(gd4spxPath, project, "-e")
	case "runweb":
		return impl.RunWebServer(gd4spxPath, project, 8005)
	case "exportweb":
		return impl.ExportWeb(gd4spxPath, project)
	case "export":
		return impl.Export(gd4spxPath, project)
	}
	return nil
}
func clearProject(dir string) {
	deleteFilesAndDirs(dir)
	deleteImportFiles(dir)
}
func deleteFilesAndDirs(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		fullPath := filepath.Join(dir, file.Name())
		if file.Name() == "assets" || strings.HasSuffix(fullPath, ".spx") {
			continue
		}

		if file.IsDir() {
			err = os.RemoveAll(fullPath)
			if err != nil {
				return err
			}
		} else {
			err = os.Remove(fullPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
func deleteImportFiles(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".import") {
			err = os.Remove(path)
			if err != nil {
				return fmt.Errorf("failed to delete file: %v", err)
			}
		}

		return nil
	})
}
func showHelpInfo() {
	fmt.Println(`
Usage:

    spx <command> [path]      

The commands are:

    - init            # Create a spx project in the current directory
    - run             # Run the current project
    - editor          # Open the current project in editor mode
    - build           # Build the dynamic library
    - export          # Export the PC package (macOS, Windows, Linux) (TODO)
    - runweb          # Launch the web server
    - buildweb        # Build for WebAssembly (WASM)
    - exportweb       # Export the web package
    - clear           # Clear the project 

 eg:

    spx init                      # create a project in current path
    spx init ./test/demo01        # create a project at path ./test/demo01 
	`)
}

func CopyEmbed(dst string) error {
	enginePath := filepath.Join(dst, "engine")
	if _, err := os.Stat(enginePath); !os.IsNotExist(err) {
		err := os.RemoveAll(enginePath)
		if err != nil {
			return err
		}
	}

	fsys, err := fs.Sub(engineFiles, "template/engine")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		dstPath := filepath.Join(enginePath, path)
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		} else {
			srcFile, err := fsys.Open(path)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			dstFile, err := os.Create(dstPath)
			if err != nil {
				return err
			}
			defer dstFile.Close()

			_, err = io.Copy(dstFile, srcFile)
			return err
		}
	})
}

func BuildDll(project, outputPath string) {
	os.Remove(path.Join(project, "main.go"))
	rawdir, _ := os.Getwd()
	os.Chdir(project)
	envVars := []string{""}
	RunGoplus(envVars, "build")
	os.Chdir(rawdir)
	os.Rename(path.Join(project, "gop_autogen.go"), path.Join(project, "main.go"))
	os.Remove(path.Join(project, "gdspx-demo.exe"))
	impl.BuildDll(project, outputPath)
}

func RunGoplus(envVars []string, args ...string) error {
	golang := exec.Command("gop", args...)

	if envVars != nil {
		golang.Env = append(os.Environ(), envVars...)
	}
	golang.Stderr = os.Stderr
	golang.Stdout = os.Stdout
	golang.Stdin = os.Stdin
	return golang.Run()
}

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

	_ "embed"
)

var (
	//go:embed template/engine/*
	engineFiles embed.FS

	//go:embed template/go.mod.txt
	go_mode_txt string
)

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

func main() {
	impl.CheckPresetEnvironment()
	impl.TargetDir = "."
	if len(os.Args) > 2 {
		impl.TargetDir = os.Args[2]
	}
	if len(os.Args) <= 1 {
		impl.ShowHelpInfo()
		return
	}
	switch os.Args[1] {
	case "help", "version":
		impl.ShowHelpInfo()
		return
	case "init":
		impl.PrepareGoEnv(go_mode_txt)
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
	case "buildweb", "exportweb":
		impl.BuildWasm(project)
	}

	switch os.Args[1] {
	case "run":
		return impl.RunGdspx(gd4spxPath, project, "")
	case "editor":
		return impl.RunGdspx(gd4spxPath, project, "-e")
	}
	return nil
}
func BuildDll(project, outputPath string) {
	os.Remove(path.Join(project, "main.go"))
	rawdir, _ := os.Getwd()
	os.Chdir(project)
	envVars := []string{""}
	RunGoplus(envVars, "build")
	os.Chdir(rawdir)
	os.Rename(path.Join(project, "gop_autogen.go"), path.Join(project, "main.go"))
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
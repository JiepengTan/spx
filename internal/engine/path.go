package engine

import (
	"encoding/json"
	"path"
	"path/filepath"
)

var (
	extassetDir = ""
	assetsDir   = "res://assets/"
)

const (
	configPath         = "res://.config"
	engineExtAssetPath = "extasset"
)

type engineConfig struct {
	ExtAsset string `json:"extasset"`
}

func SetAssetDir(dir string) {
	// load config
	if SyncResHasFile(configPath) {
		configJson := SyncResReadAllText(configPath)
		var config engineConfig
		json.Unmarshal([]byte(configJson), config)
		extassetDir = config.ExtAsset
	}
	assetsDir = "res://" + dir + "/"
	println("=========SetAssetDir:=======", assetsDir, "extasset", extassetDir)
}

func ToAssetPath(relPath string) string {
	replacedPath := replacePathIfInExtAssetDir(relPath, extassetDir, engineExtAssetPath)
	if replacedPath != "" {
		return replacedPath
	}
	return path.Join(assetsDir, relPath)
}

func replacePathIfInExtAssetDir(path string, extassetDir string, newAssetDir string) string {
	if extassetDir == "" {
		return ""
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		println("path error", path, err.Error())
		return ""
	}

	absExtassetDir, err := filepath.Abs(extassetDir)
	if err != nil {
		println("path error", path, err.Error())
		return ""
	}

	relPath, err := filepath.Rel(absExtassetDir, absPath)
	if err != nil {
		println("path error", path, err.Error())
		return ""
	}

	if filepath.IsAbs(relPath) || relPath == ".." {
		println("path error", path, err.Error())
		return path
	}

	newPath := filepath.Join("res://", newAssetDir, relPath)
	return newPath
}

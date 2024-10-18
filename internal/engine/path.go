package engine

import (
	"encoding/json"
	"path/filepath"
	"strings"
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
		json.Unmarshal([]byte(configJson), &config)
		extassetDir = config.ExtAsset
	}
	assetsDir = "res://" + dir + "/"
}

func ToAssetPath(relPath string) string {
	replacedPath := replacePathIfInExtAssetDir(relPath, extassetDir, engineExtAssetPath)
	if replacedPath != "" {
		return replacedPath
	}
	return assetsDir + relPath
}

func replacePathIfInExtAssetDir(path string, extassetDir string, newAssetDir string) string {
	if extassetDir == "" {
		return ""
	}
	prefix := "../" + extassetDir
	if strings.HasPrefix(path, prefix) || strings.HasPrefix(path, extassetDir) {
		newPath := "res://" + filepath.Join(newAssetDir, path[len(prefix)+1:])
		newPath = strings.ReplaceAll(newPath, "\\", "/")
		return newPath
	}
	return ""
}

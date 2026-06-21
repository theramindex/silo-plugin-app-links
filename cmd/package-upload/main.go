package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type manifestFile struct {
	PluginID           string              `json:"plugin_id"`
	Version            string              `json:"version"`
	Checksum           string              `json:"checksum"`
	SiloAPIVersion     string              `json:"silo_api_version,omitempty"`
	SupportedPlatforms []map[string]string `json:"supported_platforms,omitempty"`
	Capabilities       []any               `json:"capabilities,omitempty"`
	HTTPRoutes         []any               `json:"http_routes,omitempty"`
	Metadata           map[string]any      `json:"metadata,omitempty"`
}

func main() {
	binaryPath := flag.String("binary", "", "binary path")
	version := flag.String("version", "", "plugin version")
	goos := flag.String("goos", "", "target os")
	goarch := flag.String("goarch", "", "target arch")
	pluginID := flag.String("plugin-id", "", "plugin id")
	flag.Parse()

	if *binaryPath == "" || *version == "" || *goos == "" || *goarch == "" || *pluginID == "" {
		panic("binary, version, goos, goarch, and plugin-id are required")
	}
	data, err := os.ReadFile(*binaryPath)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(data)
	checksum := hex.EncodeToString(sum[:])

	manifestData, err := os.ReadFile("manifest.json")
	if err != nil {
		panic(err)
	}
	var manifest manifestFile
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		panic(err)
	}
	manifest.PluginID = *pluginID
	manifest.Version = *version
	manifest.Checksum = checksum
	manifest.SupportedPlatforms = []map[string]string{{"os": *goos, "arch": *goarch}}

	outManifest := strings.TrimSuffix(*binaryPath, filepath.Ext(*binaryPath)) + ".manifest.json"
	encodedManifest, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(outManifest, append(encodedManifest, '\n'), 0o644); err != nil {
		panic(err)
	}
	if err := os.WriteFile(*binaryPath+".sha256", []byte(fmt.Sprintf("%s  %s\n", checksum, filepath.Base(*binaryPath))), 0o644); err != nil {
		panic(err)
	}
	zipPath := *binaryPath + ".silo-plugin.zip"
	if err := writeZip(zipPath, *binaryPath, outManifest); err != nil {
		panic(err)
	}
}

func writeZip(zipPath, binaryPath, manifestPath string) error {
	out, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer out.Close()
	writer := zip.NewWriter(out)
	defer writer.Close()
	for _, path := range []string{binaryPath, manifestPath} {
		if err := addFile(writer, path); err != nil {
			return err
		}
	}
	return nil
}

func addFile(writer *zip.Writer, path string) error {
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = filepath.Base(path)
	if info.Mode()&0o111 != 0 {
		header.SetMode(0o755)
	}
	fileWriter, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(fileWriter, in)
	return err
}

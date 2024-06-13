package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"sigs.k8s.io/kustomize/api/krusty"
	kustypes "sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// getSubDirectories returns a slice containing the names of all immediate subdirectories in the given directory
func getSubDirectories(dir string) ([]string, error) {
	subDirs := make([]string, 0)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return subDirs, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			subDirs = append(subDirs, filepath.Join(dir, entry.Name()))
		}
	}
	return subDirs, nil
}

// findFoldersWithPattern finds all folders matching the pattern in the given directory
func findFoldersWithPattern(rootDir string, pattern string) ([]string, error) {
	folders := make([]string, 0)
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Check if the path is a directory and matches the pattern
		if info.IsDir() && matchesPattern(path, pattern) {
			folders = append(folders, path)
		}
		return nil
	})
	if err != nil {
		return folders, err
	}
	return folders, nil
}

// matchesPattern checks if the given path matches the specified pattern
func matchesPattern(path, pattern string) bool {
	matched, err := filepath.Match(pattern, filepath.Base(path))
	return err == nil && matched
}

// validateOverlaysFolders checks if the overlay folders are able to run kustomize build
func validateOverlaysFolders(fs filesys.FileSystem, kustomizationDir string) error {
	buildOptions := &krusty.Options{
		LoadRestrictions: kustypes.LoadRestrictionsNone,
		PluginConfig:     kustypes.DisabledPluginConfig(),
	}
	k := krusty.MakeKustomizer(buildOptions)
	m, err := k.Run(fs, kustomizationDir)
	if err != nil {
		return fmt.Errorf("error with kustomizer.Run: %v", err)
	}
	_, err = m.AsYaml()
	if err != nil {
		return fmt.Errorf("error with coverting kustomization output to yaml: %v", err)
	}
	return nil
}

type KustomizationWorkloads struct {
	ParentFolder string
	ChildFolder  string
	Err          error
}

func main() {
	rootDir := "workloads"
	pattern := "overlays"
	fs := filesys.MakeFsOnDisk()
	folders, err := findFoldersWithPattern(rootDir, pattern)
	if err != nil {
		errMsg := fmt.Errorf("error finding folders: %v", err)
		slog.Error(errMsg.Error())
		panic(errMsg)
	}
	kustomizationIssues := make([]KustomizationWorkloads, 0)

	for _, folder := range folders {
		parentFolder := filepath.Dir(folder)
		slog.Info(fmt.Sprintf("Folder: %s Parent: %s\n", folder, parentFolder))
		childFolders, _ := getSubDirectories(folder)
		for _, childFolder := range childFolders {
			err := validateOverlaysFolders(fs, childFolder)
			if err != nil {
				kustomizationIssues = append(kustomizationIssues, KustomizationWorkloads{
					ParentFolder: parentFolder,
					ChildFolder:  childFolder,
					Err:          err,
				})
			}

		}
	}

	for i := 0; i < len(kustomizationIssues); i++ {
		log.Printf(`
		ChildFolder: %s
		ErrorMsg: %v
		`, kustomizationIssues[i].ChildFolder, kustomizationIssues[i].Err,
		)
	}

	if len(kustomizationIssues) == 0 {
		log.Println("No kustomization issues found")
	}
}

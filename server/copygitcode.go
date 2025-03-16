package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func CpyCode(url string) error {

	storageDir := "code-storage"

	if err := os.MkdirAll(storageDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create storage directory: %v", err)
	}

	repoName := filepath.Base(url)
	if len(repoName) > 4 && repoName[len(repoName)-4:] == ".git" {
		repoName = repoName[:len(repoName)-4]
	}

	targetPath := filepath.Join(storageDir, repoName)

	cmd := exec.Command("git", "clone", url, targetPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Cloning repository:", url, "into", targetPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repo: %v", err)
	}

	fmt.Println("Repository cloned successfully into", targetPath)
	return nil
}

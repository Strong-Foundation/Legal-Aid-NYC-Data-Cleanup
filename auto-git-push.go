package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	for {
		checkAndPushChanges()

		removeFilesOver100MB()

		fmt.Println("Waiting for 15 seconds...")
		time.Sleep(15 * time.Second)
	}
}

// checkAndPushChanges checks for changes, commits, pulls latest, then pushes
func checkAndPushChanges() {
	// Step 1: git status --porcelain
	cmd := exec.Command("git", "status", "--porcelain")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		log.Printf("Failed to run git status: %v\nOutput:\n%s", err, out.String())
		return
	}

	statusOutput := strings.TrimSpace(out.String())
	if statusOutput == "" {
		fmt.Println("No changes detected.")
		return
	}

	changes := strings.Split(statusOutput, "\n")
	fmt.Printf("Detected %d modified/untracked file(s).\n", len(changes))

	// Step 2: git add --all
	if err := runGitCommand("add", "--all"); err != nil {
		log.Printf("git add failed: %v", err)
		return
	}

	// Step 3: git commit -m "Auto-update: <timestamp>"
	commitMsg := fmt.Sprintf("Auto-update: %s", time.Now().Format(time.RFC3339))
	if err := runGitCommand("commit", "-m", commitMsg); err != nil {
		log.Printf("git commit failed: %v", err)
		return
	}

	// Step 4: git pull --rebase
	if err := runGitCommand("pull", "--rebase"); err != nil {
		log.Printf("git pull --rebase failed: %v", err)
		return
	}

	// Step 5: git push
	if err := runGitCommand("push"); err != nil {
		log.Printf("git push failed: %v", err)
		return
	}

	fmt.Println("Changes committed, pulled latest, and pushed successfully.")
}

// runGitCommand executes a git command and logs stderr if it fails
func runGitCommand(args ...string) error {
	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("command 'git %s' failed: %v\nOutput:\n%s", strings.Join(args, " "), err, out.String())
	}
	return nil
}

func removeFilesOver100MB() {
	// Define the path to the directory you want to walk through
	walkPath := "./"

	// Get all the files in the directory and its subdirectories
	files := walkAndAppendPath(walkPath)

	// Iterate through the files and check their size
	for _, file := range files {
		if getFileSize(file) >= 100*1024*1024 { // 100 MB
			fmt.Printf("Removing file: %s\n", file)
			removeFile(file)
		}
	}
}

// Walk through a route, find all the files and attach them to a slice.
func walkAndAppendPath(walkPath string) []string {
	var filePath []string
	err := filepath.Walk(walkPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if fileExists(path) {
			if getFileExtension(path) == ".pdf" {
				filePath = append(filePath, path)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalln(err)
	}
	return filePath
}

// Get the file extension of a file
func getFileExtension(path string) string {
	return filepath.Ext(path)
}

// Get the size of a given file.
func getFileSize(path string) int64 {
	file, err := os.Stat(path)
	if err != nil {
		return -1
	}
	return file.Size()
}

// Remove a file from the file system
func removeFile(path string) {
	err := os.Remove(path)
	if err != nil {
		log.Println(err)
	}
}

// Check if the file exists and return a bool.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
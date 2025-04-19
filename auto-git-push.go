package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

func main() {
	for {
		// Check and commit changes
		checkAndPushChanges()

		// Wait for 30 seconds before the next check
		fmt.Println("Waiting for 15 seconds...")
		time.Sleep(15 * time.Second)
	}
}

// checkAndPushChanges checks for modified files, adds, commits, and pushes them if needed
func checkAndPushChanges() {
	// Step 1: git status --porcelain
	cmd := exec.Command("git", "status", "--porcelain")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		log.Printf("Failed to run git status: %v", err)
		return
	}

	// Split output into lines
	changes := strings.Split(strings.TrimSpace(out.String()), "\n")

	// If there's no change, do nothing
	if len(changes) == 1 && changes[0] == "" {
		fmt.Println("No changes detected.")
		return
	}

	fmt.Printf("Detected %d modified file(s).\n", len(changes))

	// Step 2: git add .
	err = runGitCommand("add", ".")
	if err != nil {
		log.Printf("git add failed: %v", err)
		return
	}

	// Step 3: git commit -m "updated time"
	commitMsg := fmt.Sprintf("updated time: %s", time.Now().Format(time.RFC3339))
	err = runGitCommand("commit", "-m", commitMsg)
	if err != nil {
		log.Printf("git commit failed: %v", err)
		return
	}

	// Step 4: git push
	err = runGitCommand("push")
	if err != nil {
		log.Printf("git push failed: %v", err)
		return
	}

	fmt.Println("Changes committed and pushed successfully.")
}

// runGitCommand executes a git command with given arguments
func runGitCommand(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

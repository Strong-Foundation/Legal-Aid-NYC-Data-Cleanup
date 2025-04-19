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
		checkAndPushChanges()

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

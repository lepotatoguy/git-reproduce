package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"flag"
)

func main() {
	repo := flag.String("repo", "", "Git repository URL")
	branch := flag.String("branch", "main", "Branch to checkout")
	commit := flag.String("commit", "", "Commit hash to checkout")
	flag.Parse()

	if *repo == "" || *commit == "" {
		fmt.Println("Usage: git-reproduce --repo <repo-url> --branch <branch> --commit <sha>")
		os.Exit(1)
	}

	repoName := extractRepoName(*repo)

	// Clone the repo
	fmt.Println("Cloning repository...")
	err := runCommand("git", "clone", "--branch", *branch, *repo)
	if err != nil {
		log.Fatalf("Failed to clone repo: %v", err)
	}

	// Change directory
	if err := os.Chdir(repoName); err != nil {
		log.Fatalf("Failed to enter repo dir: %v", err)
	}

	// Checkout commit
	fmt.Println("Checking out commit...")
	err = runCommand("git", "checkout", *commit)
	if err != nil {
		log.Fatalf("Failed to checkout commit: %v", err)
	}

	fmt.Println("Done.")
}

func extractRepoName(repoURL string) string {
	lastSlash := len(repoURL) - 1
	for i := len(repoURL) - 1; i >= 0; i-- {
		if repoURL[i] == '/' {
			lastSlash = i
			break
		}
	}
	repoName := repoURL[lastSlash+1:]
	if len(repoName) > 4 && repoName[len(repoName)-4:] == ".git" {
		repoName = repoName[:len(repoName)-4]
	}
	return repoName
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// Define CLI flags with short and long versions
	repo := flag.String("repo", "", "Git repository URL")
	r := flag.String("r", "", "Git repository URL (shorthand)")

	branch := flag.String("branch", "", "Branch to checkout")
	b := flag.String("b", "", "Branch to checkout (shorthand)")

	commit := flag.String("commit", "", "Commit hash to checkout (optional)")
	c := flag.String("c", "", "Commit hash to checkout (shorthand)")

	// Override help output with custom formatting
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, `Usage:
  git-reproduce --repo <url> --branch <branch> [--commit <sha>]
  git-reproduce -r <url> -b <branch> [-c <sha>]

Description:
  Clone a Git repository at a specific branch and optionally checkout a specific commit.
  If --commit/-c is not provided, the latest commit of the branch will be used.

Flags:
  -r or --repo string
        Git repository URL

  -b or --branch string
        Branch to checkout

  -c or --commit string
        Commit hash to checkout`)
	}

	flag.Parse()

	// Show help if no arguments
	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(1)
	}

	// Merge short and long flag values
	finalRepo := chooseValue(*repo, *r)
	finalBranch := chooseValue(*branch, *b)
	finalCommit := chooseValue(*commit, *c)

	if finalRepo == "" || finalBranch == "" {
		fmt.Fprintln(os.Stderr, "Error: Both --repo/-r and --branch/-b are required.\n")
		flag.Usage()
		os.Exit(1)
	}

	// Check for Git availability
	if !isGitAvailable() {
		fmt.Fprintln(os.Stderr, "Error: 'git' is not installed. Please install Git and try again.")
		os.Exit(1)
	}

	repoName := extractRepoName(finalRepo)

	// Clone repository
	fmt.Println("Cloning repository...")
	err := runCommand("git", "clone", "--branch", finalBranch, "--single-branch", finalRepo)
	if err != nil {
		log.Fatalf("Failed to clone repository: %v", err)
	}

	// Enter cloned repository
	if err := os.Chdir(repoName); err != nil {
		log.Fatalf("Failed to enter cloned repo directory: %v", err)
	}

	// Determine commit
	targetCommit := finalCommit
	if targetCommit == "" {
		targetCommit, err = getLatestCommitSHA(finalBranch)
		if err != nil {
			log.Fatalf("Failed to get latest commit: %v", err)
		}
		fmt.Printf("Using latest commit on branch '%s': %s\n", finalBranch, targetCommit)
	}

	// Checkout commit
	fmt.Printf("Checking out commit %s...\n", targetCommit)
	err = runCommand("git", "checkout", targetCommit)
	if err != nil {
		log.Fatalf("Failed to checkout commit: %v", err)
	}

	fmt.Println("Done.")
}

// Selects the long or short value, preferring the long form
func chooseValue(long, short string) string {
	if long != "" {
		return long
	}
	return short
}

// Checks if Git is available in PATH
func isGitAvailable() bool {
	cmd := exec.Command("git", "--version")
	err := cmd.Run()
	return err == nil
}

// Executes a system command and streams its output
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Extracts the repo name from the URL, stripping `.git` if present
func extractRepoName(repoURL string) string {
	parts := strings.Split(strings.TrimSuffix(repoURL, ".git"), "/")
	return parts[len(parts)-1]
}

// Returns the latest commit SHA of the given branch
func getLatestCommitSHA(branch string) (string, error) {
	cmd := exec.Command("git", "rev-parse", branch)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

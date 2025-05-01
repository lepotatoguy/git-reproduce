package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Snapshot structure for YAML serialization/deserialization
type Snapshot struct {
	Repository    string     `yaml:"repository"`
	Branch        string     `yaml:"branch"`
	Author        string     `yaml:"author"`
	Commit        string     `yaml:"commit"`
	CommitMessage string     `yaml:"commit_message"`
	DateOfCommit  string     `yaml:"date_of_commit"`
	Timestamp     string     `yaml:"timestamp"`
	GitVersion    string     `yaml:"git_version"`
	System        SystemInfo `yaml:"system"`
	Notes         string     `yaml:"notes,omitempty"`
}

// SystemInfo holds OS, architecture, and hostname
type SystemInfo struct {
	OS           string `yaml:"os"`
	Architecture string `yaml:"architecture"`
	Hostname     string `yaml:"hostname"`
}

var verbose bool

func main() {
	// Define CLI flags
	repo := flag.String("repo", "", "Git repository URL")
	r := flag.String("r", "", "Git repository URL (shorthand)")

	branch := flag.String("branch", "", "Branch to checkout")
	b := flag.String("b", "", "Branch to checkout (shorthand)")

	commit := flag.String("commit", "", "Commit hash to checkout (optional)")
	c := flag.String("c", "", "Commit hash to checkout (shorthand)")

	snapshotMode := flag.Bool("snapshot", false, "Create a reproducibility snapshot interactively")
	fromSnapshot := flag.String("from-snapshot", "", "Execute clone and checkout from snapshot YAML file")

	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&verbose, "v", false, "Enable verbose logging (shorthand)")

	// Override help output
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, `Usage:
  git-reproduce --repo <url> --branch <branch> [--commit <sha>]
  git-reproduce -r <url> -b <branch> [-c <sha>]
  git-reproduce --snapshot
  git-reproduce --from-snapshot <file>

Description:
  Clone a Git repository at a specific branch and optionally checkout a specific commit.
  If --commit/-c is not provided, the latest commit of the branch will be used.
  Use --snapshot to create a reproducibility snapshot interactively.
  Use --from-snapshot <file> to reproduce from an existing snapshot YAML.

Flags:
  -r or --repo string
        Git repository URL
  -b or --branch string
        Branch to checkout
  -c or --commit string
        Commit hash to checkout
  --snapshot
        Enter interactive snapshot mode
  --from-snapshot <file>
        Execute clone and checkout from snapshot YAML file
  -v or --verbose
        Enable verbose logging`)
	}

	flag.Parse()

	// Handle --snapshot mode
	if *snapshotMode {
		handleSnapshot()
		return
	}

	// Handle --from-snapshot mode
	if *fromSnapshot != "" {
		handleFromSnapshot(*fromSnapshot)
		return
	}

	// If no flags provided, show help
	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(1)
	}

	// Resolve long/short flag values
	finalRepo := chooseValue(*repo, *r)
	finalBranch := chooseValue(*branch, *b)
	finalCommit := chooseValue(*commit, *c)

	// Validate required flags
	if finalRepo == "" || finalBranch == "" {
		fmt.Fprintln(os.Stderr, "Error: Both --repo/-r and --branch/-b are required.")
		flag.Usage()
		os.Exit(1)
	}

	// Ensure git is installed
	if !isGitAvailable() {
		fmt.Fprintln(os.Stderr, "Error: 'git' is not installed. Please install Git and try again.")
		os.Exit(1)
	}

	// Clone repository
	repoName := extractRepoName(finalRepo)
	fmt.Println("Cloning repository...")
	if err := runCommand("git", "clone", "--branch", finalBranch, "--single-branch", finalRepo); err != nil {
		log.Fatalf("Failed to clone repository: %v", err)
	}

	// Change to cloned directory
	if err := os.Chdir(repoName); err != nil {
		log.Fatalf("Failed to enter cloned repo directory: %v", err)
	}

	// Determine commit
	targetCommit := finalCommit
	if targetCommit == "" {
		var err error
		targetCommit, err = getLatestCommitSHA(finalBranch)
		if err != nil {
			log.Fatalf("Failed to get latest commit: %v", err)
		}
		fmt.Printf("Using latest commit on branch '%s': %s\n", finalBranch, targetCommit)
	}

	// Checkout commit
	fmt.Printf("Checking out commit %s...\n", targetCommit)
	if err := runCommand("git", "checkout", targetCommit); err != nil {
		log.Fatalf("Failed to checkout commit: %v", err)
	}

	fmt.Println("Done.")
}

// handleSnapshot runs the interactive snapshot creation process
func handleSnapshot() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter repository URL: ")
	repoURL, _ := reader.ReadString('\n')
	repoURL = strings.TrimSpace(repoURL)

	fmt.Print("Enter branch name: ")
	branch, _ := reader.ReadString('\n')
	branch = strings.TrimSpace(branch)

	fmt.Print("Enter commit hash (leave blank for latest): ")
	commit, _ := reader.ReadString('\n')
	commit = strings.TrimSpace(commit)

	fmt.Print("Enter snapshot file name (without extension): ")
	fileName, _ := reader.ReadString('\n')
	fileName = strings.TrimSpace(fileName)
	if !strings.HasSuffix(fileName, ".yaml") {
		fileName += ".yaml"
	}

	fmt.Print("Enter optional notes (leave blank if none): ")
	notes, _ := reader.ReadString('\n')
	notes = strings.TrimSpace(notes)

	logf("Cloning repository...")
	if err := runCommand("git", "clone", "--branch", branch, "--single-branch", repoURL); err != nil {
		fmt.Fprintln(os.Stderr, "Error: Failed to clone repository.")
		fmt.Fprintln(os.Stderr, "Suggestion: Verify that the repository URL and branch are correct and accessible.")
		os.Exit(1)
	}

	repoName := extractRepoName(repoURL)
	if err := os.Chdir(repoName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to enter cloned directory %s\n", repoName)
		os.Exit(1)
	}

	if commit == "" {
		latestCommit, err := getLatestCommitSHA(branch)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: Failed to retrieve latest commit SHA.")
			os.Exit(1)
		}
		fmt.Printf("Latest commit on branch '%s' is %s\n", branch, latestCommit)
		fmt.Print("Use this commit? [Y/n]: ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer == "n" || answer == "no" {
			fmt.Println("Aborting snapshot.")
			os.Exit(0)
		}
		commit = latestCommit
	}

	logf("Checking out commit %s...", commit)
	if err := runCommand("git", "checkout", commit); err != nil {
		fmt.Fprintln(os.Stderr, "Error: Failed to checkout commit.")
		os.Exit(1)
	}

	// Gather commit metadata
	author, message, commitDate, err := getCommitMetadata(commit)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: Failed to retrieve commit metadata.")
		os.Exit(1)
	}

	// Gather system info
	gitVersion, _ := getCommandOutput("git", "--version")
	sysInfo := getSystemInfo()

	// Build snapshot struct
	snapshot := Snapshot{
		Repository:    repoURL,
		Branch:        branch,
		Author:        author,
		Commit:        commit,
		CommitMessage: message,
		DateOfCommit:  commitDate,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		GitVersion:    strings.TrimSpace(gitVersion),
		System:        sysInfo,
	}

	if notes != "" {
		snapshot.Notes = notes
	}

	// Write YAML snapshot file
	data, err := yaml.Marshal(&snapshot)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: Failed to serialize snapshot YAML.")
		os.Exit(1)
	}

	if err := os.WriteFile(fileName, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to write snapshot file %s\n", fileName)
		os.Exit(1)
	}

	fmt.Printf("Snapshot written to %s\n", fileName)
}

// handleFromSnapshot runs the reproducibility process from YAML snapshot file
func handleFromSnapshot(file string) {
	// Read YAML snapshot file
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to read snapshot file %s\n", file)
		os.Exit(1)
	}

	var snap Snapshot
	if err := yaml.Unmarshal(data, &snap); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to parse YAML snapshot %s\n", file)
		os.Exit(1)
	}

	if !isGitAvailable() {
		fmt.Fprintln(os.Stderr, "Error: 'git' is not installed. Please install Git and try again.")
		os.Exit(1)
	}

	logf("Cloning repository from snapshot: %s", snap.Repository)
	if err := runCommand("git", "clone", "--branch", snap.Branch, "--single-branch", snap.Repository); err != nil {
		fmt.Fprintln(os.Stderr, "Error: Failed to clone repository from snapshot.")
		os.Exit(1)
	}

	repoName := extractRepoName(snap.Repository)
	if err := os.Chdir(repoName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to enter cloned directory %s\n", repoName)
		os.Exit(1)
	}

	logf("Checking out commit %s from snapshot", snap.Commit)
	if err := runCommand("git", "checkout", snap.Commit); err != nil {
		fmt.Fprintln(os.Stderr, "Error: Failed to checkout commit from snapshot.")
		os.Exit(1)
	}

	// Print summary
	fmt.Println("\nRepository successfully cloned and commit checked out from snapshot.")
	fmt.Printf("Repository: %s\n", snap.Repository)
	fmt.Printf("Branch: %s\n", snap.Branch)
	fmt.Printf("Commit: %s\n", snap.Commit)
	fmt.Printf("Author: %s\n", snap.Author)
	fmt.Printf("Commit Message: %s\n", snap.CommitMessage)
	fmt.Printf("Commit Date: %s\n", snap.DateOfCommit)
	if snap.Notes != "" {
		fmt.Printf("Notes: %s\n", snap.Notes)
	}
}

// chooseValue picks the long-form or short-form flag value
func chooseValue(long, short string) string {
	if long != "" {
		return long
	}
	return short
}

// logf prints logs if verbose is enabled
func logf(format string, args ...interface{}) {
	if verbose {
		log.Printf(format, args...)
	}
}

// runCommand runs a shell command with optional verbose logging
func runCommand(name string, args ...string) error {
	if verbose {
		log.Printf("Executing: %s %s", name, strings.Join(args, " "))
	}
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// getCommandOutput returns output of a command
func getCommandOutput(name string, args ...string) (string, error) {
	if verbose {
		log.Printf("Executing: %s %s", name, strings.Join(args, " "))
	}
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	return string(out), err
}

// getLatestCommitSHA returns latest commit hash for a branch
func getLatestCommitSHA(branch string) (string, error) {
	return getCommandOutput("git", "rev-parse", branch)
}

// getCommitMetadata returns author, message, date for a commit
func getCommitMetadata(commit string) (author, message, date string, err error) {
	a, err := getCommandOutput("git", "show", "-s", "--format=%an <%ae>", commit)
	if err != nil {
		return "", "", "", err
	}
	m, err := getCommandOutput("git", "show", "-s", "--format=%s", commit)
	if err != nil {
		return "", "", "", err
	}
	d, err := getCommandOutput("git", "show", "-s", "--format=%cI", commit)
	if err != nil {
		return "", "", "", err
	}
	return strings.TrimSpace(a), strings.TrimSpace(m), strings.TrimSpace(d), nil
}

// extractRepoName extracts repo name from a URL
func extractRepoName(repoURL string) string {
	parts := strings.Split(strings.TrimSuffix(repoURL, ".git"), "/")
	return parts[len(parts)-1]
}

// getSystemInfo gathers basic system info
func getSystemInfo() SystemInfo {
	host, _ := os.Hostname()
	arch := getEnvOrDefault("HOSTTYPE", "unknown")
	osName := getEnvOrDefault("OSTYPE", "unknown")
	return SystemInfo{
		OS:           osName,
		Architecture: arch,
		Hostname:     host,
	}
}

// getEnvOrDefault returns env var or fallback
func getEnvOrDefault(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

// isGitAvailable checks if git is installed
func isGitAvailable() bool {
	cmd := exec.Command("git", "--version")
	err := cmd.Run()
	return err == nil
}

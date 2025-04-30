# git-reproduce

`git-reproduce` is a lightweight, system-level command-line tool that ensures reproducible Git checkouts.  
It allows you to clone a Git repository and optionally checkout a specific commit — all in one command.

> Think of it as `git clone` + `git checkout <sha>` in a single, reproducible, and platform-consistent step.


---

## Why This Tool?

### For Researchers and Engineers Focused on Reproducibility

Scientific computing, machine learning experiments, and long-term software engineering all require **reproducible environments**.  
However, Git's default behavior clones only the latest commit of a branch — making it difficult to trace or recreate historical runs.

**`git-reproduce` solves this** by making it easy to:

- Clone a repo and pin it to a specific commit **automatically**
- Track and share experiment code states
- Reduce human error when setting up reproducible pipelines

This tool is especially useful for:
- Research groups publishing reproducible code
- Data scientists archiving experiment environments
- CI/CD pipelines that must replicate code snapshots
- Anyone who needs Git-level reproducibility with minimal effort


---

## Features

- ✅ Clone a Git repository at a specified branch
- ✅ Optionally checkout a specific commit (if omitted, uses the latest commit on that branch)
- ✅ Works across **Linux**, **macOS**, and **Windows**
- ✅ No dependencies except `git`

---

## Installation

### Linux (Debian/Ubuntu)

Coming soon: `.deb` packages and APT repo

### macOS (via Homebrew)

Coming soon.  

---

## Usage

```
$ git-reproduce --help

Usage:
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
        Commit hash to checkout

```
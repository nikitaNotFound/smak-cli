package internal

import (
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Branch struct {
	Name              string
	LastCommitDate    time.Time
	LastCommitMessage string
	CommitsAhead      int
	CommitsBehind     int
	Selected          bool
}

type Commit struct {
	Hash    string
	Message string
	Date    time.Time
	Author  string
	Diff    string
}

func GetBranches() ([]Branch, error) {
	cmd := exec.Command("git", "branch", "-v", "--format=%(refname:short)|%(committerdate:iso)|%(subject)")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var branches []Branch
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		parts := strings.SplitN(line, "|", 3)
		if len(parts) != 3 {
			continue
		}
		
		branchName := strings.TrimSpace(parts[0])
		if branchName == "" {
			continue
		}
		
		dateStr := strings.TrimSpace(parts[1])
		commitDate, err := time.Parse("2006-01-02 15:04:05 -0700", dateStr)
		if err != nil {
			commitDate = time.Now()
		}
		
		commitMsg := strings.TrimSpace(parts[2])
		
		ahead, behind := getBranchStatus(branchName)
		
		branches = append(branches, Branch{
			Name:              branchName,
			LastCommitDate:    commitDate,
			LastCommitMessage: commitMsg,
			CommitsAhead:      ahead,
			CommitsBehind:     behind,
			Selected:          false,
		})
	}
	
	sort.Slice(branches, func(i, j int) bool {
		return branches[i].LastCommitDate.After(branches[j].LastCommitDate)
	})
	
	return branches, nil
}

func getBranchStatus(branchName string) (ahead, behind int) {
	cmd := exec.Command("git", "rev-list", "--left-right", "--count", "origin/"+branchName+"..."+branchName)
	output, err := cmd.Output()
	if err != nil {
		return 0, 0
	}
	
	parts := strings.Fields(strings.TrimSpace(string(output)))
	if len(parts) == 2 {
		behind, _ = strconv.Atoi(parts[0])
		ahead, _ = strconv.Atoi(parts[1])
	}
	
	return ahead, behind
}

func CheckoutBranch(branchName string) error {
	cmd := exec.Command("git", "checkout", branchName)
	return cmd.Run()
}

func DeleteBranches(branches []string) error {
	for _, branch := range branches {
		cmd := exec.Command("git", "branch", "-d", branch)
		if err := cmd.Run(); err != nil {
			cmd = exec.Command("git", "branch", "-D", branch)
			if err := cmd.Run(); err != nil {
				return err
			}
		}
	}
	return nil
}

func GetCommits() ([]Commit, error) {
	cmd := exec.Command("git", "log", "--pretty=format:%H|%s|%ad|%an", "--date=iso")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	var commits []Commit
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		parts := strings.SplitN(line, "|", 4)
		if len(parts) != 4 {
			continue
		}
		
		hash := parts[0]
		message := parts[1]
		dateStr := parts[2]
		author := parts[3]
		
		commitDate, err := time.Parse("2006-01-02 15:04:05 -0700", dateStr)
		if err != nil {
			commitDate = time.Now()
		}
		
		commits = append(commits, Commit{
			Hash:    hash,
			Message: message,
			Date:    commitDate,
			Author:  author,
		})
	}
	
	return commits, nil
}

func GetCommitDiff(hash string) (string, error) {
	cmd := exec.Command("git", "show", "--stat", "--format=fuller", hash)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
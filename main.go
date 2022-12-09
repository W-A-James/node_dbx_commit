package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type CommitType = string

type ChangeType = string

const (
	Fix      CommitType = "fix"
	Feat                = "feat"
	Docs                = "docs"
	Style               = "style"
	Refactor            = "refactor"
	Perf                = "perf"
	Test                = "test"
	Ci                  = "ci"
	Chore               = "chore"
)

const (
	Breaking    ChangeType = "breaking"
	NonBreaking            = "nonbreaking"
)

func commitTypeValid(commitType string) bool {
	switch commitType {
	case Fix, Feat, Docs, Style, Refactor, Perf, Test, Ci, Chore:
		return true
	default:
		return false
	}

}

func buildCommitMessage(commitType, ticketName, commitMsg string, isBreaking bool) string {
	var breaking string
	if isBreaking {
		breaking = "!"
	} else {
		breaking = ""
	}
	return fmt.Sprintf("%s(%s)%s: %s", commitType, ticketName, breaking, commitMsg)
}

func getTicketNameFromGit() (string, error) {
	statusCmd := exec.Command("git", "branch", "--show-current")
	statusOut, err := statusCmd.Output()
	if err != nil {
		return "", err
	}
	branchName := strings.TrimSpace(string(statusOut))
	ticketName := strings.Split(branchName, "/")[0]

	return ticketName, nil
}

func validateTicketName(ticketName string) bool {
	match, _ := regexp.MatchString("NODE-[0-9]+", ticketName)
	return match
}

func usage() {
}

func main() {
	var commitType string
	var ticketName string
	var isBreaking bool
	var readTicketFromGit bool
	var dryRun bool
	var listCommitTypes bool
	var commitMsg string

	flag.StringVar(&commitType, "type", string(Fix), "")
	flag.StringVar(&commitMsg, "message", "", "Commit message (required)")
	flag.StringVar(&ticketName, "ticket", "", "Name of ticket. (implies autoticket not set)")
	flag.BoolVar(&isBreaking, "breaking", false, "Does this commit contain a breaking change? (false by default)")
	flag.BoolVar(&readTicketFromGit, "autoticket", true, "Gets ticket name from branch name using branch Node driver branch conventions")
	flag.BoolVar(&dryRun, "dryrun", false, "Doesn't perform 'git commit', only prints out the command that would be executed.")
	flag.BoolVar(&listCommitTypes, "list-types", false, "Lists valid commit types and their meanings")
	flag.Parse()

	if listCommitTypes {
		fmt.Printf(`Commit Types
    fix: A bug fix
    feat: Adding a new feature or deprecating (not removing) an existing feature
    docs: Documentation only changes
    style: Changes that do not affect the meaning of the code (white-space, formatting, missing semicolons, etc)
    refactor: A code change that neither fixes a bug nor adds a feature
    perf: A code change that improves performance
    test: Adding missing or correcting existing tests
    ci: Changes to the build process or testing infrastructure
    chore: Changes to auxiliary tools / scripts or dev-dependency upgrades
`)
		os.Exit(0)
	}

	if ticketName != "" {
		readTicketFromGit = false
	}

	if !commitTypeValid(commitType) {
		fmt.Fprintf(os.Stderr, "Invalid commit type: %s\n", commitType)
		flag.Usage()
		os.Exit(1)
	}

	if commitMsg == "" {
		fmt.Fprintln(os.Stderr, "Empty commit message")
		flag.Usage()
		os.Exit(1)
	}

	if readTicketFromGit {
		var err error
		ticketName, err = getTicketNameFromGit()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get ticket name from git. Failed with: '%v'", err)
			os.Exit(1)
		}
	}

	if !validateTicketName(ticketName) {
		fmt.Fprintln(os.Stderr, "Ticket name does not match expected pattern: NODE-xxxx")
		os.Exit(1)
	}

	fullCommitMsg := buildCommitMessage(commitType, ticketName, commitMsg, isBreaking)

	if !dryRun {
		commitCmd := exec.Command("git", "commit", "--message", fullCommitMsg)
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Create new commit with message \"%s\"? (y/n):\n", fullCommitMsg)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(strings.ToLower(text))
		if text == "y" || text == "yes" {
			commandOut, err := commitCmd.Output()

			if err != nil {
				fmt.Fprintf(os.Stderr, "git commit failed with error: %s", err)
				os.Exit(1)
			}
			fmt.Println()
			fmt.Println(string(commandOut))
		} else if text == "n" || text == "no" {
			fmt.Println("Aborting")
			os.Exit(0)
		} else {
			fmt.Println("Invalid option. Aborting")
			os.Exit(1)
		}
	} else {
		fmt.Printf("Commit message: \"%s\"\n", fullCommitMsg)
	}
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "Show help information",
	Long:  `Display help information for all available commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Smak CLI - Git interaction made easier")
		fmt.Println()
		fmt.Println("Available commands:")
		fmt.Println("  smak b      Browse and manage branches interactively")
		fmt.Println("  smak c      Browse commits in current branch")
		fmt.Println("  smak help   Show this help information")
		fmt.Println()
		fmt.Println("Interactive controls:")
		fmt.Println("  ↑↓          Navigate through items")
		fmt.Println("  Enter       Select item")
		fmt.Println("  d           Toggle selection for deletion (in branch view)")
		fmt.Println("  Escape      Return to previous screen")
		fmt.Println("  q           Quit")
	},
}

func init() {
	rootCmd.AddCommand(helpCmd)
}

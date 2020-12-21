package main

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	clientCmd = &cobra.Command{
		Use:  "client <passphrase>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return client(args[0])
		},
	}

	serverCmd = &cobra.Command{
		Use:  "server <passphrase> [default mac address]",
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				defaultAddr = args[1]
			}
			return server(args[0])
		},
	}

	rootCmd = &cobra.Command{
		Use: "wol_proxy",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
)

func main() {
	rootCmd.AddCommand(clientCmd, serverCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

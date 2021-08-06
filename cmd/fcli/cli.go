package main

import (
	"fmt"
	"github.com/ds/depaas/crypto"
	"github.com/spf13/cobra"
)

var cmds = []*cobra.Command{
	{
		Use:     "genKey [CAPrivateKey] [addr]",
		Short:   "gen private pkey for user by admin",
		Long:    `eg. \n ipfs-fcli genKey hello `,
		Aliases: []string{"g"},
		Args:    cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(crypto.GenerateKey(args[0], args[1]))
		},
	},
	{
		Use:     "check [pkey] [file]",
		Short:   "gen file token",
		Long:    `eg. \n ipfs-fcli check aa.png `,
		Aliases: []string{"c"},
		Args:    cobra.MinimumNArgs(2),
		Example: "check \\ \n" +
			"3vQB7B6L2UNY3uZyUSmUTBHXtbCreyMsYVVNw4LRHJZTDyys8ADQt2d33kYDxneSidPtwp3Y1a8BkrPh4YQV3jcqcg8xXSCXe58YdGgbMx4w9r \\ \n" +
			"aa.png",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(crypto.FileKey(args[0], args[1]))
		},
	},
	{
		Use:     "verify [pkey]",
		Short:   "verify pkey owner by",
		Long:    `eg. \n ipfs-fcli v pkey `,
		Aliases: []string{"v"},
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("owner by: %s\n", crypto.RecoverOwner(args[0]))
			fmt.Printf("username: %s\n", crypto.GetUserAddr(args[0]))
		},
	},
	{
		Use:     "token [md5] [fkey]",
		Short:   "verify token information",
		Long:    `eg. \n ipfs-fcli t md5 fkey`,
		Aliases: []string{"t"},
		Args:    cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("pkey: %s", crypto.DeCryptFKey(args[0], args[1]))
		},
	},
}

func main() {
	var rootCmd = &cobra.Command{Use: "fcli"}
	rootCmd.AddCommand(cmds...)
	rootCmd.Execute()
}

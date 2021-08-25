package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/ds/depaas/crypto"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"io"
	"os"
)

var cmds = []*cobra.Command{
	{
		Use:     "genKey [CAPrivateKey] [addr]",
		Short:   "Gen private pkey for user by admin",
		Long:    `eg. \n ipfs-fcli genKey hello `,
		Aliases: []string{"g"},
		Args:    cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(crypto.GenerateKey(args[0], args[1]))
		},
	},
	{
		Use:     "check [pkey] [file]",
		Short:   "Gen file token",
		Long:    `eg. \n ipfs-fcli check aa.png `,
		Aliases: []string{"c"},
		Args:    cobra.MinimumNArgs(2),
		Example: "check \\ \n" +
			"3vQB7B6L2UNY3uZyUSmUTBHXtbCreyMsYVVNw4LRHJZTDyys8ADQt2d33kYDxneSidPtwp3Y1a8BkrPh4YQV3jcqcg8xXSCXe58YdGgbMx4w9r \\ \n" +
			"aa.png",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("fkey", crypto.FileKey(args[0], args[1]))
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
	{
		Use:     "upload [pkey] [file]",
		Short:   "upload file",
		Long:    `eg. \n ipfs-fcli t md5 fkey`,
		Aliases: []string{"t"},
		Args:    cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			bufMd5 := MD5Progress(args[1])
			fmt.Printf("MD5 %s\n", hex.EncodeToString(bufMd5))
			fKey := crypto.FKeyWithMd5(args[0], bufMd5)
			UploadFile(args[1], fKey)
		},
	},
}

func MD5Progress(f string) []byte {
	open, err := os.Open(f)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer open.Close()

	if s, e := os.Stat(f); e == nil {
		bar := progressbar.DefaultBytes(
			s.Size(),
			"CHECK MD5 ",
		)
		h := md5.New()
		if _, err := io.Copy(io.MultiWriter(h, bar), open); err != nil {
			panic(err)
		}
		return h.Sum(nil)
	}
	return nil
}

func main() {
	var rootCmd = &cobra.Command{Use: "fcli"}
	rootCmd.AddCommand(cmds...)
	rootCmd.Execute()
}

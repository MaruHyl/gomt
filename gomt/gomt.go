package main

import (
	"fmt"

	"github.com/MaruHyl/gomt"
	"github.com/spf13/cobra"
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	cmd.Flags().IntVarP(
		&maxLevel, "max_level", "l", 0,
		"set max level")
	cmd.Flags().BoolVarP(
		&json, "json", "j", false,
		"prints out an JSON representation of the tree")
	cmd.Flags().StringVarP(
		&target, "target", "t", "",
		"go mod why ${Target} with better human readability")
}

var maxLevel int
var json bool
var target string

var cmd = &cobra.Command{
	Use:   "gomt",
	Short: "go mod graph with better human readability.",
	Long:  "go mod graph with better human readability.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		opts := []gomt.Option{
			gomt.WithMaxLevel(maxLevel),
			gomt.WithJson(json),
			gomt.WithTarget(target),
		}
		result, err := gomt.Tree(opts...)
		if err != nil {
			panic(err)
		}
		fmt.Println(result)
	},
}

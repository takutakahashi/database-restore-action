/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/takutakahashi/database-restore-action/pkg/config"
	"github.com/takutakahashi/database-restore-action/pkg/database"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "database-restore-action",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		cfgpath, err := cmd.Flags().GetString("config")
		if err != nil {
			logrus.Fatal(err)
		}
		cfg, err := config.Load(cfgpath)
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Info(cfg)
		d, err := database.New(cfg)
		if err != nil {
			logrus.Fatal(err)
		}
		if err := d.Run(); err != nil {
			logrus.Fatal(err)
		}
		logrus.Info("succeeded")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("config", "c", "./sample/config.yaml", "config file path")
}

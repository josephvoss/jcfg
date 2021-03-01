package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Verbose bool
var Debug bool

var log = logrus.New()

var rootCmd = &cobra.Command{
	Use:   "jcfg",
	Short: "Not sure yet, but applies json resources",
}

func init() {
	cobra.OnInitialize()
	// Add base flags here
	rootCmd.PersistentFlags().BoolVarP(
		&Verbose, "verbose", "v", false, "verbose output",
	)
	rootCmd.PersistentFlags().BoolVarP(
		&Debug, "debug", "d", false, "debug output",
	)
}

func setupLogger() {
	log.SetLevel(logrus.WarnLevel)
	if Verbose {
		log.SetLevel(logrus.InfoLevel)
	}
	if Debug {
		log.SetLevel(logrus.DebugLevel)
	}
}

func main() {
	setupLogger()
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf(err.Error())
	}
}

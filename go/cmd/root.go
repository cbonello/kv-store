package cmd

import (
	"fmt"
	"log"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	optIP      string
	optVerbose bool
)

type netAddr struct {
	host string
	port int
}

func (n netAddr) String() string {
	return fmt.Sprintf("%v:%v", n.host, n.port)
}

func printError(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}

func explain(format string, a ...interface{}) {
	if optVerbose {
		msg := fmt.Sprintf(format, a...)
		log.Println(msg)
	}
}

var rootCmd = &cobra.Command{
	Use:   "kv-store",
	Short: "A distributed key-value store",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolVarP(&optVerbose, "verbose", "v", false, "verbosely list operations performed")
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".kv-store" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".kv-store")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

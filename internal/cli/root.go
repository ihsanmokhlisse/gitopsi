package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	output  string
	dryRun  bool
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "gitopsi",
	Short: "Bootstrap production-ready GitOps repositories",
	Long: `gitopsi is a CLI tool that generates complete GitOps repository structures
for Kubernetes, OpenShift, AKS, and EKS platforms.

It supports ArgoCD and Flux, handles both infrastructure and application scopes,
and generates documentation, scripts, and CI/CD configurations.

Examples:
  gitopsi init                     Interactive mode
  gitopsi init --config gitops.yaml   Config file mode
  gitopsi init --dry-run           Preview without writing`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: gitops.yaml)")
	rootCmd.PersistentFlags().StringVar(&output, "output", ".", "output directory")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "preview without writing files")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "verbose output")

	_ = viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	_ = viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	_ = viper.BindPFlag("dry-run", rootCmd.PersistentFlags().Lookup("dry-run"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("gitops")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}

func GetConfig() string {
	return cfgFile
}

func GetOutput() string {
	return output
}

func IsDryRun() bool {
	return dryRun
}

func IsVerbose() bool {
	return verbose
}

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"testing"
)

func TestRootCmd(t *testing.T) {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func TestHelpCmd(t *testing.T) {
	var helpCmd = &cobra.Command{
		Use: "connector-mongodb --help",
	}
	if err := helpCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func TestVersionCmd(t *testing.T) {
	rootCmd.AddCommand(versionCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func TestStoreCommand(t *testing.T) {
	rootCmd.AddCommand(storeCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func TestUpload(t *testing.T) {
	uploadCmd := &cobra.Command{
		Use: "./connector-mongodb store --mongob C:/Users/Ady/Desktop/connector-mongodb/config/db_property_test.json --storj C:/Users/Ady/Desktop/connector-mongodb/storj_config_test.json",
	}
	if err := uploadCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func TestUploadAccessKey(t *testing.T) {
	uploadCmd := &cobra.Command{
		Use: "./connector-mongodb store --accesskey",
	}
	if err := uploadCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func TestUploadShare(t *testing.T) {
	uploadCmd := &cobra.Command{
		Use: "./connector-mongodb store --share",
	}
	if err := uploadCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

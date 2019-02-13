// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/fatih/color"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "handshake",
	Short: "This is a rough POC for handshake core",
	Long: `This is a rough POC for handshake core. To start a new handshake run:

	handshake init
	`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./handshake.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name ".handshake" (without extension).
		viper.AddConfigPath(".")
		viper.SetConfigName("handshake")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// Config is used to save important settings
type Config struct {
	Password string
	ChatID   string
}

// Save saves a config to disk as a yaml file in the existing directory
func (c Config) Save() error {
	d, err := yaml.Marshal(&c)
	if err != nil {
		return err
	}
	os.Remove("handshake.yaml")
	return ioutil.WriteFile("handshake.yaml", d, 0644)
}

func genRandBytes(l int) []byte {
	b := make([]byte, l)
	rand.Read(b)
	return b
}

type Entry struct {
	ID     string `json:"id"`
	Sender string `json:"sender"`
	Sent   int64  `json:"sent"`
	TTL    int64  `json:"ttl"`
	Data   ChatData
}

type ChatData struct {
	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message"`
	TTL       int64  `json:"ttl"`
}

func logPrinter(chatLog []byte, myPeerID string) error {
	var entries []Entry
	if err := json.Unmarshal(chatLog, &entries); err != nil {
		return err
	}
	for _, entry := range entries {
		timeStamp := time.Unix(entry.Sent/1000000000, 0).Format("2006-01-02 15:04:05")
		line := fmt.Sprintf("(%v) %v: %v", timeStamp, entry.Sender[:6], entry.Data.Message)
		if entry.Sender == myPeerID {
			color.Green(line)
		} else {
			color.Yellow(line)
		}
	}
	return nil
}

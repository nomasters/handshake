// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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
	"bufio"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nomasters/handshake"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// newHandshakeCmd represents the newHandshake command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		password := viper.GetString("Password")
		session, err := handshake.NewDefaultSession(password)
		if err != nil {
			log.Fatal(err)
		}
		defer session.Close()

		if len(args) < 1 {
			log.Fatal("invalid arg, must be joiner or initiator")
		}

		switch args[0] {
		case "joiner":
			session.NewPeerWithDefaults()
			share, err := session.ShareHandshakePosition()
			if err != nil {
				log.Fatal(err)
			}
			shareHex := hex.EncodeToString(share)
			fmt.Printf(`share this code with the initiator:
	%v
	
and add the initiator code below.`, shareHex)
			fmt.Print("Enter the initiator code: ")
			reader := bufio.NewReader(os.Stdin)
			hexText, err := reader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}
			hexText = strings.TrimSpace(hexText)
			initiatorShare, err := hex.DecodeString(hexText)
			if err != nil {
				log.Fatal(err)
			}
			if _, err := session.AddPeerToHandshake(initiatorShare); err != nil {
				log.Fatal(err)
			}
			id, err := session.NewChat()
			if err != nil {
				log.Fatal(err)
			}
			config := Config{
				Password: password,
				ChatID:   id,
			}
			config.Save()

		case "initiator":
			session.NewInitiatorWithDefaults()
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter the joiner code: ")
			hexText, err := reader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}
			hexText = strings.TrimSpace(hexText)
			joinerShare, err := hex.DecodeString(hexText)
			if err != nil {
				log.Fatal(err)
			}
			if _, err := session.AddPeerToHandshake(joinerShare); err != nil {
				log.Fatal(err)
			}
			share, err := session.GetHandshakePeerConfig(1)
			if err != nil {
				log.Fatal(err)
			}
			shareHex := hex.EncodeToString(share)
			fmt.Printf(`share this code with the joiner:
	%v
	
and add the initiator code below.`, shareHex)
			id, err := session.NewChat()
			if err != nil {
				log.Fatal(err)
			}
			config := Config{
				Password: password,
				ChatID:   id,
			}
			config.Save()

		default:
			log.Fatal("invalid arg, must be joiner or initiator")
		}
		fmt.Println("congrats. new chat successfully created.")
	},
}

// reader := bufio.NewReader(os.Stdin)
// fmt.Print("Enter text: ")
// text, _ := reader.ReadString('\n')
// fmt.Println(text)

func init() {
	rootCmd.AddCommand(newCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// newHandshakeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// newHandshakeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

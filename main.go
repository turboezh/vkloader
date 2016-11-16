package main

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"vkloader/auth"
	"vkloader/util"
	"vkloader/vkloader"
)

var (
	app          = kingpin.New("vkloader", "VK audio downloader")
	clientId     = app.Arg("clientId", "Client (application) ID").Required().String()
	outputDir    = app.Arg("dir", "Output directory").Required().String()
	userId       = app.Flag("auth-user-id", "VK user ID").Short('u').String()
	accessToken  = app.Flag("aith-access-token", "VK access token").Short('t').String()
	downloadPool = app.Flag("concurent", "Concurent downloads").Short('c').Default(fmt.Sprint(vkloader.DEFAULT_DOWNLOAD_POOL)).Int()
	skipIfExists = app.Flag("skip", "Skip file download if already exists and has non-null length").Short('s').Bool()
)

func main() {
	_, err := app.Parse(os.Args[1:])
	app.FatalIfError(err, "")

	fmt.Printf("Saving to: %s\n", *outputDir)

	authData := &auth.Auth{}

	if len(*accessToken) == 0 || len(*userId) == 0 {
		fmt.Printf("Go to the URL and give me back a redirected URL from browser: %s\n", authData.OAuthUrl(*clientId))

		fmt.Print("URL: ")
		var url string
		fmt.Scanln(&url)

		err := authData.ParseAuthURL(url)
		util.CheckError(err)

	} else {
		authData.SetUserId(*userId)
		authData.SetToken(*accessToken)
	}

	fmt.Printf("User ID: %s\n", authData.UserId())
	fmt.Printf("Token: %s\n", authData.Token())

	loader := vkloader.New(authData, *outputDir)
	loader.SetDownloadPool(*downloadPool).SetSkipIfExists(*skipIfExists)
	loader.Run()
}

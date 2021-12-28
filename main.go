package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"

	"flag"
	"fmt"
)

const API_URL = "http://143.110.156.164:1850/predict"

// Variables used for command line parameters
var (
	Token string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

type Gopher struct {
	Name string `json: "name"`
}

type Judgement struct {
	Judgement int `json:"class_id"`
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// if m.Content == "!gopher" {

	// 	//Call the ML Model API POST the image and get prediction (0->UglyNFT, 1->Not NFT)
	// 	response, err := http.Get(API_URL)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	defer response.Body.Close()

	// 	if response.StatusCode == 200 {
	// 		_, err = s.ChannelFileSend(m.ChannelID, "dr-who.png", response.Body)
	// 		if err != nil {
	// 			fmt.Println(err)
	// 		}
	// 	} else {
	// 		fmt.Println("Error: Can't get dr-who Gopher! :-(")
	// 	}
	// }

	// if m.Content == "!random" {

	// 	//Call the KuteGo API and retrieve a random Gopher
	// 	response, err := http.Get(KuteGoAPIURL + "/gopher/random/")
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	defer response.Body.Close()

	// 	if response.StatusCode == 200 {
	// 		_, err = s.ChannelFileSend(m.ChannelID, "random-gopher.png", response.Body)
	// 		if err != nil {
	// 			fmt.Println(err)
	// 		}
	// 	} else {
	// 		fmt.Println("Error: Can't get random Gopher! :-(")
	// 	}
	// }
	//WE HAVE A PROBLEM BECAUSE THE AVATAR URL CDN
	//GIVES 403 FOR BOT TRYING TO ACCESS IT
	userImageURL := m.Author.AvatarURL("100")
	userImageURL = strings.Replace(userImageURL, "png", "jpg", 1)
	fmt.Println(userImageURL)
	response, e := http.Get(userImageURL)
	if e != nil {
		log.Fatal(e)
	}
	defer response.Body.Close()

	//open file for writing
	file, err := os.Create(fmt.Sprintf("asdf%v.jpg", rand.Intn(100)))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Use io.Copy to dump the response body to file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.Fatal(err)
	}

	// var ready = false
	var judgement = Judgement{}
	//spawn off goroutine to process image using ML model
	// go func() {
	// 	resp, err := http.NewRequest(http.MethodPost, API_URL, file)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	if resp.Response.StatusCode != http.StatusOK {
	// 		fmt.Println("Got bad http response from server")
	// 		return
	// 	}
	// 	content, err := ioutil.ReadAll(resp.Body)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Printf("%s", content)
	// 	judgement = string(content)
	// 	ready = true
	// }()

	s.ChannelMessageSend(m.ChannelID, "Processing Image...")
	//oof
	byteArrResult := SendPostRequestImageURL(API_URL, userImageURL)

	//remove newlines from response
	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, byteArrResult); err != nil {
		fmt.Println(err)
   }
	fmt.Println(string(buffer.String()))
	err = json.Unmarshal(buffer.Bytes(), &judgement)
	if err != nil {
		log.Fatal(err)
	}
	// ready = true
	//var judgeString = ""
	s.ChannelMessageSend(m.ChannelID, userImageURL)
	fmt.Println(judgement.Judgement)
	if judgement.Judgement == 1 {
		s.ChannelMessageSend(m.ChannelID, "Non-NFT")
		//not NFT
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Nice profile pic %s", m.Author.Username))
	} else {
		//DIEEEE NFT BRO
		s.ChannelMessageSend(m.ChannelID, "Ugly NFT")
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s , prepare to die, nft monkey.", m.Author.Username))
	}


	if m.Content == "!hello" {

		// Send a text message with the list of Gophers
		_, err := s.ChannelMessageSend(m.ChannelID, "Hello there.")
		if err != nil {
			fmt.Println(err)
		}
	}
}

func SendPostRequestImageURL(url string, imgURL string) []byte {

	values := map[string]string{"imgURL": imgURL}
	jsonValue, _ := json.Marshal(values)
	response, err := http.Post(API_URL,
		"application/json", bytes.NewBuffer(jsonValue))

	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	fmt.Println(response.Status)
	if err != nil {
		log.Fatal(err)
	}

	return content
}

func SendPostRequestFile(url string, file *os.File, filetype string) []byte {

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(filetype, filepath.Base(file.Name()))

	if err != nil {
		log.Fatal(err)
	}

	io.Copy(part, file)
	writer.Close()
	request, err := http.NewRequest(http.MethodPost, url, body)

	if err != nil {
		log.Fatal(err)
	}

	request.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	fmt.Println(response.Status)
	if err != nil {
		log.Fatal(err)
	}

	return content
}

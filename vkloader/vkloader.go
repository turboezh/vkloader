package vkloader

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"vkloader/auth"
	"vkloader/util"
)

const API_METHOD_ENDPOINT = "https://api.vk.com/method"
const API_METHOD_GET_AUDIO = "audio.get"
const API_VERSION = "5.60"

const AUDIO_CHANNEL_BUFFER = 1000
const DEFAULT_DOWNLOAD_POOL = 10

type Audio struct {
	Id     uint64 `json:"id"`
	Url    string `json:"url"`
	Artist string `json:"artist"`
	Title  string `json:"title"`
}

type AudioCollection struct {
	Count int     `json:"count"`
	Items []Audio `json:"items"`
}

type AudioGetResponse struct {
	Response AudioCollection `json:"response"`
}

type vkLoader struct {
	auth         *auth.Auth
	outputDir    string
	downloadPool int
	skipIfExists bool
}

func New(auth *auth.Auth, outputDir string) *vkLoader {
	return &vkLoader{
		auth:         auth,
		outputDir:    outputDir,
		downloadPool: DEFAULT_DOWNLOAD_POOL,
	}
}

func (o *vkLoader) SetDownloadPool(pool int) *vkLoader {
	o.downloadPool = pool
	return o
}

func (o *vkLoader) DownloadPool() int {
	return o.downloadPool
}

func (o *vkLoader) SetSkipIfExists(skip bool) *vkLoader {
	o.skipIfExists = skip
	return o
}

func (o *vkLoader) SkipIfExists() bool {
	return o.skipIfExists
}

func (o *vkLoader) Run() {
	audios := o.requestAudio()

	wg := &sync.WaitGroup{}
	wg.Add(len(audios))

	audioChannel := make(chan Audio, AUDIO_CHANNEL_BUFFER)
	defer close(audioChannel)

	// download pool
	for i := 0; i < o.downloadPool; i++ {
		go func() {
			for audio := range audioChannel {
				o.downloadTrack(wg, audio)
			}
		}()
	}

	for _, audio := range audios {
		audioChannel <- audio
	}

	wg.Wait()
}

func (o *vkLoader) requestAudio() []Audio {
	u, err := url.Parse(API_METHOD_ENDPOINT)
	util.CheckError(err)

	u.Path = path.Join(u.Path, API_METHOD_GET_AUDIO)
	q := u.Query()
	q.Add("v", API_VERSION)
	q.Add("owner_id", o.auth.UserId())
	q.Add("access_token", o.auth.Token())
	q.Add("count", "5000")

	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	util.CheckError(err)

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	util.CheckError(err)

	data := AudioGetResponse{}
	err = json.Unmarshal(content, &data)
	util.CheckError(err)

	return data.Response.Items
}

func (o *vkLoader) downloadTrack(wg *sync.WaitGroup, audio Audio) {
	defer wg.Done()

	fileName := fileName(&audio)
	filePath := o.filePath(fileName)
	defer fmt.Printf("%s\n", fileName)

	if o.skipIfExists {
		if info, err := os.Stat(filePath); err == nil && info.Size() > 0 {
			return
		}
	}

	resp, err := http.Get(audio.Url)
	util.CheckError(err)
	defer resp.Body.Close()

	file, err := os.Create(filePath)
	util.CheckError(err)
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	util.CheckError(err)
}

func (o *vkLoader) filePath(fileName string) string {
	return path.Join(o.outputDir, fileName)
}

func fileName(audio *Audio) string {
	fileName := fmt.Sprintf("%s - %s.mp3", strings.TrimSpace(audio.Artist), strings.TrimSpace(audio.Title))
	fileName = strings.Replace(fileName, "/", "|", -1)

	return fileName
}

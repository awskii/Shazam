package main

import (
	"bytes"
	"encoding/hex"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const (
	ConfigURI = "http://msft.shazamid.com/orbit/RequestConfig1"
	TagURI    = "http://msft.shazamid.com/orbit/DoAmpTag1"
	boundary  = "AJ8xP50454bf20Gp"

	cryptToken  = "#0xf45452897e53872ea83d740df73282dc"
	deviceId    = "#0x0c7ea0a06633592901d6482f6b3a152cbdfc146fd730a9472758a112c040b61b194ae74f8be3e090" //encrypted guid
	service     = "cn=V12,cn=Config,cn=SmartClub,cn=ShazamiD,cn=services"
	language    = "en-US"
	region      = "#0x584781ba27047d4c"
	deviceModel = "#0xbf428a4ef949ab344a4ddc502417a2b61f5bcbd95215b81fb927dd1e704eb24c4f977a362e7a8f15b6d07aae7961f8b67b8437b5bfa4ba18d95c90a2264ddaeaa83d740df73282dc"
	appID       = "#0xa82b12bc03518af7902c809377343ab9217a5548b44ad048bd94390c354a3ae2"
	duid        = "#0xa82b12bc03518af7902c809377343ab9217a5548b44ad048bd94390c354a3ae2"
	imsi        = "#0x77e6a3eec184455703aeafef7ce062d9452818945cdb0ce59c5a79bd7519ffe2819d7277044c48350f93013088928badff8cf9d15126bad9a83d740df73282dc"

	timezone   = "#0xde4ba34eee1dbbf084dbc274e5865578"
	latitude   = ""
	longtitude = ""
	date       = "#0xf252e7cd866a043cf19eb6c932b2d84982ae4638801ac925b03179237c30e373"
	coverSize  = "#0x6ca5199fc833ba0e"
)

func main() {
	DoRecognition()
}

func newContext() map[string]string {
	args := make(map[string]string)
	args["cryptToken"] = cryptToken
	args["deviceId"] = deviceId
	args["service"] = service
	args["language"] = language
	args["deviceModel"] = deviceModel
	args["applicationIdentifier"] = appID
	args["tagLatitude"] = latitude
	args["tagLongitude"] = longtitude
	args["tagTimezone"] = timezone
	args["tagDate"] = key.EncString(time.Now().Format(time.RFC3339))
	args["coverartSize"] = coverSize
	return args
}

func addEcnrypted(ctx map[string]string) []byte {
	key := new(IceKey)
	key.Init(1)
	guid := getGUID()
	tagId := key.EncString(guid)

	f, _ := os.Open("./sample.wav")
	defer f.Close()
	info, _ := f.Stat()
	fsize := info.Size()

	buf := make([]byte, fsize)
	f.Read(buf)
	sample := key.EncBinary(buf)
	sampleLen := fmt.Sprintf("%d", len(sample))

	ctx["tagId"] = tagId
	ctx["sampleBytes"] = key.EncString(sampleLen)
	ctx["tagTime"] = key.EncString(time.Now().Unix())

	return sample
}

func populateBody(ctx map[string]string, sample []byte) *bytes.Buffer {
	buf := new(bytes.Buffer)
	form := multipart.NewWriter(buf)
	defer form.Close()

	form.SetBoundary(boundary)
	for k, v := range ctx {
		form.WriteField(k, v)
	}
	file, _ := form.CreateFormFile("sample", "sample.sig")
	io.Copy(file, bytes.NewBuffer(sample))
	return buf
}

func DoRecognition() {
	ctx := newContext()
	sample := addEcnrypted(ctx)
	body := populateBody(ctx, sample)

	req, _ := http.NewRequest("POST", TagURI, body)
	req.Header.Add("Content-Type", "multipart/form-data; boundary="+boundary)
	req.Header.Set("Host", "msft.shazamid.com")

	t := &http.Transport{
		DisableKeepAlives:  false,
		DisableCompression: true,
	}

	client := &http.Client{Transport: t}
	resp, _ := client.Do(req)
	io.Copy(os.Stdout, resp.Body)
	resp.Body.Close()
}

func getGUID() string {
	s, _ := exec.Command("/usr/bin/uuidgen").Output()
	return string(s[:])
}

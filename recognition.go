package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const (
	TagURI     = "http://msft.shazamid.com/orbit/DoAmpTag1"
	boundary   = "AJ8xP50454bf20Gp"
	cryptToken = "#0xf45452897e53872ea83d740df73282dc"

	deviceId = "#0x758fd8fd56e13e2a1b7ec2ba3907a10146237cade50a5894174329be2f86f65f16c222d66630b4ec"

	service     = "cn=V12,cn=Config,cn=SmartClub,cn=ShazamiD,cn=services"
	language    = "en-US"
	deviceModel = "#0xbf428a4ef949ab344a4ddc502417a2b61f5bcbd95215b81fb927dd1e704eb24c4f977a362e7a8f15b6d07aae7961f8b67b8437b5bfa4ba18d95c90a2264ddaeaa83d740df73282dc"
	appID       = "#0xa82b12bc03518af7902c809377343ab9217a5548b44ad048bd94390c354a3ae2"
	timezone    = "#0x13fa018b1392e765"
	latitude    = "#0x633448f53f13ea89"
	longtitude  = "#0x34dad30ba2c9c016"
	coverSize   = "#0x6ca5199fc833ba0e"
)

var key = NewIceKey()

func main() {
	DoRecognition()
}

func DoRecognition() {
	body := NewEncryptedBody("./sample.sig")

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

func NewEncryptedBody(fileName string) *bytes.Buffer {
	f, _ := os.Open(fileName)
	defer f.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(f)

	// sample := key.EncBinary(buf.Bytes())
	// sampleLen := fmt.Sprintf("%d", len(sample))

	// fmt.Println(sampleLen, "bytes")
	// fmt.Println(key.EncString(sampleLen))

	ctx := newContext()
	ctx["tagId"] = key.EncString(getGUID())
	// ctx["sampleBytes"] = key.EncString(sampleLen)
	ctx["sampleBytes"] = "#0xbb55a0bd1b0b831a"
	ctx["tagTime"] = key.EncString(fmt.Sprintf("%d", time.Now().UnixNano()))
	return populateBody(ctx, buf)
}

func newContext() map[string]string {
	return map[string]string{
		"applicationIdentifier": appID,
		"cryptToken":            cryptToken,
		"deviceId":              deviceId,
		"service":               service,
		"language":              language,
		"deviceModel":           deviceModel,
		"tagTimezone":           timezone,
		"tagDate":               key.EncString(time.Now().Format(time.RFC3339)),
		"coverartSize":          coverSize,
	}
}

func populateBody(ctx map[string]string, sample *bytes.Buffer) *bytes.Buffer {
	buf := new(bytes.Buffer)
	form := multipart.NewWriter(buf)
	defer form.Close()

	form.SetBoundary(boundary)
	for k, v := range ctx {
		form.WriteField(k, v)
	}
	file, _ := form.CreateFormFile("sample", "sample.sig")
	io.Copy(file, sample)
	return buf
}

func getGUID() string {
	s, _ := exec.Command("/usr/bin/uuidgen").Output()
	return string(s[:])
}

package bitstamp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type BitStamp struct {
}

func (request *BitStamp) getSHA256(data string) string {
	h := hmac.New(sha256.New, []byte(os.Getenv("API_SECRET")))
	h.Write([]byte(data))

	return hex.EncodeToString(h.Sum(nil))
}

func (bitstampClient *BitStamp) GetData(URL string, values url.Values) (string, error) {

	client := http.Client{}
	request, err := http.NewRequest(http.MethodPost, "https://www.bitstamp.net"+URL, strings.NewReader(values.Encode()))
	if err != nil {
		return "", err
	}

	nonce := uuid.New().String()
	timestamp := time.Now().UnixNano() / 1000000
	message := "BITSTAMP" + " " + os.Getenv("API_KEY") +
		http.MethodPost +
		"www.bitstamp.net" +
		URL
	if values != nil {
		message += "application/x-www-form-urlencoded"
	}
	message += nonce +
		strconv.Itoa(int(timestamp)) +
		"v2" +
		values.Encode()

	signature := bitstampClient.getSHA256(message)
	fmt.Println(message)
	request.Header.Set("X-Auth", "BITSTAMP "+os.Getenv("API_KEY"))
	request.Header.Set("X-Auth-Signature", signature)
	request.Header.Set("X-Auth-Nonce", nonce)
	request.Header.Set("X-Auth-Timestamp", strconv.Itoa(int(timestamp)))
	request.Header.Set("X-Auth-Version", "v2")
	if values != nil {
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	res, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", errors.New("request failed ")
	}
	if res != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Println(err)
		return "", readErr
	}

	return string(body), nil
}

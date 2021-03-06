package signal

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/apex/log"
)

// An OperatorChannel signals over a custom http server.
type OperatorChannel struct {
	url string
}

// NewOperatorChannel creates a new OperatorChannel.
func NewOperatorChannel(url string) *OperatorChannel {
	return &OperatorChannel{url: url}
}

// Recv receives a message at the given key.
func (c *OperatorChannel) Recv(key string) (data string, err error) {
	log.WithFields(log.Fields{
		"url": c.url,
		"key": key,
	}).Info("[operator] receive")

	uv := url.Values{
		"address": {key},
	}
	for {
		resp, err := http.Get(c.url + "/sub?" + uv.Encode())
		if err != nil {
			return "", err
		}
		if resp.StatusCode == http.StatusGatewayTimeout {
			resp.Body.Close()
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return "", errors.New(resp.Status)
		}

		bs, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		log.WithFields(log.Fields{
			"key":  key,
			"data": string(bs),
		}).Info("[operator] received")

		return string(bs), nil
	}
}

// Send sends a message to the given key with the given data.
func (c *OperatorChannel) Send(key, data string) error {
	log.WithFields(log.Fields{
		"url":  c.url,
		"key":  key,
		"data": data,
	}).Info("[operator] send")

	uv := url.Values{
		"address": {key},
		"data":    {data},
	}
	for {
		resp, err := http.Get(c.url + "/pub?" + uv.Encode())
		if err != nil {
			return err
		}

		if resp.StatusCode == http.StatusGatewayTimeout {
			resp.Body.Close()
			continue
		}

		ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		return nil
	}
}

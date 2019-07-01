package conhttp

import (
	"io/ioutil"
	"net/http"
)

type HTTPResponse struct {
	Id string
	Status string
	Body   []byte
}

func MakeRequest(id string, url string, ch chan<- HTTPResponse) {
	resp, _ := http.Get(url)
	body, _ := ioutil.ReadAll(resp.Body)
	ch <- HTTPResponse{id, resp.Status, body}
}

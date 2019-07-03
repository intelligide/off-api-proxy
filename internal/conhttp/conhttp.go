package conhttp

import (
	"github.com/astaxie/beego"
	"io/ioutil"
	"net/http"
)

type HTTPResponse struct {
	Id string
	Status string
	Body   []byte
	Err bool
}

const(
	retry = 3
)

func MakeRequest(id string, url string, ch chan<- HTTPResponse) {
	for i := 0; i < retry; i++ {
		resp, err := http.Get(url)
		if err == nil {
			body, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				ch <- HTTPResponse{id, resp.Status, body, false}

				return
			} else {
				beego.Debug(err)

			}
		} else {
			beego.Debug(err)
		}
	}
	ch <- HTTPResponse{id, "", nil, true}
}

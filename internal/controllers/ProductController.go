package controllers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/validation"
	"github.com/intelligide/off-api-proxy/internal/app"
	"github.com/intelligide/off-api-proxy/internal/conhttp"
	"github.com/patrickmn/go-cache"
)

type ProductController struct {
	beego.Controller
}

func createProductUrl(productId string, params url.Values) string {
	provider := app.Config.DataProvider()
	u, err := url.Parse(provider)
	if err != nil {
		panic(err)
	}

	u.Path = path.Join(u.Path, "/api/v0/product/" + productId + ".json")

	u.RawQuery = params.Encode()
	return u.String()
}

func (this *ProductController) GetProduct() {
	id := this.Ctx.Input.Param(":id")

	if app.Config.CacheEnabled() {
		product := app.Cache.Get(id)
		if product != nil {
			beego.Debug("Fetch product " + id + " from cache")
			this.Data["json"] = &product
			this.ServeJSON()
			return
		}
	}

	if this.Ctx.Request.Form == nil {
		this.Ctx.Request.ParseForm()
	}

	q := this.Ctx.Request.Form
	delete(q, "filters")
	urlstring := createProductUrl(id, q)

	beego.Debug("Fetch product " + id + " from " + app.Config.DataProvider() + "(" + urlstring + ")")
	resp, err := http.Get(urlstring)
	if err != nil {
		beego.Error(err)
		this.Data["json"] = nil
		this.Ctx.Output.SetStatus(500)
		this.ServeJSON()
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		beego.Error(err)
		this.Data["json"] = nil
		this.Ctx.Output.SetStatus(500)
		this.ServeJSON()
		return
	}

	result := processResponse(id, body)
	this.Data["json"] = &result
	this.ServeJSON()
}

func (this *ProductController) Batch() {

	// Fetch params
	ids := make([]string, 0, 2)
	this.Ctx.Input.Bind(&ids, "ids")
	ids = unique(ids)

	if len(ids) <= 0 {
		this.Ctx.Output.SetStatus(400)
		return
	}

	// Validation
	valid := validation.Validation{}

	valid.MinSize(ids, 1, "ids");

	for idx, id := range ids {
		valid.Required(id, "ids[" + strconv.Itoa(idx) + "]");
		valid.Numeric(id, "ids[" + strconv.Itoa(idx) + "]");
	}

	if valid.HasErrors() {
		resp := make(map[string]string)

		// If there are error messages it means the validation didn't pass
		// Print error message
		for _, err := range valid.Errors {
			resp[err.Key] = err.Message
		}

		this.Ctx.Output.SetStatus(400)
		this.Data["json"] = &resp
		this.ServeJSON()
		return
	}

	// Process

	q := this.Ctx.Request.Form
	delete(q, "filters")

	var ch chan conhttp.HTTPResponse = make(chan conhttp.HTTPResponse)

	products := make([]interface{}, len(ids))

	thRequests := make([]int, 0, 2)

	for i, id := range ids {

		if app.Config.CacheEnabled() {
			product := app.Cache.Get(id)
			if product != nil {
				beego.Debug("Fetch product " + id + " from cache")
				products[i] = product
				continue
			}
		}

		urlstr := createProductUrl(id, q)
		go conhttp.MakeRequest(id, urlstr, ch)
		thRequests = append(thRequests, i)
	}

	for _, requestIdx := range thRequests {
		response := <-ch
		product := processResponse(response.Id, response.Body)
		products[requestIdx] = product
	}

	resp := make(map[string]interface{})
	resp["products"] = products

	this.Data["json"] = &resp
	this.ServeJSON()
}

func processResponse(idstr string, body []byte) interface{} {
/*
	tst := `{
   "name":"computers",
   "description":"List of computer products",
   "prices":[2400, 2100, 1200, 400.87, 89.90, 150.10],
   "names":["John Doe", "Jane Doe", "Tom", "Jerry", "Nicolas", "Abby"],
   "items":[
      {
         "id":1,
         "name":"MacBook Pro 13 inch retina",
         "price":1350
      },
      {
         "id":2,
         "name":"MacBook Pro 15 inch retina",
         "price":1700
      },
      {
         "id":3,
         "name":"Sony VAIO",
         "price":1200
      },
      {
         "id":4,
         "name":"Fujitsu",
         "price":850
      },
      {
         "id":null,
         "name":"HP core i3 SSD",
         "price":850
      }
   ]
}`*/

	// println(gojsonq.New().FromString(tst).Select("name").String())
	// dat := gojsonq.New().FromString(tst).Select("description", "name", "items.[0].name").Get()


	var dat map[string]interface{}

	if err := json.Unmarshal(body, &dat); err != nil {
		beego.Error(err)
		return nil
	}

	if int(dat["status"].(float64)) == 1 && app.Config.CacheEnabled() {
		_ = app.Cache.Put(idstr, dat, cache.DefaultExpiration)
	}

	return dat
}

func unique(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
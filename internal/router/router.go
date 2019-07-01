package router

import (
	"github.com/astaxie/beego"
	"github.com/intelligide/off-api-proxy/internal/controllers"
)

func init() {
	ns := beego.NewNamespace("/api/",
		beego.NSNamespace("/v0",
			beego.NSRouter("product/batch", &controllers.ProductController{}, "get:Batch"),
			beego.NSRouter("product/:id:int.json", &controllers.ProductController{}, "get:GetProduct"),
		),
	)

	beego.AddNamespace(ns);
}

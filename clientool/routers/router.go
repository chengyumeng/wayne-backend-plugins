package routers

import (
	"github.com/astaxie/beego"

	"github.com/Qihoo360/wayne/src/backend/plugins/clientool/controller"
)

func init() {
	nsWithApp := beego.NewNamespace("/api/v1",
		beego.NSNamespace("/clientool",
			beego.NSInclude(
				&controller.ClientToolController{},
			),
		),
	)

	beego.AddNamespace(nsWithApp)
}

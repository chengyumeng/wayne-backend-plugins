package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

	beego.GlobalControllerRouter["github.com/Qihoo360/wayne/src/backend/plugins/clientool/controller:ClientToolController"] = append(beego.GlobalControllerRouter["github.com/Qihoo360/wayne/src/backend/plugins/clientool/controller:ClientToolController"],
		beego.ControllerComments{
			Method:           "Exec",
			Router:           `/exec`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Params:           nil})

}

package controller

import "github.com/Qihoo360/wayne/src/backend/controllers/base"

// kubetool client 相关操作
type ClientToolController struct {
	base.APIController
}

type Result struct {
	Data interface{}
}

func (c *ClientToolController) URLMapping() {
	c.Mapping("Exec", c.Exec)
}

func (c *ClientToolController) Prepare() {
	c.APIController.Prepare()
}

package task

import (
	"github.com/sniperLx/task-agent/api"
)

type taskRouter struct {
	routes []api.Route
}

func (tr *taskRouter) Routes() []api.Route {
	return tr.routes
}

func NewRouter() api.Router {
	r := &taskRouter{}
	r.initRoutes()
	return r
}

func (tr *taskRouter) initRoutes() {
	tr.routes = []api.Route{
		//单个服务内节点网络可达，可以用这个方法。如果有多个网络不通的节点，他们来自不同服务，这个方法会导致错误
		api.NewPostRoute("/submitCmdTask", tr.submitCmdTask),
		api.NewPostRoute("/cancelCmdTask", tr.cancelCmdTask),
		api.NewPostRoute("/queryCmdTask", tr.queryCmdTask),
		api.NewGETRoute("/checkHealth", tr.checkHealth),
	}
}
package api

type taskRouter struct {
	routes []Route
}

func (tr *taskRouter) Routes() []Route {
	return tr.routes
}

func NewRouter() Router {
	r := &taskRouter{}
	r.initRoutes()
	return r
}

func (tr *taskRouter) initRoutes() {
	tr.routes = []Route{
		//单个服务内节点网络可达，可以用这个方法。如果有多个网络不通的节点，他们来自不同服务，这个方法会导致错误
		NewPostRoute("/submitCmdTask", tr.submitCmdTask),
		NewPostRoute("/cancelCmdTask", tr.cancelCmdTask),
		NewPostRoute("/queryCmdTask", tr.queryCmdTask),
		NewGETRoute("/checkHealth", tr.checkHealth),
	}
}
package common

type CmdRequest struct {
	Name  string   `json:"name,omitempty"`
	Id    string   `json:"id,omitempty"`
	Nodes []string `json:"nodes"`
	Cmd   string   `json:"cmd"`
	//读任务，写任务
	Type string `json:"type,omitempty"`
}

type CmdReply struct {
	Id string `json:"id"`
}

package controller

import (
	"encoding/json"
	"fmt"

	"github.com/Qihoo360/wayne/src/backend/client"
	"github.com/Qihoo360/wayne/src/backend/models"
	"github.com/Qihoo360/wayne/src/backend/util/logs"

	"github.com/gorilla/websocket"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	STDIN   string = "stdin"
	CONNECT string = "connect"
)

var (
	upgrader = websocket.Upgrader{}
)

type Message struct {
	Method string      `json:"method"`
	Data   interface{} `json:"data"`
}

type Options struct {
	Cluster   string
	Namespace string
	Pod       string
	Container string

	Cmd string
}

type TerminalSession struct {
	id       string
	Session  websocket.Conn
	sizeChan chan remotecommand.TerminalSize
}

func (t TerminalSession) Next() *remotecommand.TerminalSize {
	select {
	case size := <-t.sizeChan:
		return &size
	}
	return nil
}

func (t TerminalSession) Read(p []byte) (int, error) {
	_, m, err := t.Session.ReadMessage()
	if err != nil {
		return 0, err
	}
	var msg Message
	json.Unmarshal(m, &msg)
	if msg.Method != STDIN {
		return copy(p, ""), nil
	}
	return copy(p, fmt.Sprintf("%s", msg.Data)), nil
}

func (t TerminalSession) Write(p []byte) (int, error) {

	if err := t.Session.WriteMessage(websocket.TextMessage, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (t TerminalSession) Close(status uint32, reason string) {
	t.Session.Close()
	logs.Info("close socket (%s). %d, %s ", t.id, status, reason)
}

func WaitForTerminal(k8sClient *kubernetes.Clientset, cfg *rest.Config, ts TerminalSession, namespace, pod, container, cmd string) error {
	req := k8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&v1.PodExecOptions{
		Container: container,
		Command:   []string{cmd},
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		return err
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:             ts,
		Stdout:            ts,
		Stderr:            ts,
		TerminalSizeQueue: ts,
		Tty:               true,
	})
	if err != nil {
		return err
	}

	return nil

}

// @router /exec [get]
func (ct *ClientToolController) Exec() {
	fmt.Println("Hello")
	fmt.Println("你好")
	var err error
	c, err := upgrader.Upgrade(ct.Ctx.ResponseWriter, ct.Ctx.Request, nil)
	if err != nil {
		fmt.Print("upgrade:", err)
		return
	}
	getConnectInfo := false
	var opts Options
	var msg Message
	for !getConnectInfo {
		_, message, err := c.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}
		err = json.Unmarshal(message, &msg)
		if err != nil {
			fmt.Println(err)
			break
		}
		if msg.Method != CONNECT {
			fmt.Println(msg)
			continue
		}
		bt, _ := json.Marshal(msg.Data)
		err = json.Unmarshal(bt, &opts)
		if err != nil {
			fmt.Println(err)
			break
		}
		getConnectInfo = true

	}
	ns, err := models.NamespaceModel.GetByName(opts.Namespace)
	if err != nil {

	}
	json.Unmarshal([]byte(ns.MetaData), &ns.MetaDataObj)
	ct.NamespaceId = ns.Id
	ct.CheckPermission(models.PermissionTypeNamespace, models.PermissionUpdate)
	manager, err := client.Manager(opts.Cluster)
	if err == nil {
		ts := TerminalSession{
			Session:  *c,
			sizeChan: make(chan remotecommand.TerminalSize),
		}
		go WaitForTerminal(manager.Client, manager.Config, ts, ns.MetaDataObj.Namespace, opts.Pod, opts.Container, opts.Cmd)
		return
	} else {
		fmt.Println(err)
		return
	}
}

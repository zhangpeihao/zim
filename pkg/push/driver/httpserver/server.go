// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package httpserver

import (
	"github.com/golang/glog"
	"github.com/zhangpeihao/zim/pkg/protocol/serialize"
	"github.com/zhangpeihao/zim/pkg/push"
	"github.com/zhangpeihao/zim/pkg/util"
	"html/template"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	// ServerName 服务名
	ServerName = "push-httpserver"
)

// PushHTTPServerParameter 参数
type PushHTTPServerParameter struct {
	// BindAddress 绑定地址
	BindAddress string
	// Debug 调试模式
	Debug bool
}

// Server 推送服务
type Server struct {
	// PushHTTPServerParameter 参数
	PushHTTPServerParameter
	// handler 回调接口
	handler push.Handler
	// closer 安全退出锁
	closer *util.SafeCloser
	// listener HTTP侦听对象
	listener net.Listener
	// httpServer HTTP服务
	httpServer *http.Server
}

// NewServer 新建服务
func NewServer(params *PushHTTPServerParameter, handler push.Handler) (srv *Server, err error) {
	glog.Infoln("push::driver::httpserver::NewServer")
	srv = &Server{
		PushHTTPServerParameter: *params,
		handler:                 handler,
	}
	srv.httpServer = &http.Server{Handler: srv}
	if srv.Debug {
		glog.Warningln("push server in debug mode!!!")
	}

	return srv, err
}

// Run 启动WebSocket服务
func (srv *Server) Run(closer *util.SafeCloser) (err error) {
	glog.Infoln("push::driver::httpserver::Server::Run()")
	srv.closer = closer
	srv.listener, err = net.Listen("tcp4", srv.BindAddress)
	if err != nil {
		glog.Errorf("push::driver::httpserver::Server::Run() listen(%s) error: %s\n",
			srv.BindAddress, err)
		return
	}
	var httpErr error
	go func() {
		httpErr = srv.httpServer.Serve(srv.listener)
	}()
	time.Sleep(time.Second)
	if httpErr != nil {
		glog.Errorf("push::driver::httpserver::Server::Run() http.Server(%s) error: %s\n",
			srv.BindAddress, err)
		return httpErr
	}
	err = srv.closer.Add(ServerName, func() {
		glog.Warningln("push::driver::httpserver::Server::Run() to close")
		srv.listener.Close()
	})

	return err
}

// Close 退出
func (srv *Server) Close(timeout time.Duration) (err error) {
	glog.Infoln("push::driver::httpserver::Server::Close()")
	defer glog.Warningln("push::driver::httpserver::Server::Close() Done")
	defer srv.closer.Done(ServerName)
	// 关闭HTTP服务
	if srv.listener != nil {
		err = srv.listener.Close()
	}
	return err
}

// Handle 处理HTTP链接
func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("push::driver::httpserver::Server::ServeHTTP()")
	if srv.closer.IsClose() {
		glog.Warningln("push::driver::httpserver::Server::ServeHTTP() URL: ", r.URL.Path, ", Closed")
		return
	}
	route := strings.ToLower(strings.Trim(r.URL.Path, "/"))
	glog.Infoln("push::driver::httpserver::Server::ServeHTTP() route: ", route)
	switch route {
	case "p2u":
		srv.HandlePush2User(w, r)
	case "debug":
		if srv.Debug {
			srv.HandleDebug(w, r)
		} else {
			glog.Warningln("push::driver::httpserver::Server::ServeHTTP() not in debug mode")
			w.WriteHeader(404)
		}
	default:
		glog.Warningln("push::driver::httpserver::Server::ServeHTTP() unknow URL: ", r.URL.Path)
		w.WriteHeader(404)
	}
}

// HandlePush2User 处理HTTP链接
func (srv *Server) HandlePush2User(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("push::driver::httpserver::Server::HandlePush2User()")

	cmd, err := serialize.ParseReader(r.Body)
	if err != nil {
		glog.Warningln("push::driver::httpserver::Server::HandlePush2User() ParseReader error: ", err)
		w.WriteHeader(500)
		return
	}
	srv.handler.OnPushToUser(cmd)
	w.WriteHeader(200)
}

// HandleDebug 处理HTTP链接
func (srv *Server) HandleDebug(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("push::driver::httpserver::Server::HandleDebug()")
	if srv.closer.IsClose() {
		w.WriteHeader(500)
		return
	}
	strs := strings.Split(srv.BindAddress, ":")
	if len(strs) != 2 {
		w.WriteHeader(500)
		return
	}
	homeTemplate.Execute(w, "ws://localhost:"+strs[1]+"/ws")
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<head>
<meta charset="utf-8">
<script>
window.addEventListener("load", function(evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;
    var print = function(message) {
        var d = document.createElement("div");
        d.innerHTML = message;
        output.appendChild(d);
    };
    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };
    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>点击"连接"按钮建立WebSocket连接
<p>点击"关闭"按钮断开连接
<p>在文本框中输入信令内容，点击"发送"按钮，发送信令
<p>
<form>
<button id="open">连接</button>
<button id="close">关闭</button>
<p>
<textarea id="input" rows="5" cols="50"/>
t1
test
login
{"id":"123","timestamp":1234567,"token":"E6B8D4E28E8DF1C331460DE60D9792FF"}
payload
</textarea>
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))

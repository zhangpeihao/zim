// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package websocket

import (
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"github.com/zhangpeihao/zim/pkg/util"
	"html/template"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	// ServerName 服务名
	ServerName = "websocket"
)

// ServerParameter WebSocket服务构造参数
type ServerParameter struct {
	// WSBindAddress WebSocket服务绑定地址
	WSBindAddress string
	// WSSBindAddress WebSocket服务绑定地址
	WSSBindAddress string
	// Debug 调试模式
	Debug bool
	// CertFile 证书文件
	CertFile string
	// KeyFile 密钥文件
	KeyFile string
}

// Server WebSocket服务
type Server struct {
	// ServerParameter WebSocket服务构造参数
	ServerParameter
	// serverHandler Server回调
	serverHandler define.ServerHandler
	// closer 安全退出锁
	closer *util.SafeCloser
	// upgrader upgrader WebSocket upgrade参数
	upgrader *websocket.Upgrader
	// httpListenerlistener HTTP侦听对象
	httpListener net.Listener
	// httpServer HTTP服务
	httpServer *http.Server
	// httpsListener HTTPS侦听对象
	httpsListener net.Listener
	// httpsServer HTTPS服务
	httpsServer *http.Server
}

// NewServer 新建一个WebSocket服务实例
func NewServer(params *ServerParameter, serverHandler define.ServerHandler) (srv *Server, err error) {
	glog.Infoln("websocket::NewServer")
	srv = &Server{
		ServerParameter: *params,
		serverHandler:   serverHandler,
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
	if srv.Debug {
		glog.Warningln("Websocket in debug mode!!!")
	}
	srv.httpServer = &http.Server{Handler: srv}
	srv.httpsServer = &http.Server{Handler: srv}

	return srv, err
}

// Run 启动WebSocket服务
func (srv *Server) Run(closer *util.SafeCloser) (err error) {
	glog.Infoln("websocket::Server::Run()")
	srv.closer = closer
	srv.httpListener, err = net.Listen("tcp4", srv.WSBindAddress)
	if err != nil {
		glog.Errorf("websocket::Server::Run() listen(%s) error: %s\n",
			srv.WSBindAddress, err)
		return
	}
	srv.httpsListener, err = util.NewHTTPSListener(srv.CertFile, srv.KeyFile, srv.WSSBindAddress)
	if err != nil {
		glog.Errorf("websocket::Server::Run() listen(%s) error: %s\n",
			srv.WSBindAddress, err)
		return
	}
	var httpErr, httpsErr error
	go func() {
		httpErr = srv.httpServer.Serve(srv.httpListener)
	}()
	go func() {
		httpsErr = srv.httpsServer.Serve(srv.httpsListener)
	}()

	time.Sleep(time.Second)
	if httpErr != nil {
		glog.Errorf("websocket::Server::Run() http.Server(%s) error: %s\n",
			srv.WSBindAddress, httpErr)
		return httpErr
	}
	if httpsErr != nil {
		glog.Errorf("websocket::Server::Run() HTTPS http.Server(%s) error: %s\n",
			srv.WSSBindAddress, httpsErr)
		return httpsErr
	}
	err = srv.closer.Add(ServerName, func() {
		glog.Warningln("websocket::Server::Run() to close")
		srv.httpListener.Close()
	})

	return err
}

// Close 退出
func (srv *Server) Close(timeout time.Duration) (err error) {
	glog.Infoln("websocket::Server::Close()")
	defer srv.closer.Done(ServerName)
	// 关闭HTTP服务
	if srv.httpListener != nil {
		err = srv.httpListener.Close()
	}
	return err
}

// Handle 处理HTTP链接
func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("websocket::Server::ServeHTTP()")
	if srv.closer.IsClose() {
		glog.Warningln("websocket::Server::ServeHTTP() URL: ", r.URL.Path, ", Closed")
		return
	}
	route := strings.ToLower(strings.Trim(r.URL.Path, "/"))
	glog.Infoln("websocket::Server::ServeHTTP() route: ", route)
	switch route {
	case "ws":
		srv.HandleWebSocket(w, r)
	case "debug":
		if srv.Debug {
			srv.HandleDebug(w, r)
		} else {
			glog.Warningln("websocket::Server::ServeHTTP() not in debug mode")
			w.WriteHeader(404)
		}
	default:
		glog.Warningln("websocket::Server::ServeHTTP() unknow URL: ", r.URL.Path)
		w.WriteHeader(404)
	}
}

// HandleWebSocket 处理HTTP链接
func (srv *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("websocket::Server::HandleWebSocket()")
	// Upgrade到WebSocket连接
	c, err := srv.upgrader.Upgrade(w, r, nil)
	if err != nil {
		glog.Errorf("websocket::Server::Handle() upgrade error: %s\n", err)
		return
	}

	// 新建连接
	conn := NewConnection(c)
	srv.serverHandler.OnNewConnection(conn)

	var cmd *protocol.Command
FOR_LOOP:
	for !srv.closer.IsClose() {
		// 读取Command
		cmd, err = conn.ReadCommand()
		if err != nil {
			if err == define.ErrConnectionClosed {
				glog.Infoln("websocket::Server::Handle() connection to close")
				conn.Close(false)
			}
			break FOR_LOOP
		}
		srv.serverHandler.OnReceivedCommand(conn, cmd)
	}
	glog.Infoln("websocket::Server::Handle() ", conn, " closed")
	srv.serverHandler.OnCloseConnection(conn)
}

// HandleDebug 处理HTTP链接
func (srv *Server) HandleDebug(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("websocket::Server::HandleDebug()")
	if srv.closer.IsClose() {
		w.WriteHeader(500)
		return
	}
	strs := strings.Split(srv.WSBindAddress, ":")
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

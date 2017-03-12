// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package websocket

import (
	"context"
	"net"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"github.com/zhangpeihao/shutdown"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"github.com/zhangpeihao/zim/pkg/util"
)

const (
	// ServerName 服务名
	ServerName = "websocket"
)

// WSParameter WebSocket服务构造参数
type WSParameter struct {
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
	// WSParameter WebSocket服务构造参数
	WSParameter
	// serverHandler Server回调
	serverHandler define.ServerHandler
	// context 环境上下文
	ctx context.Context
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
func NewServer(serverHandler define.ServerHandler) (srv *Server, err error) {
	glog.Infoln("websocket::NewServer")
	srv = &Server{
		WSParameter: WSParameter{
			WSBindAddress:  viper.GetString("gateway.ws-bind"),
			WSSBindAddress: viper.GetString("gateway.wss-bind"),
			Debug:          viper.GetBool("debug"),
			CertFile:       viper.GetString("gateway.wss-cert-file"),
			KeyFile:        viper.GetString("gateway.wss-key-file"),
		},
		serverHandler: serverHandler,
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
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
func (srv *Server) Run(ctx context.Context) (err error) {
	glog.Infoln("websocket::Server::Run()")
	srv.ctx = ctx
	srv.httpListener, err = net.Listen("tcp4", srv.WSBindAddress)
	if err != nil {
		glog.Errorf("websocket::Server::Run() listen(%s) error: %s\n",
			srv.WSBindAddress, err)
		return
	}
	if len(srv.CertFile) == 0 || len(srv.KeyFile) == 0 || len(srv.WSSBindAddress) == 0 {
		glog.Warningln("websocket::Server::Run() https not set")
	} else {
		srv.httpsListener, err = util.NewHTTPSListener(srv.CertFile, srv.KeyFile, srv.WSSBindAddress)
		if err != nil {
			glog.Errorf("websocket::Server::Run() listen(%s) error: %s\n",
				srv.WSBindAddress, err)
			return
		}
	}
	var httpErr, httpsErr error
	go func() {
		httpErr = srv.httpServer.Serve(srv.httpListener)
	}()
	if srv.httpsListener != nil {
		go func() {
			httpsErr = srv.httpsServer.Serve(srv.httpsListener)
		}()
	}

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

	return err
}

// Close 退出
func (srv *Server) Close(timeout time.Duration) (err error) {
	glog.Infoln("websocket::Server::Close()")
	defer glog.Warningln("websocket::Server::Close() Done")
	// 关闭HTTP服务
	if srv.httpListener != nil {
		err = srv.httpListener.Close()
	}
	return err
}

// Handle 处理HTTP链接
func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("websocket::Server::ServeHTTP()")
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
	// Get token
	glog.Infoln("token: ", r.Header.Get("token"))
	// Upgrade到WebSocket连接
	c, err := srv.upgrader.Upgrade(w, r, nil)
	if err != nil {
		glog.Errorf("websocket::Server::Handle() upgrade error: %s\n", err)
		return
	}
	if err := shutdown.ExitWaitGroupAdd(srv.ctx, 1); err != nil {
		glog.Errorf("websocket::Server::HandleWebSocket() ExitWaitGroupAdd error: %s", err)
		return
	}
	defer shutdown.ExitWaitGroupDone(srv.ctx)

	// 新建连接
	conn := NewConnection(c)
	srv.serverHandler.OnNewConnection(conn)

	var cmd *protocol.Command
FOR_LOOP:
	for {
		// 读取Command
		cmd, err = conn.ReadCommand()
		if err != nil {
			if err == define.ErrNoMoreMessage {
				continue FOR_LOOP
			}
			if err == define.ErrConnectionClosed {
				glog.Infoln("websocket::Server::Handle() connection to close")
				conn.Close(false)
			}
			break FOR_LOOP
		}
		err = srv.serverHandler.OnReceivedCommand(conn, cmd)
		if err != nil {
			glog.Warningln("websocket::Server::Handle() error:", err)
			conn.Close(true)
			break FOR_LOOP
		}
	}
	glog.Infoln("websocket::Server::Handle() ", conn, " closed")
	srv.serverHandler.OnCloseConnection(conn)
}

// HandleDebug 处理HTTP链接
func (srv *Server) HandleDebug(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("websocket::Server::HandleDebug()")
	if r.TLS == nil {
		homeTemplate.Execute(w, "ws://"+r.Host+"/ws")
	} else {
		homeTemplate.Execute(w, "wss://"+r.Host+"/ws")
	}
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head lang="en">
    <meta charset="UTF-8">
    <title>zim Demo</title>
    <link rel="stylesheet" href="//zimcloud.github.io/static/vendor/bootstrap/bootstrap.min.css">
    <link rel="stylesheet" href="//zimcloud.github.io/static/css/demo.css">
</head>
<body>
    <div id="container" class="container">
        <div id="log"></div>
        <form id="form" action="" class="form-inline">
            <div id="controllers">
                <div id="controller-text">
                    <input type="text" id="msg" class="form-control" size="60" />
                </div>
                <div id="controller-submit">
                    <button type="submit" class="btn btn-success">Send</button>
                </div>
            </div>
        </form>
    </div>
    <script type="text/javascript" src="//zimcloud.github.io/static/vendor/jquery/jquery.min.js"></script>
    <script type="text/javascript" src="//zimcloud.github.io/static/vendor/blueimp/md5.min.js"></script>
    <script type="text/javascript" src="//zimcloud.github.io/static/js/zim.js"></script>
    <script type="text/javascript" src="//zimcloud.github.io/static/js/demo.js"></script>
</body>
</html>
`))

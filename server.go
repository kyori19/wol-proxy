package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

var (
	t = &tpl{
		templates: template.Must(template.ParseFiles("wol.gohtml")),
	}
	upgrader    = websocket.Upgrader{}
	connection  *clientConnection
	defaultAddr string

	errRejectMultipleClient = errors.New("multiple client tried to connect simultaneously")
	errTimeout              = errors.New("response time out")
)

const (
	cmdInfo = "info"
	cmdWake = "wake"
)

type clientConnection struct {
	socket   *websocket.Conn
	response <-chan []byte
}

type tpl struct {
	templates *template.Template
}

type data struct {
	NotConnected   bool
	ConnErr        string
	ClientResponse string
	DefaultAddr    string
	FormErr        string
	FormSuccess    bool
}

type form struct {
	Address string `form:"address"`
}

func (t *tpl) Render(w io.Writer, name string, data interface{}, _ echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func htmlController(c echo.Context) error {
	d := func() data {
		if connection == nil {
			return data{
				NotConnected: true,
			}
		}

		if err := connection.socket.WriteMessage(websocket.TextMessage, []byte(cmdInfo)); err != nil {
			return data{
				ConnErr: err.Error(),
			}
		}

		select {
		case msg := <-connection.response:
			return data{
				ClientResponse: string(msg),
				DefaultAddr:    defaultAddr,
			}
		case <-time.After(5 * time.Second):
			return data{
				NotConnected: true,
			}
		}
	}()

	sess, err := session.Get("wol", c)
	if err != nil {
		return err
	}

	if e := sess.Flashes("error"); len(e) > 0 {
		d.FormErr = e[0].(string)
	}
	if s := sess.Flashes("success"); len(s) > 0 {
		d.FormSuccess = s[0].(bool)
	}

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}
	return c.Render(http.StatusOK, "wol.gohtml", d)
}

func postController(c echo.Context) error {
	sess, err := session.Get("wol", c)
	if err != nil {
		return err
	}

	if err := func() error {
		var f form
		if err := c.Bind(&f); err != nil {
			return err
		}

		if err := connection.socket.
			WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s %s", cmdWake, f.Address))); err != nil {
			return err
		}

		select {
		case msg := <-connection.response:
			switch {
			case string(msg) == resDone:
				return nil
			case strings.HasPrefix(string(msg), resError):
				return errors.New(strings.SplitN(string(msg), " ", 2)[1])
			}
		case <-time.After(5 * time.Second):
			return errTimeout
		}
		return nil
	}(); err != nil {
		sess.AddFlash(err.Error(), "error")
	} else {
		sess.AddFlash(true, "success")
	}

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, c.Request().URL.Path)
}

func wsController(c echo.Context) error {
	log.Infof("Connection from client: %s", c.RealIP())
	if connection != nil {
		return errRejectMultipleClient
	}

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer func() {
		ws.Close()
		connection = nil
	}()

	response := make(chan []byte)
	connection = &clientConnection{
		socket:   ws,
		response: response,
	}

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			if strings.Contains(err.Error(), "close 1000") {
				return nil
			} else {
				log.Fatal(err)
			}
		}
		response <- msg
	}
}

func server(pass string) error {
	e := echo.New()
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Renderer = t

	e.File("/favicon.ico", "favicon.ico")

	r := e.Group(fmt.Sprintf("/%s", pass))
	r.GET("/streaming", wsController)
	r.GET("/wol", htmlController)
	r.POST("/wol", postController)

	return e.Start(":3000")
}

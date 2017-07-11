package main

import (
	"fmt" // optional

	"gopkg.in/kataras/iris.v6"
	"gopkg.in/kataras/iris.v6/adaptors/httprouter"
	"gopkg.in/kataras/iris.v6/adaptors/view"
	"gopkg.in/kataras/iris.v6/adaptors/websocket"
	"io/ioutil"
	"encoding/json"
	"net/http"
	"strings"
	"github.com/dgrijalva/jwt-go"
)

type TargetUrl struct {
	HangupUrl   string
	CallUrl     string
	Token       string
}

var myChatRoom = "room1"

/*
func newCorsMiddleware() iris.HandlerFunc {
	options := cors.Options{
		AllowedOrigins: []string{"*"},
		AllowCredentials: true,
	}
	handlerWithNext := cors.New(options)

	// this is the only func you will have to use if you're going
	// to make use of any external net/http middleware.
	// iris.ToHandler converts the net/http middleware to an iris-compatible.
	return iris.ToHandler(handlerWithNext)
}
*/

func main() {
	config,err := getConfig()
	if err != nil {
		panic(err)
	}
	app := iris.New()
	app.Adapt(iris.DevLogger())                  // enable all (error) logs
	app.Adapt(httprouter.New())                  // select the httprouter as the servemux
	app.Adapt(view.HTML("./templates", ".html")) // select the html engine to serve templates

	ws := websocket.New(websocket.Config{
		// the path which the websocket client should listen/registered to,
		Endpoint: "/check_call",
		// the client-side javascript static file path
		// which will be served by Iris.
		// default is /iris-ws.js
		// if you change that you have to change the bottom of templates/client.html
		// script tag:
		ClientSourcePath: "/iris-ws.js",
		//
		// Set the timeouts, 0 means no timeout
		// websocket has more configuration, go to ../../config.go for more:
		// WriteTimeout: 0,
		// ReadTimeout:  0,
		// by-default all origins are accepted, you can change this behavior by setting:
		// CheckOrigin: (r *http.Request ) bool {},
		//
		//
		// IDGenerator used to create (and later on, set)
		// an ID for each incoming websocket connections (clients).
		// The request is an argument which you can use to generate the ID (from headers for example).
		// If empty then the ID is generated by DefaultIDGenerator: randomString(64):
		// IDGenerator func(ctx *iris.Context) string {},
		CheckOrigin    :  func(r *http.Request) bool {
			return true
		},
	})

	app.Adapt(ws) // adapt the websocket server, you can adapt more than one with different Endpoint

	app.StaticWeb("/js", "./static/js") // serve our custom javascript code
	app.Get("/", func(ctx *iris.Context) {
		ctx.Render("client.html", iris.Map{"Client Page": ctx.Host(),"Host":ctx.Host()})
	})

	app.Post("/hangup", func(ctx *iris.Context) {
		ctx.FormValues()
		urlData := ctx.Request.Form
		resp, err := http.Post(config.CallUrl, "application/x-www-form-urlencoded", strings.NewReader(urlData.Encode()))

		if err != nil {
			ctx.JSON(iris.StatusNotFound,iris.Map{
				"error" : err,
			})
		}else{
			body, _ := ioutil.ReadAll(resp.Body)
			ctx.SetHeader("Content-Type","application/json")
			ctx.Data(resp.StatusCode,body)
		}
	})

	app.Post("/call", func(ctx *iris.Context) {
		ctx.FormValues()
		urlData := ctx.Request.Form
		resp, err := http.Post(config.CallUrl, "application/x-www-form-urlencoded", strings.NewReader(urlData.Encode()))

		if err != nil {
			ctx.JSON(iris.StatusNotFound,iris.Map{
				"error" : err,
			})
		}else{
			body, _ := ioutil.ReadAll(resp.Body)
			data := CallResponse{}
			if err := json.Unmarshal(body,&data);err == nil {
				//clientId :=
				if data.Status == "success" {

					sendCall(ws,ctx.FormValue("clientId"),ctx.FormValue("user_username"),data)
				}
			}
			ctx.SetHeader("Content-Type","application/json")
			ctx.Data(resp.StatusCode,body)
		}
	})



	ws.GetConnectionsByRoom(myChatRoom)

	ws.OnConnection(func(c websocket.Connection) {
		c.Join(myChatRoom)
		a := c.Context()
		tokenString := a.GetCookie("JWT")
		hmacSampleSecret := []byte(config.Token)
		//tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJuYmYiOjE0NDQ0Nzg0MDB9.u1riaD1rW97opCoAuRCTy4w58Br-Zk-bh7vLiRIsrpU"

		// Parse takes the token string and a function for looking up the key. The latter is especially
		// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
		// head of the token to identify which key to use, but the parsed token (head and claims) is provided
		// to the callback, providing flexibility.
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return hmacSampleSecret, nil
		})

		var flag = false

		if err == nil {
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				var clientId,userName string;
				if clientId,ok = claims["client_id"].(string);ok{
					if userName,ok = claims["username"].(string);ok{
						roomName := clientId + "_" + userName
						c.Join(roomName)
						flag = true
					}
				}
				/*
				 $clarm = [
					'client_id'     => $userData['client_id'],
					'department_id' => $userData['department_id'],
					'username'      => $userData['username'],
					'fullname'      => $userData['fullname'],
					'gender_id'     => $userData['gender_id'],
					'date_of_birth' => $userData['date_of_birth'],
					'avatar'        => $userData['avatar'],
					'email'         => $userData['email'],
					'client_name'   => $userData['client_id'],
				];
				*/
			}
		}

		if !flag {
			c.Disconnect()
		}

		c.On("chat", func(message string) {
			if message == "leave" {
				c.Leave(myChatRoom)
				c.To(myChatRoom).Emit("chat", "Client with ID: "+c.ID()+" left from the room and cannot send or receive message to/from this room.")
				c.Emit("chat", "You have left from the room: "+myChatRoom+" you cannot send or receive any messages from others inside that room.")
				return
			}
			c.To(myChatRoom).Emit("chat", "From: "+c.ID()+": "+message)
		})

		// or create a new leave event
		// c.On("leave", func() {
		// 	c.Leave(myChatRoom)
		// })

		c.OnDisconnect(func() {
			fmt.Printf("Connection with ID: %s has been disconnected!\n", c.ID())
		})
	})

	app.Listen(":8080")
}

func sendCall(s websocket.Server,clientId,userName string,data CallResponse )  {
	room := clientId + "_" + userName
	all := s.GetConnectionsByRoom(room)
	for _,conn := range all {
		conn.Emit("call",data)
	}
}

func getConfig() (target TargetUrl,err error) {
	content, err := ioutil.ReadFile("./config/config.json")
	if err != nil {
		return
	}
	err = json.Unmarshal(content,&target)
	return
}

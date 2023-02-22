package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"math"
	"math/rand"
	"net/http"
	"time"
)

/**

Database

create table "Player"
(
    "Login"            varchar   not null primary key,
    "Password"         varchar   not null,
    "Nickname"         varchar   not null,
    "RegistrationDate" timestamp not null
);

create table "Message"
(
    "ID"     serial primary key,
    "Player" varchar references "Player",
    "Text"   varchar,
    "Date"   timestamp
);

*/

const (
	TypeMessage    = 1
	TypeTyping     = 2
	TypeUpdateGame = 3
	TypeOnlineList = 4
)

var router *gin.Engine
var players map[string]*Player

func main() {
	Connect()

	players = make(map[string]*Player)

	for i := 0; i < 25; i++ {
		name := fmt.Sprintf("bot-%d", i)
		players[name] = &Player{
			Name: name,
			Conn: nil,
			Coords: Point{
				X: float64(rand.Int63n(600)),
				Y: float64(rand.Int63n(800)),
			},
			Color:  "#FF00FF",
			Radius: 20,
		}
	}

	router = gin.Default()

	router.LoadHTMLGlob("html/*.html")
	router.Static("assets", "assets")

	router.GET("/", func(context *gin.Context) {
		context.HTML(200, "index.html", nil)
	})
	router.GET("/game/:nickname", func(context *gin.Context) {
		context.HTML(200, "game.html", context.Param("nickname"))
	})
	router.GET("/sign-up", func(context *gin.Context) {
		context.HTML(200, "sign-up.html", nil)
	})
	router.POST("/sign-up", signUpHandler)
	router.POST("/sign-in", signInHandler)
	router.GET("/user/list", selectUserList)

	websocketConnector := websocket.Upgrader{
		ReadBufferSize:    4096,
		WriteBufferSize:   4096,
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// WS://
	router.GET("/connect/:name", func(context *gin.Context) {
		name := context.Param("name")

		connection, e := websocketConnector.Upgrade(
			context.Writer,
			context.Request, nil)
		if e != nil {
			fmt.Println(e)
			return
		}

		fmt.Println("Connect client from",
			connection.RemoteAddr().String())

		player := &Player{
			Name:   name,
			Conn:   connection,
			Coords: Point{},
			Color:  "#FF0000",
			Radius: 100,
		}

		players[name] = player

		SendOnlineUser()

		go ReadPlayer(player)
	})

	go UpdateGame()

	_ = router.Run("127.0.0.1:8085")
}

func signUpHandler(context *gin.Context) {
	player := &PlayerType{}
	e := context.BindJSON(player)
	if e != nil {
		context.Status(400)
		return
	}

	if IsExistsName(player.Login) {
		context.AbortWithStatusJSON(400, "Пользователь с таким логином уже существует")
		return
	}

	InsertPlayerName(player)
	context.Status(200)
}

func signInHandler(context *gin.Context) {
	player := &PlayerType{}
	e := context.BindJSON(player)
	if e != nil {
		context.Status(400)
		return
	}

	if SingIn(player) {
		context.JSON(200, player.Nickname)
	} else {
		context.JSON(400, "неверный логин/пароль")
	}
}

func selectUserList(context *gin.Context) {
	context.JSON(200, SelectUsers())
}

func UpdateGame() {
	for {
		playerList := make([]*Player, 0)
		for _, player := range players {
			playerList = append(playerList, player)
		}
		transferData, e := json.Marshal(playerList)
		if e != nil {
			continue
		}
		data, _ := json.Marshal(Transfer{
			Type: 3,
			Data: string(transferData),
		})
		for _, player := range players {
			if player.Conn == nil {
				continue
			}
			_ = player.Conn.WriteMessage(
				websocket.TextMessage,
				data,
			)
		}

		time.Sleep(time.Millisecond * 25)
	}
}

func SendOnlineUser() {
	data := GetOnlineList()
	for _, p := range players {
		if p.Conn != nil {
			_ = p.Conn.WriteMessage(websocket.TextMessage, data)
		}
	}
}

func GetOnlineList() []byte {
	onlineList := make([]string, 0)
	for name, p := range players {
		if p.Conn != nil {
			onlineList = append(onlineList, name)
		}
	}
	data, e := json.Marshal(onlineList)
	if e != nil {
		fmt.Println(e)
		return nil
	}
	transfer := &Transfer{
		Type: 4,
		Data: string(data),
	}
	data, e = json.Marshal(transfer)
	if e != nil {
		fmt.Println(e)
		return nil
	}
	return data
}

func ReadPlayer(player *Player) {
	defer func() {
		e := recover()
		if e != nil {
			fmt.Println(e)
		}
		delete(players, player.Name)

		SendOnlineUser()
	}()

	for {
		_, data, e := player.Conn.ReadMessage()
		if e != nil {
			continue
		}

		transfer := &Transfer{}
		e = json.Unmarshal(data, transfer)
		if e != nil {
			continue
		}

		switch transfer.Type {
		case TypeMessage:
			transfer := &Transfer{}
			e = json.Unmarshal(data, transfer)
			if e != nil {
				continue
			}

			message := &Message{}
			e = json.Unmarshal([]byte(transfer.Data), message)
			if e != nil {
				continue
			}

			InsertMessage(message)

			data, e = json.Marshal(message)
			if e != nil {
				// TODO отпарвить пользователю, что его сообщение не было доставлено
				continue
			}

			transfer.Data = string(data)

			data, e = json.Marshal(transfer)
			if e != nil {
				// TODO отпарвить пользователю, что его сообщение не было доставлено
				continue
			}

			for _, p := range players {
				if player != p && p.Conn != nil {
					_ = p.Conn.WriteMessage(websocket.TextMessage, data)
				}
			}
		case TypeTyping:
			for _, p := range players {
				if player != p && p.Conn != nil {
					_ = p.Conn.WriteMessage(websocket.TextMessage, data)
				}
			}
		case TypeUpdateGame:
			direction := &Point{}
			e = json.Unmarshal([]byte(transfer.Data), direction)
			if e != nil {
				continue
			}

			if direction.IsMove() {
				player.Coords.X += direction.X
				player.Coords.Y += direction.Y

				FindIntersection(player)
			}
		case TypeOnlineList:
			_ = player.Conn.WriteMessage(websocket.TextMessage, GetOnlineList())
		}
	}
}

func FindIntersection(player *Player) {
	if player.IsDead {
		return
	}
	var radiusSum float64
	playerRadius := player.Radius * 0.5
	for _, p := range players {
		if p == player {
			continue
		}
		if p.IsDead {
			continue
		}

		otherPlayerRadius := p.Radius * 0.5

		radiusSum = otherPlayerRadius + playerRadius
		radiusSum = radiusSum * radiusSum
		x := player.Coords.X - p.Coords.X
		y := player.Coords.Y - p.Coords.Y

		x = x*x + y*y

		if radiusSum >= x {
			if playerRadius > otherPlayerRadius {
				p.IsDead = true
				player.Radius += math.Sqrt(otherPlayerRadius)
			} else {
				player.IsDead = true
				p.Radius += math.Sqrt(playerRadius)
			}
		}
	}
}

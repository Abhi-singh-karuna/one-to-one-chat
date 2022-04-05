package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/googollee/go-socket.io"
	"log"
)

type Msg struct {
	UserId    string   `json:"user_id"`
	Text      string   `json:"text"`
	State     string   `json:"state"`
	NameSpace string   `json:"name_space"`
	Rooms     []string `json:"rooms"`
}

type Connect struct {
	UserId string   `json:"user_id"`
	Text   string   `json:"text"`
	Rooms  []string `json:"rooms"`
}

func main() {
	router := gin.New()

	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		msg := Msg{
			UserId: s.ID(),
			Text:   "on connect",
		}
		s.SetContext("")
		s.Emit("connection", msg)
		fmt.Println("connected : ", s.ID())
		return nil
	})

	server.OnEvent("/", "join", func(s socketio.Conn, room string) {
		s.Join(room)
		x := s.Rooms()
		msg := Msg{
			s.ID(),
			" join " + room,
			"state",
			s.Namespace(),
			x,
		}
		fmt.Println("join", room, s.Namespace(), s.Rooms())
		server.BroadcastToRoom("/", room, "res", msg)
	})

	server.OnEvent("/", "leave", func(s socketio.Conn, room string) {
		s.Leave(room)
		msg := Msg{
			s.ID(),
			" leave " + room, "state",
			s.Namespace(),
			s.Rooms(),
		}
		fmt.Println("/:leave ", room, s.Namespace(), s.Rooms())
		server.BroadcastToRoom("/", room, "res", msg)
	})

	server.OnEvent("/", "chat", func(s socketio.Conn, msg string) {
		res := Msg{
			s.ID(),
			msg,
			"normal",
			s.Namespace(),
			s.Rooms(),
		}

		s.SetContext(res)

		fmt.Println("chat received", msg, s.Namespace(), s.Rooms(), s.Rooms())

		rooms := s.Rooms()
		
		if len(rooms) > 0 {
			fmt.Println("broadcast to ", rooms)
			for i := range rooms {
				server.BroadcastToRoom("/", rooms[i], "res", res)
			}
		}
	})

	server.OnEvent("/", "notice", func(s socketio.Conn, msg string) {
		fmt.Println("notice", msg)
		s.Emit("reply", "have "+msg)
	})

	server.OnEvent("/chat", "msg", func(s socketio.Conn, msg string) string {
		fmt.Println("/chat:msg received", msg)
		return "recv" + msg
	})

	server.OnEvent("/", "bye", func(s socketio.Conn, msg string) string {
		last := s.Context().(Msg)
		s.Emit("bye", last)
		res := Msg{s.ID(), s.ID() + " leave", "state", s.Namespace(), s.Rooms()}
		rooms := s.Rooms()
		s.LeaveAll()
		s.Close()
		if len(rooms) > 0 {
			fmt.Println("broadcast to ", rooms)
			for i := range rooms {
				server.BroadcastToRoom("/", rooms[i], "res", res)
			}
		}
		fmt.Printf("/:bye last context: %+v \n", last)
		return last.Text
	})

	server.OnError("/", func(s socketio.Conn, err error) {
		fmt.Println("/:error", err)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("/:closed", s.ID(), reason)
	})

	go server.Serve()
	defer server.Close()

	log.Println("Severing at localhost:5000...")

	router.GET("/socket.io/*any", gin.WrapH(server))
	router.POST("/socket.io/*any", gin.WrapH(server))

	router.Run(":5000")
}

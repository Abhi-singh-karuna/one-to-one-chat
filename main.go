package main

import (
	"fmt"

	"github.com/gin-gonic/gin"

	socketio "github.com/googollee/go-socket.io"
)

type Messaage struct {
	SenderID  string `json:"sender_id"`
	ReciverID string `json:"reciver_id"`
	Text      string `json:"text"`
}

type Msg struct {
	UserId    string   `json:"user_id"`
	Text      string   `json:"text"`
	Rooms     []string `json:"rooms"`
}

func main() {
	router := gin.New()
	server := socketio.NewServer(nil)

	var users = make(map[string]interface{})

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		fmt.Println("connected:", s.ID())
		s.Emit("allconectedusers", users)
		// s.Join("bcast")
		return nil
	})

	server.OnEvent("/", "username", func(s socketio.Conn, name string) {
		fmt.Println("new-users:", name)
		users[s.ID()] = name
		// s.Join("bcast")
		s.Emit("newuser", users[s.ID()])
		fmt.Println("allconectedusers : ", users)
		for key, _ := range users {
			server.BroadcastToRoom("/", key, "allconectedusers", users)
		}
	})

	server.OnEvent("/", "chat", func(s socketio.Conn, res Messaage) {
		res.SenderID = s.ID()
		s.SetContext(res)
		// fmt.Println("chat received", res, s.Namespace(), s.Rooms(), s.Rooms())
		fmt.Println("send to  ID:  " ,res.ReciverID+" from  ID :"+ s.ID()+"  message  :"+res.Text)
		var x string = res.ReciverID
		server.BroadcastToRoom("/", x, "message", res)
	})

	server.OnDisconnect("/", func(s socketio.Conn, msg string) {
		fmt.Println("disconnected:", users[s.ID()])
		s.Emit("disconnecteduser", users[s.ID()])
		delete(users, s.ID())
		for key, _ := range users {
			fmt.Println("key : ", key)
			server.BroadcastToRoom("/", key, "allconectedusers", users)
		}
		fmt.Println("closed", msg)
	})

	/// chat in room

	server.OnEvent("/", "join", func(s socketio.Conn, room string) {
		s.Join(room)
		fmt.Println("join     :    ", room)
		x := s.Rooms()
		msg := Msg{
			s.ID(),
			" join " + room,
			x,

		}
		fmt.Println("join", room, s.Namespace(), s.Rooms())
		server.BroadcastToRoom("/", room, "join_room", msg)
	})

	server.OnEvent("/", "leave", func(s socketio.Conn, room string) {
		s.Leave(room)
		msg := Msg{
			s.ID(),
			" leave " + room,
			s.Rooms(),
		}
		fmt.Println("leave", room, s.Namespace(), s.Rooms())
		server.BroadcastToRoom("/", room, "leave_from_room", msg)
	})

	server.OnEvent("/", "chatingroup", func(s socketio.Conn, msg string) {
		res := Msg{
			s.ID(),
			msg,
			s.Rooms(),
		}

		s.SetContext(res)

		fmt.Println("chat received", msg, s.Namespace(), s.Rooms(), s.Rooms())

		rooms := s.Rooms()
		
		if len(rooms) > 0 {
			fmt.Println("broadcast to ", rooms)
			for i := range rooms {
				server.BroadcastToRoom("/", rooms[i], "chat_in_group", res)
			}
		}
	})

	/// ---- 



	go server.Serve()
	defer server.Close()

	router.GET("/socket.io/*any", gin.WrapH(server))
	router.POST("/socket.io/*any", gin.WrapH(server))

	router.Run(":5000")
}

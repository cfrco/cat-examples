package main

import (
    "io"
    "fmt"
    "html"

    "github.com/gin-gonic/gin"
)

var (
    rooms map[string]*Broker
)

func main() {
    rooms = make(map[string]*Broker)

    r := gin.Default()
    r.LoadHTMLGlob("./templates/*.html")
    r.Static("/static", "./static")

    r.GET("/chat/:chat_room", chatGet)
    r.GET("/receiver/:chat_room", receiverGet)
    r.POST("/send/:chat_room", sendPost)

    r.Run()
}

func getRoom(name string) *Broker {
    b, ok := rooms[name]
    if !ok {
        b = NewBroker()
        rooms[name] = b
    }

    return b
}

func chatGet(c *gin.Context) {
    c.HTML(200, "chat.html", gin.H{
        "RoomName": c.Param("chat_room"),
    });
}

func receiverGet(c *gin.Context) {
    // Get room
    roomName := c.Param("chat_room")
    b := getRoom(roomName)

    // New channel
    client := make(chan string, 20)
    b.Add(client)
    defer b.Remove(client)

    // Stream
    c.Stream(func(w io.Writer) bool {
        select {
            case <- c.Writer.CloseNotify():
                return false
            case message := <-client:
                c.SSEvent("message", message)
        }
        return true
    })
}

func sendPost(c *gin.Context) {
    roomName := c.Param("chat_room")
    b := getRoom(roomName)

    message := html.EscapeString(c.PostForm("message"))
    userName := html.EscapeString(c.PostForm("user_name"))

    b.Broadcast(fmt.Sprintf("%s : %s", userName, message))
    c.String(200, "OK")
}

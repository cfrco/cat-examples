package main

import (
    "bufio"
    "html"

    "golang.org/x/net/websocket"
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
    r.GET("/ws/:chat_room", chatWs)

    r.Run()
}

func chatGet(c *gin.Context) {
    c.HTML(200, "chat.html", gin.H{
        "RoomName": c.Param("chat_room"),
    });
}

func chatWs(c *gin.Context) {
    // Get room
    roomName := c.Param("chat_room")
    userName := html.EscapeString(c.Query("user_name"))
    b := getRoom(roomName)

    // New channel
    client := make(chan string, 20)
    b.Add(client)
    defer b.Remove(client)

    handler := websocket.Handler(getHandler(b, client, userName))
    handler.ServeHTTP(c.Writer, c.Request)
}

func getHandler(b *Broker, out chan string, name string) func(ws *websocket.Conn) {
    return func(ws *websocket.Conn) {
        go readAndBroadcast(ws, b, name)
        writeTo(out, ws)
    }
}

func readAndBroadcast(ws *websocket.Conn, b *Broker, name string) {
    reader := bufio.NewReader(ws)
    line := ""
    for {
        buf, isPrefix, err := reader.ReadLine()
        if err != nil {
            break
        }
        line += string(buf)
        if !isPrefix {
            b.Broadcast(name + ":" + line)
            line = ""
        }
    }
}

func writeTo(out chan string, ws *websocket.Conn) {
    for {
        msg, ok := <-out
        if !ok {
            break
        }

        msg = msg + "\n"
        ws.Write([]byte(msg))
    }
}

func getRoom(name string) *Broker {
    b, ok := rooms[name]
    if !ok {
        b = NewBroker()
        rooms[name] = b
    }

    return b
}

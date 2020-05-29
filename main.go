package main


import(
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
	"bufio"

	"github.com/gorilla/websocket"
)


var  addr= flag.String("addr","echo.websocket.org", "http service address")

func main() {

	flag.Parse()
	log.SetFlags(0)
	interrupt := make(chan os.Signal,1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{
		Scheme: "ws",
		Host: *addr,
		Path: "/",
	}
	log.Printf("connecting to %s", u.String())
	c,_,err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:",err)
	}
	defer c.Close()
	done:= make(chan struct{})

	go func(){
		defer close(done)
		for {
			_,message,err := c.ReadMessage()
			if err != nil {
				log.Println("read:" , err)
				return
			}
			log.Printf("recv: %s" , message)

		}
	}()

	reader := bufio.NewReader(os.Stdin)
	tc := make(chan string)

	go func (){
		for {
			select {
			case <-done:
				return
			default:
				text , _  := reader.ReadString('\n')
				tc <- text
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case t := <-tc:
			fmt.Printf("out x\n")
			err:= c.WriteMessage(websocket.TextMessage,[]byte(t))
			if err != nil {
				log.Println("write:",err)
				return
			}

		case <-interrupt:
			log.Println("interrupt")

			err:= c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure,""))
			if err != nil {
				log.Println("write close:",err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return 
		}

	}
}



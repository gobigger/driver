package socket_default

import (
	"sync"
)


type Hub struct {
	mutex		sync.Mutex
	clients		map[string]*Client
	broadcast	chan *Msg
	register	chan *Client
	unregister	chan *Client
	closer		chan bool
}

type Msg struct {
	id		string	//个人消息=id， 广播时=channel
	data	[]byte
}


func newHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan *Msg),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		closer:		make(chan bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			if client != nil {
				h.clients[client.id] = client
			}
		case client := <-h.unregister:
			if client != nil {
				close(client.sender)
				close(client.closer)
				client.ws.Close()
				delete(h.clients, client.id)
			}
		case msg := <-h.broadcast:
			if msg != nil {
				for _,client := range h.clients {
					if yes,ok := client.channels[msg.id]; ok && yes {
						client.sender <- msg.data
					}
				}
			}
		case <-h.closer:
			break
		}
		
	}
}



//整个关的时候，不用每一个client的chan都去关
//因为基本都是退出进程才会
func (h *Hub) close() {
	close(h.broadcast)
	close(h.register)
	close(h.unregister)
	close(h.closer)
}
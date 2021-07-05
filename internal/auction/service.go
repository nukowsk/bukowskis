package auction

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/nukowsk/bukowskis/internal/sender"
	"github.com/nukowsk/bukowskis/internal/store"
)

type AuctionService struct {
	mx     sync.Mutex
	server *http.Server
}

// XXX: This can probably just be called Service in the acution package
func NewAuctionService(
	port string,
	proxy http.Handler,
	sender sender.Sender,
	store store.Store,
	gasGetter GasGetter) (*AuctionService, error) { // XXX: remove error?

	handler := NewHandler(gasGetter, store, sender, proxy)
	server := &http.Server{Addr: ":" + port, Handler: handler}
	return &AuctionService{
		mx:     sync.Mutex{},
		server: server,
	}, nil

}

func (t *AuctionService) Run() {
	if err := t.server.ListenAndServe(); err != nil {
		log.Println(err)
	}
}

func (t *AuctionService) Stop() {
	t.mx.Lock()
	defer t.mx.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := t.server.Shutdown(ctx); err != nil {
		panic(err)
	}
}

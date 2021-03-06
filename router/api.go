package router

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/struckoff/kvrouter/rpcapi"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

func (h *Router) HTTPHandler() *httprouter.Router {
	r := httprouter.New()
	//r.POST("/node", h.HTTPRegister)
	r.GET("/nodes", h.Nodes)
	r.POST("/put/:key", h.Store)
	r.GET("/get/:key", h.Receive)
	r.GET("/list", h.Explore)
	return r
}

// Store value for a given key on the remote node
func (h *Router) Store(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if r.Body != nil {
		defer r.Body.Close()
	}
	key := ps.ByName("key")
	n, err := h.LocateKey(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := n.Store(key, b); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := fmt.Fprint(w, "OK"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

//Receive value for a given key from the remote node
func (h *Router) Receive(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if r.Body != nil {
		defer r.Body.Close()
	}
	key := ps.ByName("key")
	n, err := h.LocateKey(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var body []byte
	body, err = n.Receive(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

//Explore returns a list of keys on nodes
func (h *Router) Explore(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	res, err := h.nodeKeys()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b, err := res.JsonMarshal()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(b); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Nodes returns a list of nodes
func (h *Router) Nodes(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	metas, err := h.nodes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(metas); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Router) nodes() ([]rpcapi.NodeMeta, error) {
	ns, err := h.bal.Nodes()
	if err != nil {
		return nil, err
	}
	metas := make([]rpcapi.NodeMeta, len(ns))
	for iter, n := range ns {
		metas[iter] = n.Meta()
	}
	return metas, nil
}

func (h *Router) nodeKeys() (*SyncMap, error) {
	var wg sync.WaitGroup
	res := NewSyncMap()
	ns, err := h.bal.Nodes()
	if err != nil {
		return nil, err
	}
	for _, n := range ns {
		wg.Add(1)
		go func(wg *sync.WaitGroup, n Node, sm *SyncMap) {
			defer wg.Done()
			keys, err := n.Explore()
			if err != nil {
				log.Printf("%s: %s", n.ID(), err.Error())
				return
			}
			sm.Put(n.ID(), keys)
		}(&wg, n, res)
	}
	wg.Wait()
	return res, nil
}

package mongo

import (
	"github.com/funkygao/fae/config"
	"github.com/funkygao/golib/breaker"
	log "github.com/funkygao/log4go"
	"labix.org/v2/mgo"
	"sync"
	"time"
)

type Client struct {
	conf *config.ConfigMongodb

	selector ServerSelector

	lk           sync.Mutex
	freeconn     map[string][]*mgo.Session
	breakers     map[string]*breaker.Consecutive
	throttleConn map[string]chan interface{}
}

func New(cf *config.ConfigMongodb) (this *Client) {
	this = new(Client)
	this.conf = cf
	this.breakers = make(map[string]*breaker.Consecutive)
	this.throttleConn = make(map[string]chan interface{})

	switch cf.ShardStrategy {
	case "legacy":
		this.selector = NewLegacyServerSelector(cf.ShardBaseNum)

	default:
		this.selector = NewStandardServerSelector(cf.ShardBaseNum)
	}
	this.selector.SetServers(cf.Servers)

	go this.runWatchdog()

	return
}

func (this *Client) FreeConnMap() map[string][]*mgo.Session {
	this.lk.Lock()
	defer this.lk.Unlock()
	return this.freeconn
}

func (this *Client) Session(pool string, shardId int32) (*Session, error) {
	server, err := this.selector.PickServer(pool, int(shardId))
	if err != nil {
		return nil, err
	}

	sess, err := this.getConn(server.Uri())
	if err != nil {
		return nil, err
	}

	return &Session{Session: sess, client: this, server: server}, nil
}

func (this *Client) WarmUp() {
	var (
		sess *mgo.Session
		err  error
		t1   = time.Now()
	)
	for retries := 0; retries < 3; retries++ {
		for _, server := range this.selector.ServerList() {
			sess, err = this.getConn(server.Uri())
			if err != nil {
				log.Error("Warmup %v fail: %s", server.Uri(), err)
				break
			} else {
				this.putFreeConn(server.Uri(), sess)
			}
		}

		if err == nil {
			break
		}
	}

	if err == nil {
		log.Trace("Mongodb warmed up within %s: %+v",
			time.Since(t1), this.freeconn)
	} else {
		log.Error("Mongodb failed to warm up within %s: %s",
			time.Since(t1), err)
	}

}

func (this *Client) getConn(uri string) (*mgo.Session, error) {
	sess, ok := this.getFreeConn(uri)
	if ok {
		return sess, nil
	}

	return this.dial(uri)
}

func (this *Client) dial(uri string) (*mgo.Session, error) {
	if this.breakers[uri].Open() {
		return nil, ErrCircuitOpen
	}

	this.throttleConn[uri] <- true
	defer func() {
		// release throttle
		<-this.throttleConn[uri]
	}()

	sess, err := mgo.DialWithTimeout(uri, this.conf.ConnectTimeout)
	if err != nil {
		this.breakers[uri].Fail()
		return nil, err
	}

	this.breakers[uri].Succeed()
	sess.SetSocketTimeout(this.conf.IoTimeout)
	sess.SetMode(mgo.Monotonic, true)

	return sess, nil
}

func (this *Client) putFreeConn(uri string, sess *mgo.Session) {
	this.lk.Lock()
	defer this.lk.Unlock()
	if this.freeconn == nil {
		this.freeconn = make(map[string][]*mgo.Session)
	}
	freelist := this.freeconn[uri]
	if len(freelist) >= this.conf.MaxIdleConnsPerServer {
		sess.Close()
		return
	}
	this.freeconn[uri] = append(this.freeconn[uri], sess)
}

func (this *Client) getFreeConn(uri string) (sess *mgo.Session, ok bool) {
	this.lk.Lock()
	defer this.lk.Unlock()
	if _, present := this.breakers[uri]; !present {
		this.breakers[uri] = &breaker.Consecutive{
			FailureAllowance: this.conf.Breaker.FailureAllowance,
			RetryTimeout:     this.conf.Breaker.RetryTimeout}
		this.throttleConn[uri] = make(chan interface{}, this.conf.MaxConnsPerServer)
	}

	if this.freeconn == nil {
		return nil, false
	}
	freelist, present := this.freeconn[uri]
	if !present || len(freelist) == 0 {
		return nil, false
	}

	// it is no longer free
	sess = freelist[len(freelist)-1] // last item
	this.freeconn[uri] = freelist[:len(freelist)-1]
	return sess, true
}

// caller is responsible for lock
func (this *Client) killConn(session *mgo.Session) {
	for addr, sessions := range this.freeconn {
		for idx, sess := range sessions {
			if sess == session { // pointer addr compare
				// https://code.google.com/p/go-wiki/wiki/SliceTricks
				this.freeconn[addr] = append(this.freeconn[addr][:idx],
					this.freeconn[addr][idx+1:]...)
			}
		}
	}
}

package peer

import (
	"bufio"
	log "github.com/funkygao/log4go"
	"time"
)

func (this *Peer) runHeartbeat() {
	var msg = peerMessage{}
	var err error
	msg["ip"] = this.selfAddr
	sleepInterval := time.Duration(this.heartbeatInterval) * time.Second
	for {
		if err = this.publish(msg); err != nil {
			log.Error("Publish fails: %v", err)
		}

		for ip, lastAccess := range this.neighbors {
			if time.Since(lastAccess).Seconds() > this.deadThreshold {
				// he is dead
				this.killNeighbor(ip)
			}
		}

		time.Sleep(sleepInterval)
	}

}

func (this *Peer) discoverPeers() {
	defer func() {
		this.c.Close() // leave the multicast group
	}()

	var msg peerMessage
	reader := bufio.NewReader(this.c)
	for {
		// net.ListenMulticastUDP sets IP_MULTICAST_LOOP=0 as
		// default, so you never receive your own sent data
		// if you run both sender and receiver on (logically) same IP host
		line, _, err := reader.ReadLine()
		if err != nil {
			log.Error("discoverPeers: %v", err)
			continue
		}

		if err := msg.unmarshal(line); err != nil {
			// Not our protocol, it may be SSDP or else
			continue
		}

		neighborIp, present := msg["ip"]
		if !present {
			log.Info("Peer msg has no 'ip'")
			continue
		}

		this.refreshNeighbor(neighborIp.(string))
	}
}

func (this *Peer) killNeighbor(ip string) {
	this.rwmutex.Lock()
	defer this.rwmutex.Unlock()

	delete(this.neighbors, ip)
	this.picker.DelPeer(ip)
	log.Info("Peer[%s] killed", ip)

	log.Debug("Neighbors: %+v", this.neighbors)
}

func (this *Peer) refreshNeighbor(ip string) {
	this.rwmutex.Lock()
	defer this.rwmutex.Unlock()

	if _, present := this.neighbors[ip]; !present {
		log.Info("Peer[%s] becoming rookie", ip)
		this.picker.AddPeer(ip)
	}

	this.neighbors[ip] = time.Now()
}

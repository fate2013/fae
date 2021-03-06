package engine

import (
	"github.com/funkygao/fae/config"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
	"time"
)

type configProcessManagement struct {
	mode                   string
	maxOutstandingSessions int
	startServers           int
	minSpareServers        int32
	spawnServers           int
}

func (this *configProcessManagement) loadConfig(section *conf.Conf) {
	this.mode = section.String("mode", "static")
	this.startServers = section.Int("start_servers", 1000)
	this.minSpareServers = int32(section.Int("min_spare_servers", 200))
	this.spawnServers = section.Int("spawn_servers_n", 100)
	this.maxOutstandingSessions = section.Int("max_outstanding_sessions", 2000)
}

func (this *configProcessManagement) dynamic() bool {
	return this.mode == "dynamic"
}

type configRpc struct {
	sessionSlowThreshold float64 // in seconds per session
	callSlowThreshold    float64 // in seconds per call
	listenAddr           string
	sessionTimeout       time.Duration
	ioTimeout            time.Duration
	framed               bool
	protocol             string
	debugSession         bool
	tcpNoDelay           bool
	statsOutputInterval  time.Duration
	pm                   configProcessManagement
}

func (this *configRpc) loadConfig(section *conf.Conf) {
	this.listenAddr = section.String("listen_addr", "")
	if this.listenAddr == "" {
		panic("Empty listen_addr")
	}

	this.sessionSlowThreshold = section.Float("session_slow_threshold", 5)
	this.callSlowThreshold = section.Float("call_slow_threshold", 5)
	this.sessionTimeout = time.Duration(section.Int("session_timeout",
		0)) * time.Second
	this.ioTimeout = time.Duration(section.Int("io_timeout", 0)) * time.Second
	this.framed = section.Bool("framed", false)
	this.protocol = section.String("protocol", "binary")
	this.tcpNoDelay = section.Bool("tcp_nodelay", true)
	this.debugSession = section.Bool("debug_session", false)
	this.statsOutputInterval = time.Duration(section.Int("stats_output_interval",
		0)) * time.Second

	// pm section
	this.pm = configProcessManagement{}
	sec, err := section.Section("pm")
	if err != nil {
		panic(err)
	}
	this.pm.loadConfig(sec)

	log.Debug("rpc: %+v", *this)
}

type engineConfig struct {
	*conf.Conf

	httpListenAddr  string
	pprofListenAddr string

	rpc *configRpc
}

func (this *Engine) LoadConfigFile() *Engine {
	log.Info("Engine[%s] loading config file %s", BuildID, this.configFile)

	cf := new(engineConfig)
	var err error
	cf.Conf, err = conf.Load(this.configFile)
	if err != nil {
		panic(err)
	}

	this.conf = cf
	this.doLoadConfig()

	return this
}

func (this *Engine) doLoadConfig() {
	this.conf.httpListenAddr = this.conf.String("http_listen_addr", "")
	this.conf.pprofListenAddr = this.conf.String("pprof_listen_addr", "")

	// rpc section
	this.conf.rpc = new(configRpc)
	section, err := this.conf.Section("rpc")
	if err != nil {
		panic(err)
	}
	this.conf.rpc.loadConfig(section)

	section, err = this.conf.Section("servants")
	if err != nil {
		panic(err)
	}
	config.LoadServants(section)

	log.Debug("engine: %+v", *this.conf)
}

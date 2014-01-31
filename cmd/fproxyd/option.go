package main

import (
	log "code.google.com/p/log4go"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
)

func parseFlags() {
	flag.StringVar(&options.logLevel, "loglevel", "info", "log level")
	flag.StringVar(&options.logFile, "log", "stdout", "log file")
	flag.StringVar(&options.configFile, "conf", "etc/fxi.cf", "config file")
	flag.StringVar(&options.lockFile, "lockfile", "var/fxi.lock", "lockfile path")
	flag.BoolVar(&options.showVersion, "version", false, "show version and exit")
	flag.IntVar(&options.tick, "tick", 60*10, "watchdog ticker length in seconds")
	flag.StringVar(&options.cpuprof, "cpuprof", "", "cpu profiling file")
	flag.StringVar(&options.memprof, "memprof", "", "memory profiling file")
	flag.Usage = showUsage

	flag.Parse()

	if options.tick <= 0 {
		panic("tick must be possitive")
	}
}

func showUsage() {
	fmt.Fprint(os.Stderr, USAGE)
	flag.PrintDefaults()
}

func setupProfiler() {
	if options.cpuprof != "" {
		f, err := os.Create(options.cpuprof)
		if err != nil {
			panic(err)
		}

		globals.Printf("CPU profiler %s enabled\n", options.cpuprof)
		pprof.StartCPUProfile(f)
	}

	if options.memprof != "" {
		globals.Printf("MEM profiler %s enabled\n", options.memprof)
	}
}

func setupLogging(loggingLevel, logFile string) {
	level := log.DEBUG
	switch loggingLevel {
	case "info":
		level = log.INFO
	case "warn":
		level = log.WARNING
	case "error":
		level = log.ERROR
	}

	for _, filter := range log.Global {
		filter.Level = level
	}

	if logFile == "stdout" || logFile == "" {
		log.AddFilter("stdout", level, log.NewConsoleLogWriter())
	} else {
		logDir := filepath.Dir(logFile)
		if err := os.MkdirAll(logDir, 0744); err != nil {
			panic(err)
		}

		writer := log.NewFileLogWriter(logFile, false)
		log.AddFilter("file", level, writer)
		writer.SetFormat("[%D %t] [%L] (%S) %M")
		writer.SetRotate(true)
		writer.SetRotateSize(0)
		writer.SetRotateLines(0)
		writer.SetRotateDaily(true)
	}
}

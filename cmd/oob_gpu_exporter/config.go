package main

import (
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/smc-public/oob_gpu_exporter/internal/collector"
	"github.com/smc-public/oob_gpu_exporter/internal/config"
	"github.com/smc-public/oob_gpu_exporter/internal/log"
)

func ReloadConfig(filename string) {
	cfg := config.NewConfig()
	old := config.Config

	log.Info("Configuration reload was triggered")

	if len(filename) > 0 {
		err := cfg.FromFile(filename)
		if err != nil {
			log.Error("Failed to %v", err)
			return
		}
	}

	cfg.FromEnvironment()
	err := cfg.Validate()
	if err != nil {
		log.Error("Invalid configuration: %v", err)
		return
	}

	old.Mutex.Lock()
	defer old.Mutex.Unlock()

	for k, v := range cfg.Hosts {
		h, ok := old.Hosts[k]
		if ok {
			if h.Username != v.Username || h.Password != v.Password || h.Scheme != v.Scheme {
				old.Hosts[k] = v
				collector.Reset(k)
			}
		} else {
			old.Hosts[k] = v
		}
	}

	log.Info("Configuration reload was successful")
}

func WatchConfig(filename string) {
	lastReload := time.Now()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error("Failed to start file watcher: %v", err)
		return
	}
	defer func(){
		err := watcher.Close()
		if err != nil {
			log.Error("Failed to close file watcher: %v", err)
		}
	}()

	err = watcher.Add(filename)
	if err != nil {
		log.Error("Failed to watch configuration file: %v", err)
		return
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if time.Since(lastReload) < time.Second {
				break // needed to deduplicate e.g. multiple write events
			}
			reload := false
			if event.Has(fsnotify.Write) {
				reload = true
			} else if event.Has(fsnotify.Remove) {
				err := watcher.Remove(event.Name)
				if err != nil {
					log.Error("Failed to remove file watcher: %v", err)
					return
				}
				err = watcher.Add(filename)
				if err != nil {
					return
				}
				reload = true
			}
			if reload {
				lastReload = time.Now()
				ReloadConfig(filename)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Error("File watcher error: %v", err)
		}
	}
}

func LoadConfig(filename string) {
	cfg := config.NewConfig()

	if len(filename) > 0 {
		err := cfg.FromFile(filename)
		if err != nil {
			log.Fatal("Failed to %v", err)
		}
		log.Info("Loaded configuration file: %s", filename)
	}

	cfg.FromEnvironment()
	err := cfg.Validate()
	if err != nil {
		log.Fatal("Invalid configuration: %v", err)
	}

	config.SetConfig(cfg)

	if len(filename) > 0 {
		go WatchConfig(filename)
	}
}

package main

import (
	"hoa-control-app/cmd/store"
	"hoa-control-app/cmd/watcher"
	"log"
)

func main() {

	log.Println("starting the control app for house of apps")

	watcher := watcher.NewTaskPopulator(store.NewInMemStorage(), &watcher.TaskWatcherService{})
	watcher.StartWatching()
}

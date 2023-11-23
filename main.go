package main

import (
	"hoa-control-app/cmd/store"
	"hoa-control-app/cmd/watcher"
	"log"
)

func main() {
	log.Println("starting the control app for house of apps")

	appWatcher := watcher.NewTaskPopulator(&store.InMemStore{}, &watcher.TaskWatcherService{})

	appWatcher.StartWatching()
}

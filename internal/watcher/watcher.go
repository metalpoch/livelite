package watcher

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/metalpoch/livelite/internal/storage"
)

func WatchAndUploadHLS(localDir, remotePrefix string, store storage.Storage, playlistDelay time.Duration) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	lastPlaylistUpload := time.Now().Add(-playlistDelay)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				ext := filepath.Ext(event.Name)
				if event.Op&(fsnotify.Create|fsnotify.Write) != 0 {
					// Subir .ts apenas estén listos
					if ext == ".ts" {
						go func(filename string) {
							const minAge = 1 * time.Second
							for {
								info, err := os.Stat(filename)
								if err != nil {
									return // archivo borrado o inaccesible
								}
								if time.Since(info.ModTime()) > minAge {
									file, err := os.Open(filename)
									if err == nil {
										defer file.Close()
										remoteName := filepath.Join(remotePrefix, filepath.Base(filename))
										if err := store.Put(file, remoteName); err != nil {
											log.Printf("upload error: %v", err)
										}
									}
									break
								}
								time.Sleep(500 * time.Millisecond)
							}
						}(event.Name)
					}
					// Subir .m3u8 solo cada playlistDelay
					if ext == ".m3u8" {
						go func(filename string) {
							if time.Since(lastPlaylistUpload) >= playlistDelay {
								file, err := os.Open(filename)
								if err == nil {
									defer file.Close()
									remoteName := filepath.Join(remotePrefix, filepath.Base(filename))
									if err := store.Put(file, remoteName); err != nil {
										log.Printf("upload error: %v", err)
									}
									lastPlaylistUpload = time.Now()
								}
							}
						}(event.Name)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("watcher error: %v", err)
			}
		}
	}()
	return watcher.Add(localDir)
}

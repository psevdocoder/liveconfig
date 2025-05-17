package liveconfig

import (
	"log/slog"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
)

type CallbackFunc func(newValue, oldValue any)

type LiveConfig struct {
	mu          sync.RWMutex
	configData  map[string]any
	subscribers map[string][]CallbackFunc
	filePath    string
}

var (
	instance *LiveConfig
	once     sync.Once
)

// Init инициализирует глобальный экземпляр LiveConfig
func Init(filePath string) {
	once.Do(func() {
		instance = &LiveConfig{
			configData:  make(map[string]any),
			subscribers: make(map[string][]CallbackFunc),
			filePath:    filePath,
		}
		if err := instance.loadConfig(); err != nil {
			panic("Failed to load config: " + err.Error())
		}
		go instance.watchFile()
	})
}

func (lc *LiveConfig) loadConfig() error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	content, err := os.ReadFile(lc.filePath)
	if err != nil {
		return err
	}

	var newData map[string]any
	if err = yaml.Unmarshal(content, &newData); err != nil {
		return err
	}

	lc.compareAndNotify(newData)
	lc.configData = newData

	return nil
}

func (lc *LiveConfig) watchFile() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer func(watcher *fsnotify.Watcher) {
		err = watcher.Close()
		if err != nil {
			slog.Error("Failed to close watcher", slog.Any("err", err))
		}
	}(watcher)

	if err = watcher.Add(lc.filePath); err != nil {
		panic(err)
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				time.Sleep(20 * time.Millisecond)
				err = lc.loadConfig()
				if err != nil {
					slog.Error("Failed to load config", slog.Any("err", err))
				}
			}
		case err = <-watcher.Errors:
			panic(err)
		}
	}
}

func (lc *LiveConfig) compareAndNotify(newData map[string]any) {
	for key, callbacks := range lc.subscribers {
		oldValue := getValueByPath(lc.configData, key)
		newValue := getValueByPath(newData, key)

		if !reflect.DeepEqual(oldValue, newValue) {
			for _, callback := range callbacks {
				go callback(newValue, oldValue)
			}
		}
	}
}

func (lc *LiveConfig) Sub(key string, callback CallbackFunc) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	lc.subscribers[key] = append(lc.subscribers[key], callback)
}

func getValueByPath(data map[string]any, path string) any {
	keys := strings.Split(path, ".")
	var value any = data
	for _, key := range keys {
		if m, ok := value.(map[any]any); ok {
			value = m[key]
		} else if m, ok := value.(map[string]any); ok {
			value = m[key]
		} else {
			return nil
		}
	}
	return value
}

// Public API

// Sub подписывается на изменения параметра
func Sub(key string, callback CallbackFunc) {
	if instance == nil {
		panic("LiveConfig is not initialized. Call liveconfig.Init() first.")
	}
	instance.Sub(key, callback)
}

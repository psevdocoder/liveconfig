package main

import (
	"fmt"

	"git.server.home/pkg/liveconfig"
)

func main() {
	liveconfig.Init("example/config.yaml")

	// Подписка на изменения параметра
	liveconfig.Sub("app.notify_before", func(newValue, oldValue any) {
		fmt.Printf("Параметр 'app.notify_before' изменился с %v на %v\n", oldValue, newValue)
	})

	liveconfig.Sub("app.notify_after", func(newValue, oldValue any) {
		fmt.Printf("Параметр 'app.notify_after' изменился с %v на %v\n", oldValue, newValue)
	})

	liveconfig.Sub("app.array", func(newValue, oldValue any) {
		fmt.Printf("Параметр 'app.array' изменился с %v на %v\n", oldValue, newValue)
	})

	// Блокируем выполнение
	select {}
}

package main

import (
	"fmt"
	"os"
	"strings"
)

func mainBroadcast() {
	fmt.Println("🎯 Демонстрация различных подходов к Broadcast вещанию в Go")
	fmt.Println(strings.Repeat("=", 60))

	// Проверяем аргументы: os.Args[0] - имя программы, os.Args[1] - "broadcast", os.Args[2] - тип
	if len(os.Args) > 2 {
		switch os.Args[2] {
		case "simple":
			ExampleSimpleBroadcaster()
		case "cond":
			ExampleCondBroadcaster()
		case "typed":
			ExampleTypedBroadcaster()
		default:
			showHelp()
		}
	} else {
		// Запускаем все примеры
		ExampleSimpleBroadcaster()
		ExampleCondBroadcaster()
		ExampleTypedBroadcaster()

		fmt.Println("\n🎉 Все примеры выполнены успешно!")
	}
}

func showHelp() {
	fmt.Println(`
Использование: go run *.go broadcast [тип]

Доступные типы:
  simple  - Простой broadcaster через горутину и каналы
  cond    - Broadcaster через sync.Cond (более эффективный)
  typed   - Типизированный broadcaster с generics

Без параметров - запускает все примеры подряд.

Примеры запуска:
  go run *.go broadcast simple
  go run *.go broadcast cond  
  go run *.go broadcast typed
  go run *.go broadcast
`)
}

package main

import (
	"fmt"
	"os"
)

func main() {
	// Проверяем аргументы командной строки для broadcast примеров
	if len(os.Args) > 1 && os.Args[1] == "broadcast" {
		// Запускаем broadcast примеры
		mainBroadcast()
		return
	}

	fmt.Println("🚀 Go Secrets - Коллекция полезных паттернов и примеров Go")
	fmt.Println("===========================================================")

	fmt.Println("\n📢 Доступные демонстрации:")
	fmt.Println("1. Broadcast вещание с помощью каналов")
	fmt.Println("   Запуск: go run *.go broadcast [тип]")
	fmt.Println("   Типы: simple, cond, typed или пусто для всех примеров")

	fmt.Println("\n💡 Примеры использования:")
	fmt.Println("   go run *.go broadcast        # Все примеры broadcast")
	fmt.Println("   go run *.go broadcast simple # Простой broadcaster")
	fmt.Println("   go run *.go broadcast cond   # Broadcaster с sync.Cond")
	fmt.Println("   go run *.go broadcast typed  # Типизированный broadcaster")

	fmt.Println("\n✅ Выберите нужную демонстрацию и запустите соответствующую команду!")
}

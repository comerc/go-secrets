package main

import (
	"fmt"
	"sync"
	"time"
)

// SimpleBroadcaster - простая реализация broadcast через горутину
type SimpleBroadcaster struct {
	input          chan interface{}
	listeners      []chan interface{}
	addListener    chan chan interface{}
	removeListener chan chan interface{}
	mu             sync.RWMutex
	closed         bool
}

// NewSimpleBroadcaster создает новый broadcaster
func NewSimpleBroadcaster() *SimpleBroadcaster {
	b := &SimpleBroadcaster{
		input:          make(chan interface{}),
		addListener:    make(chan chan interface{}),
		removeListener: make(chan chan interface{}),
	}

	go b.run()
	return b
}

// run - основной цикл broadcaster'а
func (b *SimpleBroadcaster) run() {
	defer func() {
		// Закрываем все listener каналы при завершении
		b.mu.Lock()
		for _, listener := range b.listeners {
			close(listener)
		}
		b.listeners = nil
		b.mu.Unlock()
	}()

	for {
		select {
		case msg, ok := <-b.input:
			if !ok {
				// Входной канал закрыт
				return
			}
			// Отправляем сообщение всем слушателям
			b.broadcast(msg)

		case listener := <-b.addListener:
			b.mu.Lock()
			b.listeners = append(b.listeners, listener)
			b.mu.Unlock()

		case listener := <-b.removeListener:
			b.mu.Lock()
			b.removeListenerUnsafe(listener)
			b.mu.Unlock()
		}
	}
}

// broadcast отправляет сообщение всем слушателям
func (b *SimpleBroadcaster) broadcast(msg interface{}) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for i, listener := range b.listeners {
		select {
		case listener <- msg:
			// Сообщение отправлено успешно
		default:
			// Канал заблокирован, удаляем этого слушателя
			fmt.Printf("Listener %d заблокирован, удаляем\n", i)
			go func(l chan interface{}) {
				close(l)
			}(listener)
			// Удаление слушателя будет выполнено асинхронно
		}
	}
}

// removeListenerUnsafe удаляет слушателя (должен вызываться под мьютексом)
func (b *SimpleBroadcaster) removeListenerUnsafe(target chan interface{}) {
	for i, listener := range b.listeners {
		if listener == target {
			b.listeners = append(b.listeners[:i], b.listeners[i+1:]...)
			close(target)
			break
		}
	}
}

// Send отправляет сообщение для broadcast
func (b *SimpleBroadcaster) Send(msg interface{}) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return
	}

	select {
	case b.input <- msg:
	default:
		fmt.Println("Входной канал заполнен, сообщение пропущено")
	}
}

// Subscribe добавляет нового слушателя
func (b *SimpleBroadcaster) Subscribe() chan interface{} {
	listener := make(chan interface{}, 10) // Буферизованный канал

	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		close(listener)
		return listener
	}
	b.mu.RUnlock()

	select {
	case b.addListener <- listener:
	default:
		close(listener)
	}

	return listener
}

// Unsubscribe удаляет слушателя
func (b *SimpleBroadcaster) Unsubscribe(listener chan interface{}) {
	select {
	case b.removeListener <- listener:
	default:
		// Канал управления заполнен
	}
}

// Close закрывает broadcaster
func (b *SimpleBroadcaster) Close() {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return
	}
	b.closed = true
	close(b.input)
	b.mu.Unlock()
}

// Пример использования SimpleBroadcaster
func ExampleSimpleBroadcaster() {
	fmt.Println("=== Пример SimpleBroadcaster ===")

	broadcaster := NewSimpleBroadcaster()
	defer broadcaster.Close()

	// Создаем нескольких слушателей
	listener1 := broadcaster.Subscribe()
	listener2 := broadcaster.Subscribe()
	listener3 := broadcaster.Subscribe()

	// Горутина для чтения от первого слушателя
	go func() {
		for msg := range listener1 {
			fmt.Printf("Listener 1 получил: %v\n", msg)
		}
		fmt.Println("Listener 1 завершен")
	}()

	// Горутина для чтения от второго слушателя
	go func() {
		for msg := range listener2 {
			fmt.Printf("Listener 2 получил: %v\n", msg)
		}
		fmt.Println("Listener 2 завершен")
	}()

	// Горутина для чтения от третьего слушателя (медленный)
	go func() {
		for msg := range listener3 {
			fmt.Printf("Listener 3 (медленный) получил: %v\n", msg)
			time.Sleep(100 * time.Millisecond) // Имитация медленной обработки
		}
		fmt.Println("Listener 3 завершен")
	}()

	// Отправляем сообщения
	for i := 0; i < 5; i++ {
		broadcaster.Send(fmt.Sprintf("Сообщение %d", i+1))
		time.Sleep(50 * time.Millisecond)
	}

	// Отписываем одного слушателя
	broadcaster.Unsubscribe(listener2)

	// Отправляем еще сообщения
	for i := 5; i < 8; i++ {
		broadcaster.Send(fmt.Sprintf("Сообщение %d", i+1))
		time.Sleep(50 * time.Millisecond)
	}

	time.Sleep(200 * time.Millisecond) // Ждем завершения обработки
}

package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CondBroadcaster - реализация broadcast через sync.Cond
type CondBroadcaster struct {
	mu        sync.RWMutex
	cond      *sync.Cond
	listeners map[int]chan interface{}
	nextID    int
	closed    bool
	lastMsg   interface{}
	hasMsg    bool
}

// NewCondBroadcaster создает новый broadcaster на основе sync.Cond
func NewCondBroadcaster() *CondBroadcaster {
	b := &CondBroadcaster{
		listeners: make(map[int]chan interface{}),
	}
	b.cond = sync.NewCond(&b.mu)
	return b
}

// Subscribe добавляет нового слушателя
func (b *CondBroadcaster) Subscribe(ctx context.Context, bufferSize int) (int, chan interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		ch := make(chan interface{})
		close(ch)
		return -1, ch
	}

	id := b.nextID
	b.nextID++

	listener := make(chan interface{}, bufferSize)
	b.listeners[id] = listener

	// Запускаем горутину для слушания
	go b.listen(ctx, id, listener)

	return id, listener
}

// listen горутина для каждого слушателя
func (b *CondBroadcaster) listen(ctx context.Context, id int, ch chan interface{}) {
	defer func() {
		b.mu.Lock()
		delete(b.listeners, id)
		close(ch)
		b.mu.Unlock()
	}()

	for {
		b.mu.Lock()

		// Ждем сообщения или закрытия
		for !b.hasMsg && !b.closed {
			select {
			case <-ctx.Done():
				b.mu.Unlock()
				return
			default:
				b.cond.Wait()
			}
		}

		if b.closed {
			b.mu.Unlock()
			return
		}

		msg := b.lastMsg
		b.mu.Unlock()

		// Отправляем сообщение в канал
		select {
		case ch <- msg:
		case <-ctx.Done():
			return
		default:
			// Канал заполнен, пропускаем сообщение
			fmt.Printf("Listener %d: канал заполнен, сообщение пропущено\n", id)
		}
	}
}

// Broadcast отправляет сообщение всем слушателям
func (b *CondBroadcaster) Broadcast(msg interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	b.lastMsg = msg
	b.hasMsg = true

	// Уведомляем всех ожидающих горутин
	b.cond.Broadcast()

	// Сбрасываем флаг после небольшой задержки
	go func() {
		time.Sleep(time.Millisecond)
		b.mu.Lock()
		b.hasMsg = false
		b.mu.Unlock()
	}()
}

// Close закрывает broadcaster
func (b *CondBroadcaster) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	b.closed = true
	b.cond.Broadcast()
}

// GetListenerCount возвращает количество активных слушателей
func (b *CondBroadcaster) GetListenerCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.listeners)
}

// Пример использования CondBroadcaster
func ExampleCondBroadcaster() {
	fmt.Println("\n=== Пример CondBroadcaster ===")

	broadcaster := NewCondBroadcaster()
	defer broadcaster.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// Создаем несколько слушателей
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(listenerNum int) {
			defer wg.Done()

			id, ch := broadcaster.Subscribe(ctx, 5)
			fmt.Printf("Listener %d подписался с ID %d\n", listenerNum, id)

			for msg := range ch {
				fmt.Printf("Listener %d получил: %v\n", listenerNum, msg)
				if listenerNum == 2 {
					// Имитация медленной обработки для третьего слушателя
					time.Sleep(20 * time.Millisecond)
				}
			}

			fmt.Printf("Listener %d завершен\n", listenerNum)
		}(i)
	}

	// Даем время на подписку
	time.Sleep(50 * time.Millisecond)

	// Отправляем сообщения
	for i := 0; i < 10; i++ {
		broadcaster.Broadcast(fmt.Sprintf("Сообщение %d", i+1))
		fmt.Printf("Активных слушателей: %d\n", broadcaster.GetListenerCount())
		time.Sleep(30 * time.Millisecond)
	}

	// Отменяем контекст (это закроет слушателей)
	cancel()
	wg.Wait()

	fmt.Printf("Финальное количество слушателей: %d\n", broadcaster.GetListenerCount())
}

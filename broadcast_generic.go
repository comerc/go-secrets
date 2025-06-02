package main

import (
	"fmt"
	"sync"
	"time"
)

// TypedBroadcaster - типизированный broadcaster с generics
type TypedBroadcaster[T any] struct {
	mu        sync.RWMutex
	listeners map[string]chan T
	closed    bool
}

// NewTypedBroadcaster создает новый типизированный broadcaster
func NewTypedBroadcaster[T any]() *TypedBroadcaster[T] {
	return &TypedBroadcaster[T]{
		listeners: make(map[string]chan T),
	}
}

// Subscribe добавляет нового слушателя с уникальным ID
func (b *TypedBroadcaster[T]) Subscribe(id string, bufferSize int) (chan T, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil, fmt.Errorf("broadcaster закрыт")
	}

	if _, exists := b.listeners[id]; exists {
		return nil, fmt.Errorf("слушатель с ID %s уже существует", id)
	}

	listener := make(chan T, bufferSize)
	b.listeners[id] = listener

	return listener, nil
}

// Unsubscribe удаляет слушателя
func (b *TypedBroadcaster[T]) Unsubscribe(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if listener, exists := b.listeners[id]; exists {
		close(listener)
		delete(b.listeners, id)
	}
}

// Broadcast отправляет сообщение всем слушателям
func (b *TypedBroadcaster[T]) Broadcast(msg T) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return
	}

	for id, listener := range b.listeners {
		select {
		case listener <- msg:
			// Сообщение отправлено успешно
		default:
			// Канал заполнен
			fmt.Printf("Listener %s: канал заполнен, сообщение пропущено\n", id)
		}
	}
}

// BroadcastWithTimeout отправляет сообщение с таймаутом
func (b *TypedBroadcaster[T]) BroadcastWithTimeout(msg T, timeout time.Duration) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return
	}

	for id, listener := range b.listeners {
		select {
		case listener <- msg:
			// Сообщение отправлено успешно
		case <-time.After(timeout):
			fmt.Printf("Listener %s: таймаут при отправке сообщения\n", id)
		}
	}
}

// Close закрывает broadcaster и все каналы слушателей
func (b *TypedBroadcaster[T]) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	b.closed = true

	for id, listener := range b.listeners {
		close(listener)
		fmt.Printf("Закрыт канал слушателя %s\n", id)
	}

	b.listeners = nil
}

// GetListenerIDs возвращает список ID всех активных слушателей
func (b *TypedBroadcaster[T]) GetListenerIDs() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	ids := make([]string, 0, len(b.listeners))
	for id := range b.listeners {
		ids = append(ids, id)
	}

	return ids
}

// BroadcastToSpecific отправляет сообщение только указанным слушателям
func (b *TypedBroadcaster[T]) BroadcastToSpecific(msg T, targetIDs []string) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return
	}

	for _, id := range targetIDs {
		if listener, exists := b.listeners[id]; exists {
			select {
			case listener <- msg:
				fmt.Printf("Сообщение отправлено слушателю %s\n", id)
			default:
				fmt.Printf("Listener %s: канал заполнен\n", id)
			}
		}
	}
}

// Message - пример типа сообщения
type Message struct {
	ID      int       `json:"id"`
	Content string    `json:"content"`
	From    string    `json:"from"`
	Time    time.Time `json:"time"`
}

// String реализует Stringer интерфейс для Message
func (m Message) String() string {
	return fmt.Sprintf("[%d] %s от %s в %s", m.ID, m.Content, m.From, m.Time.Format("15:04:05"))
}

// Пример использования TypedBroadcaster
func ExampleTypedBroadcaster() {
	fmt.Println("\n=== Пример TypedBroadcaster ===")

	// Создаем типизированный broadcaster для Message
	broadcaster := NewTypedBroadcaster[Message]()
	defer broadcaster.Close()

	// Подписываем несколько слушателей
	listener1, _ := broadcaster.Subscribe("user1", 5)
	listener2, _ := broadcaster.Subscribe("user2", 3)
	listener3, _ := broadcaster.Subscribe("admin", 10)

	var wg sync.WaitGroup

	// Запускаем горутины для чтения сообщений
	wg.Add(3)

	go func() {
		defer wg.Done()
		for msg := range listener1 {
			fmt.Printf("User1 получил: %s\n", msg)
		}
	}()

	go func() {
		defer wg.Done()
		for msg := range listener2 {
			fmt.Printf("User2 получил: %s\n", msg)
			time.Sleep(50 * time.Millisecond) // Медленная обработка
		}
	}()

	go func() {
		defer wg.Done()
		for msg := range listener3 {
			fmt.Printf("Admin получил: %s\n", msg)
		}
	}()

	// Отправляем сообщения всем
	for i := 0; i < 5; i++ {
		msg := Message{
			ID:      i + 1,
			Content: fmt.Sprintf("Общее сообщение %d", i+1),
			From:    "system",
			Time:    time.Now(),
		}
		broadcaster.Broadcast(msg)
		time.Sleep(100 * time.Millisecond)
	}

	// Отправляем сообщение только админу
	adminMsg := Message{
		ID:      100,
		Content: "Секретное сообщение для админа",
		From:    "security",
		Time:    time.Now(),
	}
	broadcaster.BroadcastToSpecific(adminMsg, []string{"admin"})

	// Показываем активных слушателей
	fmt.Printf("Активные слушатели: %v\n", broadcaster.GetListenerIDs())

	// Отписываем одного слушателя
	broadcaster.Unsubscribe("user2")

	// Отправляем еще одно сообщение
	finalMsg := Message{
		ID:      999,
		Content: "Финальное сообщение",
		From:    "system",
		Time:    time.Now(),
	}
	broadcaster.BroadcastWithTimeout(finalMsg, 50*time.Millisecond)

	time.Sleep(200 * time.Millisecond)
	wg.Wait()
}

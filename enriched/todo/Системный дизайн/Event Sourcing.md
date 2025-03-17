#EventSourcing #go #patterns #architecture #databases #distributed_systems #consistency #auditability #CQRS #eventual_consistency #microservices

# Event Sourcing

```table-of-contents
```

## Обзор Event Sourcing

Event Sourcing (источники событий) - это шаблон проектирования, при котором изменения состояния приложения сохраняются как последовательность событий. Вместо хранения только текущего состояния, каждое изменение фиксируется как отдельное неизменяемое событие. Эти события сохраняются в хронологическом порядке в хранилище событий (Event Store). Текущее состояние приложения может быть восстановлено путем "перепроигрывания" (replay) всех событий с самого начала или с определенного момента времени (снапшота).

## Детальный разбор Event Sourcing

### Ключевые компоненты

1.  **Событие (Event):** Неизменяемый объект, представляющий собой факт, произошедший в системе. Событие содержит информацию о том, *что* произошло, *когда* произошло, и, возможно, дополнительные метаданные. Примеры событий: `OrderCreated`, `ProductAddedToCart`, `PaymentReceived`. События обычно имеют уникальный идентификатор и порядковый номер в потоке событий.

2.  **Хранилище событий (Event Store):** Специализированная база данных, предназначенная для хранения последовательности событий. Event Store обеспечивает атомарные операции добавления событий и гарантирует их хронологический порядок. Примеры: EventStoreDB, Apache Kafka (с некоторыми оговорками), AWS Kinesis, а также можно использовать реляционные базы данных (PostgreSQL, MySQL) или NoSQL базы данных (MongoDB, Cassandra).

3.  **Агрегат (Aggregate):** Объект предметной области, состояние которого изменяется посредством событий. Агрегат отвечает за применение событий и изменение своего внутреннего состояния. Примеры: `Order`, `ShoppingCart`, `Customer`. Агрегат гарантирует, что все бизнес-правила соблюдаются при применении каждого события.

4.  **Проекция (Projection):** Представление данных, построенное на основе последовательности событий. Проекции могут быть оптимизированы для конкретных запросов и могут представлять собой read-модели в CQRS. Например, проекция может содержать текущее количество товаров на складе или список всех заказов клиента.

5.  **Снапшот (Snapshot):** Периодически создаваемое "мгновенное" состояние агрегата. Снапшоты используются для ускорения процесса восстановления состояния, так как позволяют начать перепроигрывание событий не с самого начала, а с момента создания снапшота.

### Процесс работы

1.  **Команда (Command):** Запрос на изменение состояния системы. Команда обрабатывается агрегатом.

2.  **Обработка команды:** Агрегат проверяет команду на валидность и, если все правила соблюдены, генерирует одно или несколько событий.

3.  **Сохранение событий:** Сгенерированные события атомарно добавляются в Event Store.

4.  **Применение событий:** Агрегат применяет события к своему внутреннему состоянию, обновляя его.

5.  **Обновление проекций:** Подписчики (subscribers) на события получают уведомления о новых событиях и обновляют соответствующие проекции.

### Преимущества

*   **Полная история изменений:** Event Sourcing предоставляет полную историю всех изменений, произошедших в системе. Это полезно для аудита, отладки и анализа данных.

*   **Возможность отката:** Можно вернуться к любому предыдущему состоянию системы, перепроиграв события до нужного момента.

*   **Отладка и воспроизведение:** Легко воспроизвести ошибки, так как можно "переиграть" последовательность событий, приведших к ошибке.

*   **Гибкость:** Можно создавать различные проекции данных, оптимизированные для разных сценариев использования.

*   **Масштабируемость:** Event Store обычно хорошо масштабируются, особенно если используются специализированные решения.

*   **Совместимость с CQRS:** Event Sourcing отлично сочетается с шаблоном CQRS (Command Query Responsibility Segregation), разделяя ответственность за изменение и чтение данных.

### Недостатки

*   **Сложность:** Event Sourcing может быть сложнее в реализации, чем традиционные подходы к хранению данных.

*   **Сложность запросов:** Запросы к данным могут быть более сложными, так как данные распределены по множеству событий.

*   **Eventual Consistency:** Обновление проекций происходит асинхронно, поэтому система может находиться в состоянии [[Eventual Consistency]].

*   **Обработка ошибок:** Обработка ошибок в асинхронной системе может быть сложной.

*   **Версионирование событий:** С развитием системы может потребоваться версионирование событий, чтобы обеспечить обратную совместимость.

### Примеры использования

*   **Банковские системы:** Отслеживание всех транзакций и операций со счетами.

*   **Системы электронной коммерции:** Отслеживание заказов, изменений в корзине, статусов доставки.

*   **Системы управления контентом:** Отслеживание версий документов и изменений в контенте.

*   **Игровые приложения:** Отслеживание действий игроков и состояния игры.

*   **IoT (Интернет вещей):** Обработка потоков данных от датчиков и устройств.

## Пример реализации на Go

Рассмотрим упрощенный пример реализации Event Sourcing для системы управления заказами.

```go
package main

import (
	"fmt"
	"log"
	"time"
)

// EventType определяет тип события.
type EventType string

const (
	OrderCreatedEventType   EventType = "OrderCreated"
	OrderShippedEventType   EventType = "OrderShipped"
	OrderCancelledEventType EventType = "OrderCancelled"
)

// Event представляет собой событие в системе.
type Event struct {
	ID        string
	Type      EventType
	Data      interface{}
	Timestamp time.Time
	Version   int
}

// OrderCreatedEventData содержит данные для события OrderCreated.
type OrderCreatedEventData struct {
	OrderID    string
	CustomerID string
	OrderDate  time.Time
}

// OrderShippedEventData содержит данные для события OrderShipped.
type OrderShippedEventData struct {
	ShippingDate time.Time
}

type OrderCancelledEventData struct {
    Reason string
}

// Order - агрегат, представляющий заказ.
type Order struct {
	ID         string
	CustomerID string
	OrderDate  time.Time
	Shipped    bool
    Cancelled  bool
    Reason     string
	Version    int
}

// EventStore - интерфейс для хранилища событий.
type EventStore interface {
	SaveEvents(aggregateID string, events []Event) error
	GetEvents(aggregateID string) ([]Event, error)
}

// InMemoryEventStore - простая реализация EventStore в памяти.
type InMemoryEventStore struct {
	events map[string][]Event
}

// NewInMemoryEventStore создает новый InMemoryEventStore.
func NewInMemoryEventStore() *InMemoryEventStore {
	return &InMemoryEventStore{
		events: make(map[string][]Event),
	}
}

// SaveEvents сохраняет события в хранилище.
func (s *InMemoryEventStore) SaveEvents(aggregateID string, events []Event) error {
	if _, ok := s.events[aggregateID]; !ok {
		s.events[aggregateID] = []Event{}
	}
	s.events[aggregateID] = append(s.events[aggregateID], events...)
	return nil
}

// GetEvents возвращает события для заданного агрегата.
func (s *InMemoryEventStore) GetEvents(aggregateID string) ([]Event, error) {
	if events, ok := s.events[aggregateID]; ok {
		return events, nil
	}
	return []Event{}, nil
}

// ApplyEvent применяет событие к агрегату Order.
func (o *Order) ApplyEvent(event Event) {
	o.Version = event.Version
	switch event.Type {
	case OrderCreatedEventType:
		data := event.Data.(OrderCreatedEventData)
		o.ID = data.OrderID
		o.CustomerID = data.CustomerID
		o.OrderDate = data.OrderDate
	case OrderShippedEventType:
		o.Shipped = true
    case OrderCancelledEventType:
        data := event.Data.(OrderCancelledEventData)
        o.Cancelled = true
        o.Reason = data.Reason
	}
}

// CreateOrder создает новый заказ.
func CreateOrder(store EventStore, orderID, customerID string) (*Order, error) {
	order := &Order{}
	event := Event{
		ID:        generateEventID(), // Функция для генерации ID события (не показана)
		Type:      OrderCreatedEventType,
		Data:      OrderCreatedEventData{OrderID: orderID, CustomerID: customerID, OrderDate: time.Now()},
		Timestamp: time.Now(),
		Version:   1,
	}

	err := store.SaveEvents(orderID, []Event{event})
	if err != nil {
		return nil, err
	}

	order.ApplyEvent(event)
	return order, nil
}

// ShipOrder отмечает заказ как доставленный.
func ShipOrder(store EventStore, orderID string) (*Order, error) {
	order, err := reconstructOrder(store, orderID) // Восстанавливаем состояние заказа
	if err != nil {
		return nil, err
	}

    if order.Cancelled {
        return nil, fmt.Errorf("cannot ship a cancelled order")
    }

	if order.Shipped {
		return order, nil // Already shipped
	}

	event := Event{
		ID:        generateEventID(),
		Type:      OrderShippedEventType,
		Data:      OrderShippedEventData{ShippingDate: time.Now()},
		Timestamp: time.Now(),
		Version:   order.Version + 1,
	}

	err = store.SaveEvents(orderID, []Event{event})
	if err != nil {
		return nil, err
	}
	order.ApplyEvent(event)
	return order, nil
}

// CancelOrder отменяет заказ
func CancelOrder(store EventStore, orderID string, reason string) (*Order, error) {
    order, err := reconstructOrder(store, orderID)
    if err != nil {
        return nil, err
    }

    if order.Shipped {
        return nil, fmt.Errorf("cannot cancel a shipped order")
    }

     if order.Cancelled {
        return order, nil // Already cancelled
    }

    event := Event{
        ID:        generateEventID(),
        Type:      OrderCancelledEventType,
        Data:      OrderCancelledEventData{Reason: reason},
        Timestamp: time.Now(),
        Version:   order.Version + 1,
    }

    err = store.SaveEvents(orderID, []Event{event})
    if err != nil {
       return nil, err
    }
    order.ApplyEvent(event)
    return order, nil
}

// reconstructOrder восстанавливает состояние заказа из событий.
func reconstructOrder(store EventStore, orderID string) (*Order, error) {
	events, err := store.GetEvents(orderID)
	if err != nil {
		return nil, err
	}

	order := &Order{}
	for _, event := range events {
		order.ApplyEvent(event)
	}
	return order, nil
}

func generateEventID() string {
    // Placeholder for a real event ID generator
    return fmt.Sprintf("%d", time.Now().UnixNano())
}

func main() {
	store := NewInMemoryEventStore()

	// Создание заказа
	order, err := CreateOrder(store, "order-123", "customer-456")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created Order: %+v\n", order)

	// Отправка заказа
	order, err = ShipOrder(store, "order-123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Shipped Order: %+v\n", order)

    // Попытка отменить отправленный заказ
    _, err = CancelOrder(store, "order-123", "Customer request")
    if err != nil {
        fmt.Printf("Expected error: %v\n", err) // Ожидаемая ошибка
    }

    // Создание еще заказа
    order2, err := CreateOrder(store, "order-124", "customer-457")
    if err != nil {
        log.Fatal(err)
    }
	fmt.Printf("Created Order2: %+v\n", order2)

    // Отмена заказа
    order2, err = CancelOrder(store, "order-124", "Customer request")
    if err != nil {
       log.Fatal(err)
    }
    fmt.Printf("Cancelled Order: %+v\n", order2)

	// Восстановление заказа из хранилища
	reconstructedOrder, err := reconstructOrder(store, "order-123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Reconstructed Order: %+v\n", reconstructedOrder)

	reconstructedOrder2, err := reconstructOrder(store, "order-124")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Reconstructed Order2: %+v\n", reconstructedOrder2)
}
```

В этом примере:

*   Определены структуры `Event`, `OrderCreatedEventData`, `OrderShippedEventData` и `OrderCancelledEventData` для представления событий и их данных.
*   Структура `Order` представляет собой агрегат заказа.
*   Интерфейс `EventStore` определяет методы для сохранения и получения событий.
*   `InMemoryEventStore` - простая реализация `EventStore`, хранящая события в памяти.
*   Функция `ApplyEvent` применяет событие к агрегату `Order`, изменяя его состояние.
*   Функции `CreateOrder`, `ShipOrder` и `CancelOrder` создают и изменяют заказы, генерируя соответствующие события.
*   Функция `reconstructOrder` восстанавливает состояние заказа из последовательности событий.
*  Добавлен `Version` в `Event` и `Order`, для контроля изменений.
*  Добавлена проверка, что нельзя отменить отправленный заказ и нельзя отправить отмененный.

Этот пример демонстрирует базовые принципы Event Sourcing. В реальном приложении потребуется более сложная реализация, включающая обработку ошибок, версионирование событий, снапшоты, проекции и, возможно, использование специализированного хранилища событий.

## Заключение

Event Sourcing - мощный шаблон проектирования, который может быть полезен во многих ситуациях. Однако он требует тщательного планирования и реализации. Важно понимать его преимущества и недостатки, прежде чем применять его в своем проекте.

```old
Event Sourcing
```
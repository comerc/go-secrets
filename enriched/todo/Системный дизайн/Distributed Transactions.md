#DistributedTransactions #transactions #databases #consistency #distributed_systems #ACID #2PC #3PC #Saga #concurrency

# Распределенные транзакции

```table-of-contents
```

## Введение в распределенные транзакции

Распределенные транзакции — это транзакции, которые затрагивают несколько сетевых узлов, обычно баз данных. В отличие от локальных транзакций, которые выполняются в пределах одной базы данных, распределенные транзакции обеспечивают целостность и согласованность данных в распределенной среде. Это критически важно в современных приложениях, где данные часто хранятся и обрабатываются на множестве серверов, облачных сервисов или микросервисов.

## Проблемы распределенных систем

Распределенные системы сталкиваются с рядом проблем, которые делают реализацию транзакций сложнее, чем в локальных системах. Ключевые из них:

1.  **Гетерогенность:** Узлы могут использовать разные операционные системы, базы данных или протоколы связи.
2.  **Сетевые задержки и сбои:** Сеть, соединяющая узлы, может быть ненадежной, с переменной задержкой и возможностью потери пакетов.
3.  **Конкурентный доступ:** Множество клиентов могут одновременно пытаться получить доступ к одним и тем же данным на разных узлах.
4.  **Частичные сбои:** Один или несколько узлов могут выйти из строя, в то время как другие продолжают работать.
5.  **Сложность синхронизации:** Обеспечение согласованности данных между узлами требует сложной координации и синхронизации.

## Свойства ACID

ACID (Atomicity, Consistency, Isolation, Durability) — это набор свойств, гарантирующих надежность транзакций в базах данных. В контексте распределенных транзакций обеспечение ACID становится значительно сложнее.

*   **Атомарность (Atomicity):** Гарантирует, что транзакция выполняется полностью или не выполняется вовсе. В распределенной среде это означает, что все участвующие узлы должны либо зафиксировать изменения, либо откатить их.
*   **Согласованность (Consistency):** Гарантирует, что транзакция переводит систему из одного согласованного состояния в другое. Это подразумевает соблюдение всех правил и ограничений целостности данных, определенных в системе.
*   **Изоляция (Isolation):** Гарантирует, что параллельно выполняемые транзакции не влияют друг на друга. В распределенной среде это может потребовать использования сложных механизмов блокировки и управления параллелизмом.
*   **Долговечность (Durability):** Гарантирует, что после фиксации транзакции изменения сохраняются даже в случае сбоя системы. В распределенной среде это может потребовать репликации данных и механизмов восстановления после сбоев.

## Протоколы распределенных транзакций

Существует несколько протоколов, используемых для реализации распределенных транзакций. Наиболее распространенными являются двухфазный коммит (2PC) и трехфазный коммит (3PC), а также паттерн Saga.

### Двухфазный коммит (2PC)

Двухфазный коммит (Two-Phase Commit, 2PC) — это протокол, который координирует фиксацию или откат распределенной транзакции между несколькими участниками (ресурсами).

**Фазы 2PC:**

1.  **Фаза подготовки (Prepare Phase):**
    *   Координатор отправляет сообщение `Prepare` всем участникам.
    *   Каждый участник, получив сообщение `Prepare`, выполняет все необходимые операции для подготовки к фиксации транзакции (например, запись изменений в журнал, блокировка ресурсов).
    *   Если участник готов зафиксировать транзакцию, он отправляет координатору сообщение `Ready`.
    *   Если участник не может зафиксировать транзакцию (например, из-за нарушения ограничений целостности), он отправляет координатору сообщение `Abort`.

2.  **Фаза фиксации (Commit Phase):**
    *   Если координатор получил сообщения `Ready` от всех участников, он принимает решение о фиксации транзакции и отправляет сообщение `Commit` всем участникам.
    *   Если координатор получил хотя бы одно сообщение `Abort` или не получил ответа от какого-либо участника в течение заданного времени (тайм-аут), он принимает решение об откате транзакции и отправляет сообщение `Abort` всем участникам.
    *   Каждый участник, получив сообщение `Commit`, фиксирует транзакцию (применяет изменения) и отправляет координатору сообщение `Ack`.
    *   Каждый участник, получив сообщение `Abort`, откатывает транзакцию (отменяет изменения) и отправляет координатору сообщение `Ack`.
    *   Координатор завершает транзакцию после получения сообщений `Ack` от всех участников.

**Преимущества 2PC:**

*   Относительная простота реализации.
*   Гарантирует атомарность распределенной транзакции.

**Недостатки 2PC:**

*   **Блокирующий протокол:** Если координатор выходит из строя во время фазы подготовки, участники могут остаться заблокированными, ожидая решения координатора. Это может привести к недоступности ресурсов.
*   **Единая точка отказа:** Координатор является единой точкой отказа. Если он выходит из строя, транзакция не может быть завершена.
*   **Не обрабатывает сбои участников после отправки `Ready`:** Если участник выходит из строя после отправки сообщения `Ready`, но до получения сообщения `Commit` или `Abort`, он может оказаться в несогласованном состоянии.

### Трехфазный коммит (3PC)

Трехфазный коммит (Three-Phase Commit, 3PC) — это протокол распределенных транзакций, разработанный для устранения проблемы блокировки, присущей 2PC. 3PC добавляет дополнительную фазу между фазами подготовки и фиксации в 2PC.

**Фазы 3PC:**

1.  **Фаза запроса на фиксацию (CanCommit Phase):**
    *   Эта фаза аналогична фазе подготовки в 2PC. Координатор отправляет сообщение `CanCommit` всем участникам.
    *   Участники отвечают `Yes` или `No`, указывая, готовы ли они к фиксации.

2.  **Фаза предварительной фиксации (PreCommit Phase):**
    *   Если все участники ответили `Yes` на запрос `CanCommit`, координатор отправляет сообщение `PreCommit` всем участникам.
    *   Участники, получив `PreCommit`, подтверждают получение сообщения, но еще не фиксируют транзакцию. Это гарантирует, что все участники знают о решении координатора, прежде чем кто-либо из них начнет фиксацию.
    *   Если координатор получает подтверждения от всех участников, он переходит к фазе фиксации.

3.  **Фаза фиксации (Commit Phase):**
    *   Координатор отправляет сообщение `Commit` всем участникам.
    *   Участники фиксируют транзакцию и отправляют подтверждение координатору.

**Преимущества 3PC:**

*   **Неблокирующий протокол:** В отличие от 2PC, 3PC позволяет участникам восстанавливаться после сбоя координатора и завершать транзакцию.
*   Уменьшает вероятность блокировки ресурсов.

**Недостатки 3PC:**

*   Более сложный в реализации, чем 2PC.
*   Дополнительная фаза увеличивает задержку транзакции.
*   Все еще может привести к несогласованности в некоторых сценариях сбоев (например, при разделении сети).

### Паттерн Saga

Saga — это паттерн управления распределенными транзакциями, который предлагает альтернативу 2PC и 3PC. Вместо атомарной фиксации всех изменений Saga разбивает транзакцию на последовательность локальных транзакций, выполняемых каждым участником. Каждая локальная транзакция обновляет данные в своей базе данных и публикует событие, которое запускает следующую локальную транзакцию в последовательности.

**Основные компоненты Saga:**

*   **Локальные транзакции:** Каждая локальная транзакция выполняется атомарно в пределах одного участника.
*   **Компенсирующие транзакции:** Для каждой локальной транзакции определяется компенсирующая транзакция, которая отменяет ее изменения.
*   **Оркестратор (Saga Execution Coordinator, SEC):** Необязательный компонент, который управляет выполнением Saga. Он отслеживает состояние каждой локальной транзакции и запускает компенсирующие транзакции в случае сбоя.

**Типы Saga:**

*   **Оркестровка (Orchestration-based Saga):** Централизованный оркестратор управляет выполнением Saga. Он отправляет команды участникам и обрабатывает события от них.
*   **Хореография (Choreography-based Saga):** Участники обмениваются событиями напрямую, без централизованного оркестратора. Каждый участник сам решает, когда выполнять свою локальную транзакцию и компенсирующую транзакцию на основе полученных событий.

**Пример Saga (Оркестровка):**

Предположим, у нас есть сервис бронирования билетов, который включает в себя три микросервиса:

1.  **Order Service:** Создает заказ.
2.  **Payment Service:** Обрабатывает платеж.
3.  **Inventory Service:** Резервирует билеты.

Saga для бронирования билета может выглядеть следующим образом:

1.  **Order Service** создает заказ в состоянии `Pending`.
2.  **Order Service** отправляет команду `ProcessPayment` в **Payment Service**.
3.  **Payment Service** обрабатывает платеж. Если платеж успешен, он публикует событие `PaymentProcessed`.
4.  **Order Service**, получив событие `PaymentProcessed`, отправляет команду `ReserveTickets` в **Inventory Service**.
5.  **Inventory Service** резервирует билеты. Если билеты доступны, он публикует событие `TicketsReserved`.
6.  **Order Service**, получив событие `TicketsReserved`, изменяет состояние заказа на `Confirmed`.

Если на каком-либо этапе происходит сбой (например, платеж не проходит), оркестратор запускает компенсирующие транзакции:

1.  Если **Payment Service** не может обработать платеж, он публикует событие `PaymentFailed`.
2.  **Order Service**, получив событие `PaymentFailed`, запускает компенсирующую транзакцию, которая изменяет состояние заказа на `Cancelled`.

**Преимущества Saga:**

*   **Слабая связанность:** Участники Saga не блокируют друг друга.
*   **Высокая доступность:** Saga может продолжать работать даже при сбое отдельных участников.
*   Подходит для длительных транзакций.

**Недостатки Saga:**

*   **Отсутствие изоляции:** Промежуточные состояния Saga видны другим транзакциям.
*   **Сложность реализации компенсирующих транзакций:** Разработка компенсирующих транзакций может быть сложной и требовать тщательного проектирования.
*   **Потенциальные проблемы с согласованностью:** Saga гарантирует только eventual consistency (согласованность в конечном итоге).

## Пример реализации Saga на Go (с использованием оркестровки)

```go
package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Event types
const (
	OrderCreated     = "OrderCreated"
	PaymentProcessed = "PaymentProcessed"
	TicketsReserved  = "TicketsReserved"
	PaymentFailed    = "PaymentFailed"
	TicketsNotAvailable = "TicketsNotAvailable"
	OrderCancelled ="OrderCancelled"

)

// Command types
const (
	ProcessPayment = "ProcessPayment"
	ReserveTickets = "ReserveTickets"
	CancelOrder = "CancelOrder"
)

// Event represents a domain event.
type Event struct {
	Type    string
	Payload interface{}
}

// Command represents a command to be executed by a service.
type Command struct {
	Type    string
	Payload interface{}
}

// Order represents an order.
type Order struct {
	ID     string
	Status string
	// ... other order details
}

// PaymentService simulates the payment service.
type PaymentService struct {
	eventBus chan Event
}

// NewPaymentService creates a new PaymentService.
func NewPaymentService(eventBus chan Event) *PaymentService {
	return &PaymentService{eventBus: eventBus}
}

// HandleCommand handles commands for the PaymentService.
func (ps *PaymentService) HandleCommand(command Command) {
	switch command.Type {
	case ProcessPayment:
		order := command.Payload.(Order)
		// Simulate payment processing
		time.Sleep(1 * time.Second)
		// Assume payment is successful for this example
		ps.eventBus <- Event{Type: PaymentProcessed, Payload: order}

		//For simulate fail
		//ps.eventBus <- Event{Type: PaymentFailed, Payload: order}

	}
}

// InventoryService simulates the inventory service.
type InventoryService struct {
	eventBus chan Event
}

// NewInventoryService creates a new InventoryService.
func NewInventoryService(eventBus chan Event) *InventoryService {
	return &InventoryService{eventBus: eventBus}
}

// HandleCommand handles commands for the InventoryService.
func (is *InventoryService) HandleCommand(command Command) {
	switch command.Type {
	case ReserveTickets:
		order := command.Payload.(Order)
		// Simulate ticket reservation
		time.Sleep(1 * time.Second)
		// Assume tickets are available for this example
		is.eventBus <- Event{Type: TicketsReserved, Payload: order}

		// For simulate fail
		//is.eventBus <- Event{Type: TicketsNotAvailable, Payload: order}
	}
}

// OrderService manages the order lifecycle.
type OrderService struct {
	eventBus    chan Event
	commandBus  chan Command
	orders      map[string]Order // In-memory storage for simplicity
	ordersMutex sync.RWMutex
}

// NewOrderService creates a new OrderService.
func NewOrderService(eventBus chan Event, commandBus chan Command) *OrderService {
	return &OrderService{
		eventBus:   eventBus,
		commandBus: commandBus,
		orders:     make(map[string]Order),
	}
}

// CreateOrder creates a new order.
func (os *OrderService) CreateOrder(orderID string) {
	order := Order{ID: orderID, Status: "Pending"}
	os.ordersMutex.Lock()
	os.orders[orderID] = order
	os.ordersMutex.Unlock()
	os.eventBus <- Event{Type: OrderCreated, Payload: order}
}


func (os *OrderService) GetOrder(orderId string) (Order, bool)  {
	os.ordersMutex.RLock()
	defer os.ordersMutex.RUnlock()
	order, ok := os.orders[orderId]
	return order, ok
}

// HandleEvent handles events from other services.
func (os *OrderService) HandleEvent(event Event) {
	switch event.Type {
	case OrderCreated:
		order := event.Payload.(Order)
		// Send command to PaymentService
		os.commandBus <- Command{Type: ProcessPayment, Payload: order}
	case PaymentProcessed:
		order := event.Payload.(Order)
		// Send command to InventoryService
		os.commandBus <- Command{Type: ReserveTickets, Payload: order}
	case TicketsReserved:
		order := event.Payload.(Order)
		// Update order status to Confirmed
		os.ordersMutex.Lock()
		order.Status = "Confirmed"
		os.orders[order.ID] = order
		os.ordersMutex.Unlock()
		fmt.Printf("Order %s confirmed\n", order.ID)
	case PaymentFailed:
		order := event.Payload.(Order)
		os.commandBus <- Command{Type: CancelOrder, Payload: order}
	case TicketsNotAvailable:
		order := event.Payload.(Order)
		os.commandBus <- Command{Type: CancelOrder, Payload: order}
	case OrderCancelled:
		order := event.Payload.(Order)
		os.ordersMutex.Lock()
		order.Status = "Cancelled"
		os.orders[order.ID] = order
		os.ordersMutex.Unlock()
		fmt.Printf("Order %s cancelled\n", order.ID)

	}
}

func (os *OrderService) CancelOrder(order Order) {
	os.ordersMutex.Lock()
	order.Status = "Cancelled"
	os.orders[order.ID] = order
	os.ordersMutex.Unlock()
	os.eventBus <- Event{Type: OrderCancelled, Payload: order}
}

// SagaOrchestrator coordinates the Saga execution.
type SagaOrchestrator struct {
	orderService   *OrderService
	paymentService *PaymentService
	inventoryService *InventoryService
	eventBus       chan Event
	commandBus     chan Command
}

// NewSagaOrchestrator creates a new SagaOrchestrator.
func NewSagaOrchestrator(orderService *OrderService, paymentService *PaymentService, inventoryService *InventoryService, eventBus chan Event, commandBus chan Command) *SagaOrchestrator {
	return &SagaOrchestrator{
		orderService:   orderService,
		paymentService: paymentService,
		inventoryService: inventoryService,
		eventBus:       eventBus,
		commandBus:     commandBus,
	}
}

// Run starts the Saga orchestrator.
func (so *SagaOrchestrator) Run() {
	for {
		select {
		case event := <-so.eventBus:
			log.Printf("Orchestrator received event: %s\n", event.Type)
			so.orderService.HandleEvent(event)
		case command := <-so.commandBus:
			log.Printf("Orchestrator received command: %s\n", command.Type)
			switch command.Type {
			case ProcessPayment:
				so.paymentService.HandleCommand(command)
			case ReserveTickets:
				so.inventoryService.HandleCommand(command)
			case CancelOrder:
				order := command.Payload.(Order)
				so.orderService.CancelOrder(order)
			}
		}
	}
}

func main() {
	eventBus := make(chan Event)
	commandBus := make(chan Command)

	orderService := NewOrderService(eventBus, commandBus)
	paymentService := NewPaymentService(eventBus)
	inventoryService := NewInventoryService(eventBus)

	orchestrator := NewSagaOrchestrator(orderService, paymentService, inventoryService, eventBus, commandBus)

	go orchestrator.Run()

	// Create a new order
	orderService.CreateOrder("123")

	//Simulate some time for the Saga to complete.
	time.Sleep(5 * time.Second)

	// Check final order
	order, _ := orderService.GetOrder("123")

	fmt.Printf("Order final status: %+v\n", order)
}

```

**Описание примера:**

1.  **Определение типов:** Определяются типы для событий (`Event`), команд (`Command`) и заказа (`Order`).
2.  **Сервисы:** Создаются три сервиса: `PaymentService`, `InventoryService` и `OrderService`. Каждый сервис имеет метод `HandleCommand`, который обрабатывает команды, специфичные для этого сервиса.
3.  **OrderService:** `OrderService` управляет жизненным циклом заказа. Он создает заказы, обрабатывает события от других сервисов и отправляет команды другим сервисам.  Он также имеет in-memory хранилище заказов (`orders`) для простоты.
4.  **SagaOrchestrator:** `SagaOrchestrator` координирует выполнение Saga. Он прослушивает каналы событий (`eventBus`) и команд (`commandBus`) и вызывает соответствующие методы сервисов.
5.  **Каналы:** Каналы `eventBus` и `commandBus` используются для асинхронного обмена сообщениями между сервисами и оркестратором.
6.  **Запуск оркестратора:** Оркестратор запускается в отдельной горутине (`go orchestrator.Run()`).
7.  **Создание заказа:** `orderService.CreateOrder("123")` создает новый заказ с ID "123".
8. **Цикл обработки сообщений:** В методе `Run()` оркестратор непрерывно ожидает поступления событий или команд из соответствующих каналов. При получении сообщения, он определяет его тип и вызывает нужный обработчик.
9. **Взаимодействие сервисов:** Сервисы взаимодействуют друг с другом посредством отправки команд и публикации событий через каналы. Например, `OrderService`, после создания заказа, отправляет команду `ProcessPayment` в `PaymentService`.
10. **Компенсирующие транзакции:** В данном примере, при возникновении ошибок (симуляция через закоментированный код), `OrderService` иницирует компенсирующую транзакцию `CancelOrder`
11. **Вывод результата:** После небольшой задержки, необходимой для выполнения Saga, выводится итоговый статус заказа.

## Заключение

Выбор конкретного подхода к реализации распределенных транзакций зависит от требований приложения, таких как уровень согласованности, допустимая задержка, сложность реализации и требования к масштабируемости. 2PC и 3PC обеспечивают строгую согласованность, но могут привести к блокировкам и снижению доступности. Saga обеспечивает более высокую доступность и слабую связанность, но требует тщательного проектирования компенсирующих транзакций и может привести к временной несогласованности данных.

```old
Distributed Transactions
```
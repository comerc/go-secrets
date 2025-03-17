#DomainDrivenDesign #DDD #SoftwareDesign #SoftwareArchitecture #UbiquitousLanguage #BoundedContext #Entity #ValueObject #Aggregate #DomainService #Repository #Factory #AntiCorruptionLayer #EventDrivenArchitecture

# Domain-Driven Design

```table-of-contents
```

Domain-Driven Design (DDD) — это подход к разработке программного обеспечения, который ставит во главу угла предметную область (домен), для которой создается программное обеспечение. DDD предлагает набор концепций и шаблонов, которые помогают структурировать сложные системы, делая их более понятными, поддерживаемыми и приспособленными к изменениям.

## Основные принципы DDD

DDD строится вокруг нескольких ключевых принципов:

1.  **Фокус на предметной области:** Разработка начинается с глубокого понимания предметной области. Разработчики и эксперты предметной области тесно сотрудничают, чтобы создать общую модель, отражающую бизнес-процессы и правила.

2.  **Единый язык (Ubiquitous Language):** Создание общего языка, который используется всеми участниками проекта (разработчиками, экспертами, менеджерами). Этот язык описывает концепции и процессы предметной области и используется как в коде, так и в документации, и в устном общении. Это помогает избежать недопонимания и несоответствий.

3.  **Ограниченные контексты (Bounded Contexts):** Разбиение большой и сложной предметной области на более мелкие, управляемые части, называемые ограниченными контекстами. Каждый контекст имеет свою собственную модель предметной области и единый язык. Это позволяет упростить модели и избежать конфликтов между различными частями системы.

4.  **Слоистая архитектура (Layered Architecture):** Разделение приложения на слои, каждый из которых имеет свою ответственность. Обычно выделяют следующие слои:
    *   **Интерфейс пользователя (User Interface):** Отвечает за взаимодействие с пользователем.
    *   **Прикладной слой (Application Layer):** Координирует действия, делегируя выполнение задач доменному слою. Не содержит бизнес-логики.
    *   **Доменный слой (Domain Layer):** Содержит модель предметной области, бизнес-логику и правила.
    *   **Инфраструктурный слой (Infrastructure Layer):** Предоставляет технические возможности, такие как доступ к базе данных, отправка сообщений и т.д.

## Ключевые понятия DDD

Рассмотрим более детально основные концепции DDD.

### Сущность (Entity)

Сущность — это объект предметной области, который имеет уникальный идентификатор и может изменяться с течением времени. Идентификатор остается неизменным, даже если атрибуты сущности меняются. Примеры сущностей: Клиент, Заказ, Продукт.

**Пример (Go):**

```go
package domain

type Customer struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
}

func (c *Customer) ChangeEmail(newEmail string) {
	// Валидация email может быть вынесена в отдельный Value Object
	c.Email = newEmail
}
```

В этом примере `Customer` является сущностью, так как имеет уникальный идентификатор (`ID`) и его состояние (например, `Email`) может изменяться.

### Объект-значение (Value Object)

Объект-значение — это объект, который описывает характеристику или аспект предметной области. Он не имеет уникального идентификатора и полностью определяется своими атрибутами. Объекты-значения обычно неизменяемы (immutable). Если нужно изменить объект-значение, создается новый экземпляр. Примеры: Адрес, Деньги, Цвет.

**Пример (Go):**

```go
package domain

type Address struct {
	Street  string
	City    string
	ZipCode string
}

// Метод, создающий новый Address с измененным ZipCode.
func (a Address) WithZipCode(newZipCode string) Address {
	return Address{
		Street:  a.Street,
		City:    a.City,
		ZipCode: newZipCode,
	}
}
```

`Address` является объектом-значением, так как не имеет идентификатора и полностью определяется своими полями. Обратите внимание на метод `WithZipCode`, который возвращает *новый* объект `Address` вместо изменения текущего.

### Агрегат (Aggregate)

Агрегат — это кластер связанных сущностей и объектов-значений, который рассматривается как единое целое с точки зрения согласованности данных. У агрегата есть корень (Aggregate Root) — единственная сущность, через которую можно взаимодействовать с агрегатом извне. Корень агрегата гарантирует целостность всего агрегата.

**Пример (Go):**

```go
package domain

// Order - корень агрегата.
type Order struct {
	ID          string
	CustomerID  string
	OrderLines  []OrderLine
	ShippingAddress Address
	// ... другие поля
}

type OrderLine struct {
	ProductID string
	Quantity  int
	Price     Money // Value Object
}

// AddOrderLine добавляет строку заказа в заказ.
// Гарантирует целостность агрегата.
func (o *Order) AddOrderLine(productID string, quantity int, price Money) {
	// Проверки, например, что quantity > 0.
	// Может проверять дублирование ProductID.
    if quantity <=0 {
        panic("Quantity shoul be greater than zero")
    }
	orderLine := OrderLine{ProductID: productID, Quantity: quantity, Price: price}
	o.OrderLines = append(o.OrderLines, orderLine)
}

// CalculateTotalPrice вычисляет общую стоимость заказа.
func (o *Order) CalculateTotalPrice() Money {
    //Логика расчета
    var totalAmount float64 = 0
    for _, line := range o.OrderLines{
        totalAmount += line.Price.Amount * float64(line.Quantity)
    }

    return Money{Amount: totalAmount, Currency: "USD"} //Предпологаем, что валюта заказа доллары США
}

type Money struct{
    Amount float64
    Currency string
}
```

В этом примере `Order` является корнем агрегата, а `OrderLine` — частью агрегата. Взаимодействие с `OrderLine` происходит только через `Order`. Метод `AddOrderLine` гарантирует, что в заказ добавляются только корректные строки заказа.

### Доменный сервис (Domain Service)

Доменный сервис — это операция или набор операций, которые не принадлежат естественным образом какой-либо сущности или объекту-значению, но важны для предметной области. Доменные сервисы моделируют бизнес-процессы, которые включают в себя несколько сущностей или требуют взаимодействия с внешними системами.

**Пример (Go):**

```go
package domain

// OrderPlacementService - доменный сервис для размещения заказа.
type OrderPlacementService struct {
	productRepository  ProductRepository //Интерфейс репозитория
	customerRepository CustomerRepository //Интерфейс репозитория
	orderRepository    OrderRepository    //Интерфейс репозитория
}

// PlaceOrder размещает заказ.
func (s *OrderPlacementService) PlaceOrder(customerID string, productIDs []string, quantities []int, shippingAddress Address) (*Order, error) {
	// 1. Получить клиента.
	customer, err := s.customerRepository.FindByID(customerID)
	if err != nil {
		return nil, err
	}

	// 2. Создать заказ.
	order := &Order{
		ID:          generateOrderID(), // Функция для генерации ID.
		CustomerID:  customer.ID,
		OrderLines:  make([]OrderLine, 0),
        ShippingAddress: shippingAddress,
	}

	// 3. Добавить строки заказа.
	for i, productID := range productIDs {
		// 3.1. Получить продукт.
		product, err := s.productRepository.FindByID(productID)
		if err != nil {
			return nil, err
		}

		// 3.2. Добавить строку заказа.
		order.AddOrderLine(product.ID, quantities[i], product.Price)
	}

	// 4. Сохранить заказ.
	err = s.orderRepository.Save(order)
	if err != nil {
		return nil, err
	}

	return order, nil
}

// Вспомогательные интерфейсы

type ProductRepository interface {
	FindByID(id string) (*Product, error)
	// ... другие методы
}

type CustomerRepository interface {
	FindByID(id string) (*Customer, error)
	// ... другие методы
}

type OrderRepository interface{
    Save(order *Order) error
    // ... другие методы
}

type Product struct{
    ID string
    Price Money
}

func generateOrderID() string{
    //Some logic for generating order id
    return "some_order_id"
}
```

`OrderPlacementService` — это доменный сервис, который координирует процесс размещения заказа. Он использует репозитории (`ProductRepository`, `CustomerRepository`, `OrderRepository`) для получения и сохранения данных.

### Репозиторий (Repository)

Репозиторий предоставляет абстракцию над механизмом хранения данных. Он скрывает детали работы с базой данных (или другим хранилищем) и предоставляет интерфейс, ориентированный на доменные объекты. Репозитории обычно используются для получения и сохранения агрегатов.

Пример интерфейса `OrderRepository` был приведен выше в разделе про Доменный Сервис.

### Фабрика (Factory)

Фабрика — это объект, который отвечает за создание сложных объектов (например, агрегатов) или объектов, создание которых требует сложной логики. Фабрики инкапсулируют логику создания и гарантируют, что создаваемые объекты находятся в корректном состоянии.

**Пример (Go):**

```go
package domain

// OrderFactory - фабрика для создания заказов.
type OrderFactory struct {
}

// CreateOrder создает новый заказ.
func (f *OrderFactory) CreateOrder(customerID string, shippingAddress Address) *Order {
	return &Order{
		ID:          generateOrderID(), // Функция для генерации ID.
		CustomerID:  customerID,
		ShippingAddress: shippingAddress,
		OrderLines:  make([]OrderLine, 0),
	}
}
```

`OrderFactory` инкапсулирует логику создания заказа, предоставляя простой интерфейс для создания новых заказов.

### Антикоррупционный слой (Anti-Corruption Layer - ACL)

Антикоррупционный слой — это промежуточный слой между вашей моделью предметной области и другой системой (например, устаревшей системой или сторонним сервисом). ACL преобразует данные из формата внешней системы в формат вашей модели и наоборот, защищая вашу модель от влияния внешней системы.

**Пример (Go):**

Предположим, у нас есть устаревшая система, которая возвращает информацию о клиенте в следующем формате:

```go
// LegacyCustomer - структура данных из устаревшей системы.
type LegacyCustomer struct {
	CustomerID   string
	FullName     string // Вместо FirstName и LastName
	ContactEmail string // Вместо Email
}
```

Нам нужно преобразовать `LegacyCustomer` в нашу доменную модель `Customer`.

```go
package domain

// CustomerAdapter - адаптер для преобразования LegacyCustomer в Customer.
type CustomerAdapter struct {
	legacyCustomer LegacyCustomer
}

// ToCustomer преобразует LegacyCustomer в Customer.
func (a *CustomerAdapter) ToCustomer() *Customer {
	// Разделение FullName на FirstName и LastName.
	nameParts := strings.Split(a.legacyCustomer.FullName, " ")
    firstName := ""
    lastName := ""

    if len(nameParts) > 0{
        firstName = nameParts[0]
    }

    if len(nameParts) > 1{
	    lastName = nameParts[1]
    }

	return &Customer{
		ID:        a.legacyCustomer.CustomerID,
		FirstName: firstName,
		LastName:  lastName,
		Email:     a.legacyCustomer.ContactEmail,
	}
}

import "strings"
```

`CustomerAdapter` действует как антикоррупционный слой, преобразуя данные из формата устаревшей системы в формат нашей доменной модели.

## [[Event-Driven Architecture]] и DDD

DDD хорошо сочетается с событийно-ориентированной архитектурой (Event-Driven Architecture, EDA). Доменные события (Domain Events) — это значимые изменения в состоянии предметной области, которые могут быть использованы для уведомления других частей системы или внешних систем.

**Пример (Go):**

```go
package domain

// DomainEvent - интерфейс для доменных событий.
type DomainEvent interface {
	EventName() string
}

// OrderPlaced - событие, возникающее при размещении заказа.
type OrderPlaced struct {
	OrderID    string
	CustomerID string
	OrderTotal Money
}

func (e *OrderPlaced) EventName() string {
	return "OrderPlaced"
}

// В Order добавляем метод для генерации события.
func (o *Order) Place() []DomainEvent {
	// ... другая логика размещения заказа ...
    //Расчет итоговой суммы заказа
    total := o.CalculateTotalPrice()

	return []DomainEvent{&OrderPlaced{OrderID: o.ID, CustomerID: o.CustomerID, OrderTotal: total}}
}

//Пример использования
// order := placeOrder(...params...)
// events := order.Place()
// for _, event := range events {
// 	eventBus.Publish(event) // Публикация события в шину событий.
// }
```
В этом примере `OrderPlaced` это [[Domain Event|доменное событие]], которое генерируется при размещении заказа. Это событие может быть опубликовано в шину событий и обработано другими частями системы (например, для обновления складских запасов или отправки уведомления клиенту).

## Преимущества и недостатки DDD

**Преимущества:**

*   **Улучшенное понимание предметной области:** DDD способствует глубокому пониманию предметной области и созданию модели, которая точно отражает бизнес-процессы.
*   **Более гибкая и поддерживаемая архитектура:** DDD помогает структурировать сложные системы, делая их более понятными и легкими для изменения.
*   **Улучшенное взаимодействие между разработчиками и экспертами:** Единый язык и общая модель способствуют лучшему взаимопониманию между всеми участниками проекта.
*   **Более тесная связь с бизнес-потребностями:** DDD помогает создавать программное обеспечение, которое лучше соответствует реальным потребностям бизнеса.

**Недостатки:**

*   **Более высокая начальная сложность:** DDD требует значительных усилий на начальном этапе для анализа предметной области и создания модели.
*   **Требует высокой квалификации команды:** Для успешного применения DDD требуется, чтобы команда имела хорошее понимание принципов и шаблонов DDD.
*   **Может быть избыточным для простых проектов:** Для небольших и простых проектов DDD может оказаться излишне сложным.

## Заключение

Domain-Driven Design — это мощный подход к разработке программного обеспечения, который может значительно улучшить качество и поддерживаемость сложных систем. Однако он требует тщательного анализа предметной области и хорошего понимания принципов DDD. DDD не является серебряной пулей и подходит не для всех проектов. Важно оценивать сложность проекта и выбирать подходящий подход к разработке.

```old
Domain-Driven Design (DDD)
```
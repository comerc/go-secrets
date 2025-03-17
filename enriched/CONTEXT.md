#golang #go #context #programming #concurrency #synchronization #deadline #timeout #cancellation #values

# Понятие контекста (context) в Go

```table-of-contents
```

## Введение в контекст

Контекст в Go, как и в повседневной жизни, представляет собой совокупность обстоятельств, влияющих на выполнение программы в определенный момент времени.  Представьте ситуацию: человек следит за модой или финансовыми рынками. Эти внешние факторы (мода, состояние рынка) формируют "контекст", в котором действует человек.  Аналогично, в программировании контекст определяет набор параметров и условий, в которых выполняются функции.

В Go контекст часто используется для передачи параметров состояния по иерархии вызовов функций и, что особенно важно, для управления сигналами завершения программы.  Это похоже на ситуацию в очереди в кассу, где последний покупатель сообщает: "За мной не занимать".  Этот сигнал ("за мной не занимать") распространяется по очереди, информируя всех о скором завершении обслуживания. В Go контекст позволяет элегантно управлять подобными сигналами в асинхронных операциях.

Контекст в Go реализует принцип [[KISS]] & [[DRY]], позволяя избежать дублирования передачи одних и тех же параметров в разные функции. Вместо этого создается единый объект контекста, содержащий всю необходимую информацию.

## Основные цели использования контекста

Контекст в Go решает следующие задачи:

1.  **Управление временем жизни (Deadline):** Контекст позволяет установить крайний срок выполнения операции. По истечении этого срока операция может быть автоматически прервана. Это полезно, например, при работе с сетью, когда необходимо ограничить время ожидания ответа от сервера.

2.  **Синхронизация с помощью каналов (Done):** Контекст предоставляет канал, который закрывается при необходимости отмены операции.  Это позволяет горутинам, использующим контекст, узнать о необходимости завершения работы.  Этот механизм является ключевым для реализации [[паттернов отмены]] и [[graceful shutdown]].

3.  **Обработка ошибок (Err):** После завершения контекста (по причине отмены или истечения времени), можно получить информацию о причине завершения через метод `Err()`.

4.  **Хранение произвольных данных (Value):** Контекст может хранить произвольные данные, связанные с запросом. Эти данные доступны всем функциям, получающим контекст. Важно использовать эту возможность только для данных, относящихся к области действия запроса, и не передавать таким образом необязательные параметры функций. Это помогает избежать [[неявных зависимостей]] и [[улучшает читаемость кода]].

## Интерфейс `context.Context`

Интерфейс `context.Context` в Go определяет четыре основных метода:

```go
type Context interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key any) any
}
```

Разберем каждый метод подробнее:

### `Deadline()`

Метод `Deadline()` возвращает время, до которого должна быть завершена операция, связанная с контекстом.  Возвращаемые значения:

*   `deadline time.Time`:  Время дедлайна.
*   `ok bool`:  Логическое значение, указывающее, был ли установлен дедлайн.  Если `ok` равно `false`, значит дедлайн не установлен.

Последовательные вызовы `Deadline()` возвращают одни и те же значения.

### `Done()`

Метод `Done()` возвращает канал типа `<-chan struct{}`. Этот канал закрывается, когда работа, связанная с контекстом, должна быть отменена.  Закрытие канала `Done` сигнализирует всем горутинам, ожидающим на этом канале, о необходимости завершения.

Важно отметить, что закрытие канала `Done` может произойти асинхронно после вызова функции отмены.

Канал `Done` обычно используется в операторе `select` для ожидания сигнала отмены наряду с другими событиями:

```go
func Stream(ctx context.Context, out chan<- int) error {
	for {
		// Получаем данные
		v, err := DoSomething(ctx)
		if err != nil {
			return err
		}

		// Ожидаем либо сигнала отмены, либо возможности отправить данные
		select {
		case <-ctx.Done():
			// Контекст отменен, возвращаем ошибку
			return ctx.Err()
		case out <- v:
			// Данные успешно отправлены
		}
	}
}
```

В этом примере горутина `Stream` генерирует значения и отправляет их в канал `out`.  Оператор `select` позволяет одновременно ожидать двух событий: закрытия канала `ctx.Done()` (сигнал отмены) и возможности отправить данные в канал `out`.  Если контекст отменен, горутина завершается и возвращает ошибку.

### `Err()`

Метод `Err()` возвращает ошибку, объясняющую причину завершения контекста.  Если канал `Done` еще не закрыт, `Err()` возвращает `nil`.  Если `Done` закрыт, `Err()` возвращает ненулевую ошибку:

*   `context.Canceled`:  Если контекст был отменен явно.
*   `context.DeadlineExceeded`:  Если истек срок действия контекста.

Последовательные вызовы `Err()` после закрытия канала `Done` возвращают одну и ту же ошибку.

### `Value(key any)`

Метод `Value(key any)` позволяет получить значение, связанное с контекстом по ключу.  Если значение по данному ключу не найдено, возвращается `nil`.  Последовательные вызовы `Value()` с одним и тем же ключом возвращают один и тот же результат.

Ключи для хранения значений в контексте должны быть сравнимыми типами.  Рекомендуется определять ключи как неэкспортируемые типы, чтобы избежать коллизий имен:

```go
package user

import "context"

// User - тип данных пользователя.
type User struct {
	Name string
	ID   int
}

// key - неэкспортируемый тип для ключей контекста.
type key int

// userKey - ключ для хранения данных пользователя в контексте.
const userKey key = 0

// NewContext создает новый контекст с данными пользователя.
func NewContext(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

// FromContext извлекает данные пользователя из контекста.
func FromContext(ctx context.Context) (*User, bool) {
	u, ok := ctx.Value(userKey).(*User)
	return u, ok
}
```

В этом примере определен тип `User` для хранения данных пользователя и неэкспортируемый тип `key` для ключей контекста.  Константа `userKey` используется в качестве ключа для хранения данных пользователя.  Функции `NewContext` и `FromContext` обеспечивают типобезопасный доступ к данным пользователя в контексте.

## Создание контекстов

Пакет `context` предоставляет несколько функций для создания контекстов:

*   `context.Background()`: Возвращает пустой контекст.  Он никогда не отменяется, не имеет срока действия и не содержит значений.  Обычно используется в качестве корневого контекста в `main` функции, тестах и при инициализации.

*   `context.TODO()`: Возвращает пустой контекст.  Используется, когда неясно, какой контекст использовать, или когда контекст еще не доступен.  Это своего рода заглушка, которую следует заменить на более конкретный контекст в будущем.

*   `context.WithCancel(parent context.Context)`: Возвращает копию родительского контекста с новой функцией отмены.  Канал `Done` нового контекста закрывается при вызове функции отмены или при закрытии канала `Done` родительского контекста (в зависимости от того, что произойдет раньше).

*   `context.WithDeadline(parent context.Context, d time.Time)`: Возвращает копию родительского контекста с установленным сроком действия.  Канал `Done` нового контекста закрывается, когда наступает указанное время или когда закрывается канал `Done` родительского контекста.

*   `context.WithTimeout(parent context.Context, timeout time.Duration)`: Возвращает копию родительского контекста с установленным таймаутом.  Эквивалентно вызову `context.WithDeadline(parent, time.Now().Add(timeout))`.

*   `context.WithValue(parent context.Context, key, val any)`: Возвращает копию родительского контекста, в котором значение `val` ассоциировано с ключом `key`.

Пример создания и использования контекста с отменой:

```go
package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	// Создаем контекст с отменой.
	ctx, cancel := context.WithCancel(context.Background())

	// Запускаем горутину, которая работает в контексте.
	go worker(ctx)

	// Ждем 3 секунды.
	time.Sleep(3 * time.Second)

	// Отменяем контекст.
	cancel()

	// Ждем еще немного, чтобы горутина завершилась.
	time.Sleep(1 * time.Second)
	fmt.Println("main: finished")
}

func worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// Контекст отменен, завершаем работу.
			fmt.Println("worker: context cancelled")
			return
		default:
			// Выполняем какую-то работу.
			fmt.Println("worker: working...")
			time.Sleep(500 * time.Millisecond)
		}
	}
}

```

В этом примере создается контекст с функцией отмены `cancel`.  Горутина `worker` выполняет работу в цикле, периодически проверяя, не отменен ли контекст.  Через 3 секунды вызывается функция `cancel`, что приводит к закрытию канала `ctx.Done()` и завершению горутины `worker`.

Пример создания и использования контекста с таймаутом.
```go
package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	// Создаем контекст с таймаутом 2 секунды.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel() // Важно вызывать cancel, даже если контекст истек по таймауту!

	// Запускаем горутину, которая работает в контексте.
	go worker(ctx)

    // ждем, пока воркер отработает
	time.Sleep(3 * time.Second)
	fmt.Println("main: finished")
}

func worker(ctx context.Context) {
	select {
	case <-time.After(5 * time.Second):
		fmt.Println("worker: task completed")
	case <-ctx.Done():
		fmt.Println("worker: context cancelled or timed out:", ctx.Err())
	}
}
```
В этом примере `defer cancel()` освобождает ресурсы, если горутина завершилась раньше таймаута.

## Распространение контекста

Контекст должен явно передаваться между функциями.  Не следует хранить контекст в структурах или глобальных переменных.  Это обеспечивает прозрачность и контроль над распространением контекста.

Пример правильного распространения контекста:

```go
func DoSomething(ctx context.Context, arg string) error {
	// Используем ctx.
	// ...

	// Передаем ctx в другую функцию.
	if err := AnotherFunction(ctx, arg); err != nil {
		return err
	}

	return nil
}

func AnotherFunction(ctx context.Context, arg string) error {
	// Используем ctx.
	// ...

	return nil
}
```
В этом примере контекст `ctx` явно передается в функции `DoSomething` и `AnotherFunction`.
Контекст создается, например, в обработчике HTTP-запроса и передаётся дальше по цепочке вызовов.

## Отмена контекста и освобождение ресурсов

Функция отмены, возвращаемая функциями `WithCancel`, `WithDeadline` и `WithTimeout`, должна быть вызвана, когда контекст больше не нужен.  Это освобождает ресурсы, связанные с контекстом. Невызов функции отмены может привести к утечке ресурсов. Вызов `cancel()` идемпотентен, т.е. многократный вызов не приведёт к ошибке.

## Резюме

Контекст в Go — это мощный инструмент для управления временем жизни операций, синхронизации горутин, обработки ошибок и передачи данных, связанных с запросом.  Правильное использование контекста позволяет создавать надежные и масштабируемые приложения. Ключевые моменты:

*   Явно передавайте контекст между функциями.
*   Используйте `context.Background()` в качестве корневого контекста.
*   Используйте `context.TODO()`, когда неясно, какой контекст использовать.
*   Всегда вызывайте функцию отмены, когда контекст больше не нужен.
*   Используйте значения контекста только для данных, относящихся к области действия запроса.
*   Не храните контекст в структурах.
*   Используйте неэкспортируемые типы для ключей контекста.

Контекст является неотъемлемой частью идиоматичного Go кода и широко используется в стандартной библиотеке и сторонних пакетах. Понимание и правильное применение контекста - важный навык для любого Go разработчика.

```old
## Понятие контекста (context) в Go

Что такое контекст вообще? Это совокупность обстоятельств, которые окружают субъект в определённый момент времени. Например, X не отстаёт от моды, или Y отслеживает финансовые рынки. Если говорить о коде, то мы можем задать набор параметров для функций, и передавать туда конкретные значения текущего состояния программы. Но, когда у нас несколько разных функций, которым нужно передать тот же самый набор параметров, то напрашивается обобщение - контекст выполнения программы. Принцип KISS & DRY. В Go очень часто применяется контекст для ослеживания сигнала на завершение программы. А так же для передачи каких-либо других параметров состояния по иерархии вызова функций. Например, очередь в кассу обрабатывается до крайнего покупателя, который дальше сообщает "попросили за мной не занимать". 

### Зачем контекст

- Получение дедлайна - Deadline
- Синхронизация при помощи каналов - Done
- Получение причины завершения - Err
- Хранение произвольных данных - Value

### Интерфейс

\`\`\`go
// A Context carries a deadline, a cancellation signal, and other values across
// API boundaries.
//
// Context's methods may be called by multiple goroutines simultaneously.
type Context interface {
	// Deadline возвращает время, когда работа, выполняемая на основе этого контекста, должна быть отменена. Deadline возвращает ok==false, когда не установлено время ожидания. Последовательные вызовы Deadline возвращают одинаковые результаты.

	// Deadline returns the time when work done on behalf of this context
	// should be canceled. Deadline returns ok==false when no deadline is
	// set. Successive calls to Deadline return the same results.
	Deadline() (deadline time.Time, ok bool)

	// Done возвращает канал, который закрывается, когда необходимо отменить работу,
	// выполненную на основе этого контекста. Done может вернуть nil, если этот
	// контекст никогда не может быть отменен. Последующие вызовы Done возвращают
	// то же значение. Закрытие канала Done может произойти асинхронно, после
	// вызова функции отмены.
	//
	// WithCancel организует закрытие канала Done при вызове функции отмены;
	// WithDeadline организует закрытие канала Done при истечении срока
	// действия; WithTimeout организует закрытие канала Done при истечении
	// таймера.
	//
	// Done предоставляется для использования в операторах select:
	//
	//   // Stream генерирует значения с помощью DoSomething и отправляет их в
	//   // out, пока DoSomething не вернет ошибку или ctx.Done не будет закрыт.
	//   func Stream(ctx context.Context, out chan<- Value) error {
	//   	for {
	//   		v, err := DoSomething(ctx)
	//   		if err != nil {
	//   			return err
	//   		}
	//   		select {
	//   		case <-ctx.Done():
	//   			return ctx.Err()
	//   		case out <- v:
	//   		}
	//   	}
	//   }
	//
	// Смотрите https://blog.golang.org/pipelines для дополнительных примеров того,
	// как использовать канал Done для отмены.

	// Done returns a channel that's closed when work done on behalf of this
	// context should be canceled. Done may return nil if this context can
	// never be canceled. Successive calls to Done return the same value.
	// The close of the Done channel may happen asynchronously,
	// after the cancel function returns.
	//
	// WithCancel arranges for Done to be closed when cancel is called;
	// WithDeadline arranges for Done to be closed when the deadline
	// expires; WithTimeout arranges for Done to be closed when the timeout
	// elapses.
	//
	// Done is provided for use in select statements:
	//
	//  // Stream generates values with DoSomething and sends them to out
	//  // until DoSomething returns an error or ctx.Done is closed.
	//  func Stream(ctx context.Context, out chan<- Value) error {
	//  	for {
	//  		v, err := DoSomething(ctx)
	//  		if err != nil {
	//  			return err
	//  		}
	//  		select {
	//  		case <-ctx.Done():
	//  			return ctx.Err()
	//  		case out <- v:
	//  		}
	//  	}
	//  }
	//
	// See https://blog.golang.org/pipelines for more examples of how to use
	// a Done channel for cancellation.
	Done() <-chan struct{}

	// If Done is not yet closed, Err returns nil.
	// If Done is closed, Err returns a non-nil error explaining why:
	// Canceled if the context was canceled
	// or DeadlineExceeded if the context's deadline passed.
	// After Err returns a non-nil error, successive calls to Err return the same error.
	Err() error

	// Value returns the value associated with this context for key, or nil
	// if no value is associated with key. Successive calls to Value with
	// the same key returns the same result.
	//
	// Use context values only for request-scoped data that transits
	// processes and API boundaries, not for passing optional parameters to
	// functions.
	//
	// A key identifies a specific value in a Context. Functions that wish
	// to store values in Context typically allocate a key in a global
	// variable then use that key as the argument to context.WithValue and
	// Context.Value. A key can be any type that supports equality;
	// packages should define keys as an unexported type to avoid
	// collisions.
	//
	// Packages that define a Context key should provide type-safe accessors
	// for the values stored using that key:
	//
	// 	// Package user defines a User type that's stored in Contexts.
	// 	package user
	//
	// 	import "context"
	//
	// 	// User is the type of value stored in the Contexts.
	// 	type User struct {...}
	//
	// 	// key is an unexported type for keys defined in this package.
	// 	// This prevents collisions with keys defined in other packages.
	// 	type key int
	//
	// 	// userKey is the key for user.User values in Contexts. It is
	// 	// unexported; clients use user.NewContext and user.FromContext
	// 	// instead of using this key directly.
	// 	var userKey key
	//
	// 	// NewContext returns a new Context that carries value u.
	// 	func NewContext(ctx context.Context, u *User) context.Context {
	// 		return context.WithValue(ctx, userKey, u)
	// 	}
	//
	// 	// FromContext returns the User value stored in ctx, if any.
	// 	func FromContext(ctx context.Context) (*User, bool) {
	// 		u, ok := ctx.Value(userKey).(*User)
	// 		return u, ok
	// 	}
	Value(key any) any
}
\`\`\`

```
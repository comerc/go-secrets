#concurrency #deadlock #golang #multithreading #synchronization

# Взаимная блокировка (Deadlock)

```table-of-contents
```

## Что такое взаимная блокировка

Взаимная блокировка (deadlock) — это ситуация в многопоточном программировании, когда два или более потока блокируются навсегда, ожидая друг от друга освобождения ресурсов. Это критическое состояние, при котором процесс выполнения программы останавливается без возможности восстановления без внешнего вмешательства.

В Go взаимная блокировка обнаруживается во время выполнения программы и приводит к панике с сообщением "fatal error: all goroutines are asleep - deadlock!".

## Условия возникновения взаимной блокировки

Для возникновения взаимной блокировки должны одновременно выполняться четыре условия Коффмана:

1. **Взаимное исключение**: ресурс может использовать только один поток одновременно
2. **Удержание и ожидание**: поток удерживает как минимум один ресурс и ожидает дополнительные ресурсы, которые в данный момент удерживаются другими потоками
3. **Отсутствие перехвата**: ресурс не может быть принудительно отобран у потока, удерживающего его
4. **Циклическое ожидание**: существует цепочка потоков ожидает ресурс, удерживаемый следующим потоком в цепочке

## Примеры взаимной блокировки в Go

### Пример 1: Блокировка с использованием мьютексов

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var mutex1, mutex2 sync.Mutex

	go func() {
		mutex1.Lock()
		fmt.Println("Горутина 1: Захвачен mutex1")
		time.Sleep(100 * time.Millisecond)
		fmt.Println("Горутина 1: Ожидание mutex2")
		mutex2.Lock()
		fmt.Println("Горутина 1: Захвачены оба мьютекса")
		mutex2.Unlock()
		mutex1.Unlock()
	}()

	go func() {
		mutex2.Lock()
		fmt.Println("Горутина 2: Захвачен mutex2")
		time.Sleep(100 * time.Millisecond)
		fmt.Println("Горутина 2: Ожидание mutex1")
		mutex1.Lock()
		fmt.Println("Горутина 2: Захвачены оба мьютекса")
		mutex1.Unlock()
		mutex2.Unlock()
	}()

	time.Sleep(2 * time.Second)
}
```

В этом примере две горутины захватывают мьютексы в разном порядке, что приводит к взаимной блокировке: первая горутина ждет освобождения mutex2, который удерживается второй горутиной, а вторая горутина ждет освобождения mutex1, который удерживается первой горутиной.

### Пример 2: Блокировка при неправильном использовании каналов

```go
package main

func main() {
	ch := make(chan int) // Небуферизованный канал
	ch <- 1              // Отправка в канал блокирует выполнение
	// Здесь никто не читает из канала, поэтому программа блокируется
	<-ch
}
```

В этом примере происходит блокировка, потому что отправка в небуферизованный канал блокирует горутину до тех пор, пока другая горутина не будет готова принять значение. Поскольку отправка и получение происходят в одной горутине последовательно, программа блокируется навсегда.

## Методы предотвращения взаимной блокировки

### 1. Использование тайм-аутов

```go
package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	ch := make(chan int)

	go func() {
		time.Sleep(2 * time.Second) // Имитация долгой работы
		ch <- 42
	}()

	select {
	case result := <-ch:
		fmt.Println("Получен результат:", result)
	case <-ctx.Done():
		fmt.Println("Операция прервана по тайм-ауту:", ctx.Err())
	}
}
```

### 2. Соблюдение порядка блокировок

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var mutex1, mutex2 sync.Mutex

	// Обе горутины захватывают мьютексы в одинаковом порядке
	go func() {
		mutex1.Lock()
		fmt.Println("Горутина 1: Захвачен mutex1")
		time.Sleep(100 * time.Millisecond)
		mutex2.Lock()
		fmt.Println("Горутина 1: Захвачены оба мьютекса")
		mutex2.Unlock()
		mutex1.Unlock()
	}()

	go func() {
		mutex1.Lock()
		fmt.Println("Горутина 2: Захвачен mutex1")
		time.Sleep(100 * time.Millisecond)
		mutex2.Lock()
		fmt.Println("Горутина 2: Захвачены оба мьютекса")
		mutex2.Unlock()
		mutex1.Unlock()
	}()

	time.Sleep(2 * time.Second)
}
```

### 3. Использование TryLock

В стандартной библиотеке Go нет прямого эквивалента TryLock, но его можно реализовать с помощью каналов или пакета `sync/atomic`:

```go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Мьютекс с поддержкой TryLock
type TryMutex struct {
	locked int32
	m      sync.Mutex
}

// Пытается захватить мьютекс без блокировки
func (m *TryMutex) TryLock() bool {
	return atomic.CompareAndSwapInt32(&m.locked, 0, 1)
}

// Обычная блокировка
func (m *TryMutex) Lock() {
	m.m.Lock()
	atomic.StoreInt32(&m.locked, 1)
}

// Разблокировка
func (m *TryMutex) Unlock() {
	atomic.StoreInt32(&m.locked, 0)
	m.m.Unlock()
}

func main() {
	var mutex1, mutex2 TryMutex

	go func() {
		if mutex1.TryLock() {
			fmt.Println("Горутина 1: Захвачен mutex1")
			time.Sleep(100 * time.Millisecond)
			
			if mutex2.TryLock() {
				fmt.Println("Горутина 1: Захвачены оба мьютекса")
				time.Sleep(100 * time.Millisecond)
				mutex2.Unlock()
			} else {
				fmt.Println("Горутина 1: Не удалось захватить mutex2, освобождаем mutex1")
			}
			
			mutex1.Unlock()
		}
	}()

	go func() {
		if mutex2.TryLock() {
			fmt.Println("Горутина 2: Захвачен mutex2")
			time.Sleep(100 * time.Millisecond)
			
			if mutex1.TryLock() {
				fmt.Println("Горутина 2: Захвачены оба мьютекса")
				time.Sleep(100 * time.Millisecond)
				mutex1.Unlock()
			} else {
				fmt.Println("Горутина 2: Не удалось захватить mutex1, освобождаем mutex2")
			}
			
			mutex2.Unlock()
		}
	}()

	time.Sleep(2 * time.Second)
}
```

### 4. Использование буферизованных каналов

```go
package main

import "fmt"

func main() {
	ch := make(chan int, 1) // Буферизованный канал с емкостью 1
	ch <- 1                 // Не блокируется, так как канал имеет буфер
	fmt.Println(<-ch)       // Успешно прочитано из канала
}
```

## Обнаружение взаимной блокировки

В Go встроен детектор взаимных блокировок, который выводит сообщение "fatal error: all goroutines are asleep - deadlock!" когда все горутины заблокированы.

Для обнаружения потенциальных взаимных блокировок можно использовать:

1. **Статический анализ кода** - инструменты типа `go vet` и сторонние анализаторы
2. **Профилирование** - инструменты профилирования Go могут помочь обнаружить проблемы с блокировками
3. **Трассировка** - использование `runtime/trace` для анализа выполнения программы
4. **Таймауты и контексты** - добавление таймаутов помогает обнаружить долгие блокировки

## Паттерны предотвращения взаимных блокировок

### 1. Иерархия блокировок

Установите глобальный порядок получения блокировок и всегда соблюдайте этот порядок.

```go
// Всегда блокируйте в порядке возрастания идентификаторов ресурсов
func SafeOperation(resource1, resource2 *Resource) {
	if resource1.ID < resource2.ID {
		resource1.Lock()
		resource2.Lock()
		defer resource2.Unlock()
		defer resource1.Unlock()
	} else {
		resource2.Lock()
		resource1.Lock()
		defer resource1.Unlock()
		defer resource2.Unlock()
	}
	
	// Выполнение операций с ресурсами
}
```

### 2. Использование контекстов с таймаутом

```go
func performOperationWithTimeout(ctx context.Context) error {
	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	resultCh := make(chan result)
	errCh := make(chan error)
	
	go func() {
		// Выполнение потенциально блокирующей операции
		res, err := performBlockingOperation()
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- res
	}()
	
	select {
	case res := <-resultCh:
		// Обработка результата
		return nil
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return fmt.Errorf("операция прервана: %w", ctx.Err())
	}
}
```

### 3. Использование WaitGroup для синхронизации

```go
func safeParallelExecution() {
	var wg sync.WaitGroup
	results := make(chan int, 10) // Буферизованный канал для результатов
	
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// Выполнение работы
			results <- id * 2
		}(i)
	}
	
	// Закрытие канала после завершения всех горутин
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Чтение результатов
	for result := range results {
		fmt.Println(result)
	}
}
```

## Заключение

Взаимные блокировки — серьезная проблема в многопоточном программировании, которая может привести к полной остановке программы. В Go есть встроенные механизмы обнаружения взаимных блокировок, но ответственность за их предотвращение лежит на разработчике.

Основные методы предотвращения взаимных блокировок:
- Соблюдение порядка блокировок
- Использование таймаутов и контекстов
- Избегание циклических зависимостей
- Использование буферизованных каналов
- Применение неблокирующих алгоритмов, когда это возможно

Понимание причин возникновения взаимных блокировок и применение правильных паттернов синхронизации поможет избежать этой опасной ситуации в многопоточных программах.
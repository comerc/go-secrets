#concurrency #starvation #deadlock #golang #performance

# Голодание (Starvation) в многопоточных системах

```table-of-contents
```

## Что такое голодание в контексте многопоточных систем

Голодание (Starvation) — это состояние в многопоточных системах, когда один или несколько потоков не получают достаточного доступа к общим ресурсам для продолжения своей работы. В отличие от взаимной блокировки (deadlock), при голодании система продолжает функционировать, но некоторые потоки не могут продвигаться в своём выполнении из-за постоянного отсутствия доступа к необходимым ресурсам.

Голодание возникает, когда потоки с более высоким приоритетом или более агрессивным поведением постоянно получают доступ к общим ресурсам, в то время как другие потоки остаются в состоянии ожидания неопределённо долгое время.

## Причины возникновения голодания

### Неправильная приоритизация потоков

Когда система отдаёт предпочтение потокам с более высоким приоритетом, потоки с низким приоритетом могут никогда не получить доступ к ресурсам:

```go
// Пример системы с неправильной приоритизацией
func ResourceManager() {
    for {
        if highPriorityTaskExists() {
            executeHighPriorityTask()
        } else if mediumPriorityTaskExists() {
            executeMediumPriorityTask()
        } else {
            executeLowPriorityTask()
        }
    }
}
```

В этом примере, если высокоприоритетные задачи постоянно поступают, низкоприоритетные никогда не будут выполнены.

### Неэффективные алгоритмы блокировки

Неправильно реализованные механизмы блокировки могут привести к тому, что некоторые потоки будут постоянно проигрывать в конкуренции за ресурсы:

```go
var mu sync.Mutex
var sharedResource int

func worker(id int, aggressive bool) {
    for {
        if aggressive {
            // Агрессивный поток постоянно пытается захватить мьютекс
            mu.Lock()
            sharedResource++
            mu.Unlock()
        } else {
            // Неагрессивный поток делает паузу между попытками
            time.Sleep(10 * time.Millisecond)
            mu.Lock()
            sharedResource++
            mu.Unlock()
        }
    }
}
```

### Несбалансированная нагрузка

Если один поток выполняет более короткие операции с ресурсом, чем другие, он может захватывать ресурс чаще:

```go
func fastWorker() {
    for {
        mu.Lock()
        // Быстрая операция
        sharedResource += 1
        mu.Unlock()
    }
}

func slowWorker() {
    for {
        mu.Lock()
        // Медленная операция
        for i := 0; i < 1000000; i++ {
            sharedResource = complexCalculation(sharedResource)
        }
        mu.Unlock()
    }
}
```

### Неэффективные планировщики задач

Планировщики, которые не учитывают время ожидания потоков, могут привести к голоданию:

```go
func unfairScheduler(tasks []Task) {
    for {
        // Выбор задачи по некоторому критерию, не учитывающему время ожидания
        task := selectTaskByPriority(tasks)
        executeTask(task)
    }
}
```

## Примеры голодания в Go

### Пример 1: Голодание при использовании мьютексов

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func main() {
    var mu sync.Mutex
    counter := 0
    
    // Агрессивный поток
    go func() {
        for i := 0; i < 1000000; i++ {
            mu.Lock()
            counter++
            mu.Unlock()
        }
        fmt.Println("Агрессивный поток завершил работу")
    }()
    
    // Неагрессивный поток
    go func() {
        for i := 0; i < 100; i++ {
            time.Sleep(1 * time.Millisecond)
            mu.Lock()
            counter++
            mu.Unlock()
        }
        fmt.Println("Неагрессивный поток завершил работу")
    }()
    
    time.Sleep(5 * time.Second)
    fmt.Printf("Итоговое значение счётчика: %d\n", counter)
}
```

В этом примере агрессивный поток может захватывать мьютекс настолько часто, что неагрессивный поток будет редко получать доступ к счётчику.

### Пример 2: Голодание при использовании каналов

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    ch := make(chan int)
    
    // Быстрый отправитель
    go func() {
        for i := 0; i < 1000000; i++ {
            select {
            case ch <- i:
                // Отправка успешна
            default:
                // Продолжаем попытки
                i--
            }
        }
    }()
    
    // Медленный отправитель
    go func() {
        for i := 1000000; i < 1000100; i++ {
            time.Sleep(1 * time.Millisecond)
            select {
            case ch <- i:
                fmt.Println("Медленный отправитель: отправлено", i)
            default:
                // Если не удалось отправить, пробуем снова
                i--
            }
        }
        fmt.Println("Медленный отправитель завершил работу")
    }()
    
    // Получатель
    count := 0
    timeout := time.After(3 * time.Second)
    loop:
    for {
        select {
        case <-ch:
            count++
        case <-timeout:
            break loop
        }
    }
    
    fmt.Printf("Получено сообщений: %d\n", count)
}
```

Здесь быстрый отправитель может занимать канал большую часть времени, не давая медленному отправителю возможности отправить свои сообщения.

## Методы предотвращения голодания

### 1. Справедливые блокировки (Fair Locks)

Реализация блокировок, учитывающих порядок запросов:

```go
type FairMutex struct {
    mu      sync.Mutex
    waiting chan struct{}
}

func NewFairMutex() *FairMutex {
    return &FairMutex{
        waiting: make(chan struct{}, 1),
    }
}

func (fm *FairMutex) Lock() {
    ticket := make(chan struct{})
    fm.mu.Lock()
    if fm.waiting != nil {
        queue := fm.waiting
        fm.waiting = ticket
        fm.mu.Unlock()
        <-queue // Ждём своей очереди
    } else {
        fm.waiting = ticket
        fm.mu.Unlock()
    }
}

func (fm *FairMutex) Unlock() {
    fm.mu.Lock()
    if fm.waiting != nil {
        close(fm.waiting)
        fm.waiting = nil
    }
    fm.mu.Unlock()
}
```

### 2. Временные лимиты и таймауты

Установка максимального времени владения ресурсом:

```go
func WithTimeout(timeout time.Duration, fn func()) bool {
    done := make(chan struct{})
    
    go func() {
        fn()
        close(done)
    }()
    
    select {
    case <-done:
        return true
    case <-time.After(timeout):
        return false // Превышен таймаут
    }
}
```

### 3. Алгоритмы планирования с учётом времени ожидания

Реализация планировщика, который учитывает, сколько времени потоки ждут ресурс:

```go
type Task struct {
    ID           int
    Priority     int
    WaitingSince time.Time
    Execute      func()
}

type FairScheduler struct {
    tasks     []Task
    maxWait   time.Duration
    mu        sync.Mutex
}

func (fs *FairScheduler) AddTask(task Task) {
    fs.mu.Lock()
    defer fs.mu.Unlock()
    
    task.WaitingSince = time.Now()
    fs.tasks = append(fs.tasks, task)
}

func (fs *FairScheduler) Run() {
    for {
        fs.mu.Lock()
        if len(fs.tasks) == 0 {
            fs.mu.Unlock()
            time.Sleep(10 * time.Millisecond)
            continue
        }
        
        // Находим задачу с наивысшим приоритетом или наиболее долго ждущую
        var selectedIndex int
        for i, task := range fs.tasks {
            waitTime := time.Since(fs.tasks[selectedIndex].WaitingSince)
            
            if waitTime > fs.maxWait || 
               (time.Since(task.WaitingSince) < fs.maxWait && 
                task.Priority > fs.tasks[selectedIndex].Priority) {
                selectedIndex = i
            }
        }
        
        task := fs.tasks[selectedIndex]
        fs.tasks = append(fs.tasks[:selectedIndex], fs.tasks[selectedIndex+1:]...)
        fs.mu.Unlock()
        
        task.Execute()
    }
}
```

### 4. Ограничение повторных попыток

Предотвращение агрессивного поведения потоков с помощью ограничения частоты попыток:

```go
func retryWithBackoff(fn func() error) error {
    backoff := 10 * time.Millisecond
    maxBackoff := 1 * time.Second
    
    for i := 0; i < 10; i++ {
        err := fn()
        if err == nil {
            return nil
        }
        
        time.Sleep(backoff)
        backoff *= 2
        if backoff > maxBackoff {
            backoff = maxBackoff
        }
    }
    
    return fmt.Errorf("превышено максимальное количество попыток")
}
```

### 5. Использование контекста с отменой

```go
func workerWithContext(ctx context.Context, id int) {
    for {
        select {
        case <-ctx.Done():
            fmt.Printf("Работник %d остановлен\n", id)
            return
        default:
            // Попытка получить доступ к ресурсу с контролем времени
            if acquireResourceWithTimeout() {
                useResource()
                releaseResource()
            }
        }
    }
}
```

## Голодание в реальных системах

### Базы данных

В системах управления базами данных голодание может возникать при неправильной настройке изоляции транзакций. Длительные транзакции могут блокировать доступ к данным для других транзакций.

Решение: использование таймаутов транзакций, оптимизация запросов, правильный выбор уровня изоляции.

### Веб-серверы

Неправильно настроенные пулы потоков могут привести к тому, что некоторые запросы будут обрабатываться с большой задержкой или вообще не обрабатываться.

```go
func configureWebServer() {
    // Настройка справедливого планировщика запросов
    server := &http.Server{
        Addr:         ":8080",
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
        Handler:      fairRequestHandler(),
    }
    
    server.ListenAndServe()
}

func fairRequestHandler() http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
        defer cancel()
        
        // Обработка запроса с учётом приоритета и времени ожидания
        processRequestWithFairness(ctx, w, r)
    })
}
```

### Микросервисная архитектура

В микросервисных системах голодание может возникать при неправильной настройке балансировщиков нагрузки или при каскадных сбоях.

Решения:
- Реализация паттерна Circuit Breaker
- Правильная настройка таймаутов и повторных попыток
- Мониторинг времени ответа и автоматическое масштабирование

```go
// Реализация Circuit Breaker для предотвращения голодания в микросервисах
type CircuitBreaker struct {
    mu           sync.Mutex
    failureCount int
    lastFailure  time.Time
    threshold    int
    timeout      time.Duration
    state        string // "closed", "open", "half-open"
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    cb.mu.Lock()
    state := cb.state
    
    if state == "open" {
        if time.Since(cb.lastFailure) > cb.timeout {
            cb.state = "half-open"
            state = "half-open"
        } else {
            cb.mu.Unlock()
            return fmt.Errorf("circuit breaker открыт")
        }
    }
    cb.mu.Unlock()
    
    err := fn()
    
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    if err != nil {
        cb.failureCount++
        cb.lastFailure = time.Now()
        
        if cb.failureCount >= cb.threshold || state == "half-open" {
            cb.state = "open"
        }
        return err
    }
    
    if state == "half-open" {
        cb.state = "closed"
    }
    cb.failureCount = 0
    return nil
}
```

## Отличие голодания от взаимной блокировки (deadlock)

Голодание и взаимная блокировка (deadlock) — это разные проблемы многопоточного программирования:

1. **Взаимная блокировка (Deadlock)** — состояние, когда два или более потоков блокируют друг друга, ожидая ресурсы, удерживаемые другими потоками. Система полностью останавливается.

2. **Голодание (Starvation)** — состояние, когда некоторые потоки не получают доступ к ресурсам, но система в целом продолжает работать.

Пример взаимной блокировки:

```go
func deadlockExample() {
    var mutex1, mutex2 sync.Mutex
    
    // Поток 1
    go func() {
        mutex1.Lock()
        time.Sleep(100 * time.Millisecond) // Имитация работы
        mutex2.Lock() // Блокировка, если Поток 2 уже захватил mutex2
        
        // Никогда не достигнет этой точки при deadlock
        mutex2.Unlock()
        mutex1.Unlock()
    }()
    
    // Поток 2
    go func() {
        mutex2.Lock()
        time.Sleep(100 * time.Millisecond) // Имитация работы
        mutex1.Lock() // Блокировка, если Поток 1 уже захватил mutex1
        
        // Никогда не достигнет этой точки при deadlock
        mutex1.Unlock()
        mutex2.Unlock()
    }()
}
```

Пример голодания:

```go
func starvationExample() {
    var mu sync.Mutex
    
    // Агрессивные потоки
    for i := 0; i < 10; i++ {
        go func() {
            for {
                mu.Lock()
                // Быстрая работа
                mu.Unlock()
            }
        }()
    }
    
    // Страдающий от голодания поток
    go func() {
        for {
            time.Sleep(1 * time.Millisecond) // Небольшая задержка
            mu.Lock()
            // Длительная работа
            mu.Unlock()
        }
    }()
}
```

## Заключение

Голодание — серьёзная проблема в многопоточных системах, которая может привести к непредсказуемому поведению, снижению производительности и даже частичной неработоспособности системы. В отличие от взаимной блокировки, голодание сложнее обнаружить, так как система продолжает функционировать, хотя и неэффективно.

Основные методы борьбы с голоданием включают:
- Реализацию справедливых механизмов блокировки
- Использование таймаутов и контекстов с отменой
- Учёт времени ожидания при планировании задач
- Ограничение агрессивного поведения потоков
- Мониторинг времени выполнения операций

При разработке многопоточных приложений на Go необходимо учитывать возможность голодания и применять соответствующие методы для его предотвращения, особенно в системах с высокой нагрузкой и критичными требованиями к отзывчивости.
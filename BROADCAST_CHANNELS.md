#broadcast
# 📢 Broadcast вещание в Go с помощью каналов

Данный репозиторий демонстрирует различные подходы к реализации broadcast (широковещательного) вещания в Go с использованием каналов и горутин.

```table-of-contents
```

## 🎯 Что такое Broadcast вещание?

Broadcast вещание — это паттерн, при котором одно сообщение доставляется одновременно множеству получателей. В Go это можно реализовать несколькими способами, каждый из которых имеет свои преимущества и недостатки.

## 📋 Реализованные подходы

### 1. SimpleBroadcaster - Простая реализация

**Файл:** `broadcast_simple.go`

Базовая реализация через центральную горутину и каналы управления.

#### Особенности:
- ✅ Простая в понимании архитектура
- ✅ Безопасное добавление/удаление слушателей
- ✅ Обработка медленных получателей
- ❌ Использует больше горутин
- ❌ Может быть менее эффективной при большом количестве слушателей

#### Ключевые компоненты:
```go
type SimpleBroadcaster struct {
    input          chan interface{}           // Входной канал для сообщений
    listeners      []chan interface{}         // Список каналов слушателей
    addListener    chan chan interface{}      // Канал для добавления слушателей
    removeListener chan chan interface{}      // Канал для удаления слушателей
    mu             sync.RWMutex               // Мьютекс для защиты доступа
    closed         bool                       // Флаг закрытия
}
```

#### Применение:
- Простые системы уведомлений
- Прототипирование
- Обучение паттернам Go

### 2. CondBroadcaster - С использованием sync.Cond

**Файл:** `broadcast_cond.go`

Более эффективная реализация через условную переменную sync.Cond.

#### Особенности:
- ✅ Более эффективное использование ресурсов
- ✅ Встроенная поддержка контекста для отмены
- ✅ Меньше горутин
- ❌ Сложнее в понимании
- ❌ Требует аккуратной работы с мьютексами

#### Ключевые компоненты:
```go
type CondBroadcaster struct {
    mu        sync.RWMutex                    // Мьютекс для синхронизации
    cond      *sync.Cond                     // Условная переменная
    listeners map[int]chan interface{}       // Карта слушателей
    nextID    int                           // Следующий ID слушателя
    closed    bool                          // Флаг закрытия
    lastMsg   interface{}                   // Последнее сообщение
    hasMsg    bool                          // Флаг наличия сообщения
}
```

#### Применение:
- Высоконагруженные системы
- Когда важна производительность
- Системы с множественными получателями

### 3. TypedBroadcaster - Типизированная реализация

**Файл:** `broadcast_generic.go`

Современная реализация с использованием generics для типобезопасности.

#### Особенности:
- ✅ Типобезопасность на этапе компиляции
- ✅ Поддержка таймаутов
- ✅ Селективное вещание (на конкретных слушателей)
- ✅ Простое API
- ❌ Требует Go 1.18+

#### Ключевые компоненты:
```go
type TypedBroadcaster[T any] struct {
    mu        sync.RWMutex       // Мьютекс для синхронизации
    listeners map[string]chan T  // Карта именованных слушателей
    closed    bool               // Флаг закрытия
}
```

#### Дополнительные возможности:
- `BroadcastWithTimeout()` - вещание с таймаутом
- `BroadcastToSpecific()` - вещание конкретным слушателям
- `GetListenerIDs()` - получение списка активных слушателей

#### Применение:
- Современные Go приложения (Go 1.18+)
- Когда важна типобезопасность
- Системы с различными типами сообщений

## 🏗️ Архитектурные паттерны

### Fan-out паттерн
Все реализации используют fan-out паттерн: одно сообщение распространяется на множество получателей.

```
Отправитель → Broadcaster → Получатель 1
                        → Получатель 2
                        → Получатель 3
                        → ...
```

### Обработка медленных получателей

Все реализации предусматривают обработку ситуаций, когда один из получателей обрабатывает сообщения медленнее других:

1. **Буферизованные каналы** - предотвращают блокировку
2. **Non-blocking send** - используется `select` с `default` case
3. **Таймауты** - автоматическое удаление неотвечающих слушателей

### Управление жизненным циклом

Каждая реализация предоставляет методы для:
- Подписки на события (`Subscribe`)
- Отписки (`Unsubscribe`)
- Отправки сообщений (`Send/Broadcast`)
- Корректного закрытия (`Close`)

## 🔄 Сравнение производительности

| Характеристика | SimpleBroadcaster | CondBroadcaster | TypedBroadcaster |
|----------------|-------------------|-----------------|------------------|
| Сложность | Низкая | Средняя | Низкая |
| Производительность | Средняя | Высокая | Средняя |
| Память | Больше | Меньше | Средняя |
| Типобезопасность | Нет | Нет | Да |
| Поддержка контекста | Частичная | Полная | Нет |

## 💡 Практические советы

### Выбор реализации

1. **SimpleBroadcaster** - для обучения и простых случаев
2. **CondBroadcaster** - для высоконагруженных систем
3. **TypedBroadcaster** - для современных типобезопасных приложений

### Лучшие практики

1. **Всегда закрывайте broadcaster** при завершении работы
2. **Используйте буферизованные каналы** для слушателей
3. **Обрабатывайте медленных получателей** через таймауты или неблокирующую отправку
4. **Используйте контекст** для отмены операций
5. **Тестируйте на race conditions** с помощью `go run -race`

### Альтернативные подходы

Кроме представленных реализаций, в Go можно использовать:

1. **sync/broadcast** - внешние библиотеки
2. **Каналы каналов** - `chan chan T`
3. **sync.Map** - для dynamic подписчиков
4. **Pub/Sub библиотеки** - NATS, Redis Pub/Sub

## 📚 Дополнительные ресурсы

- [Go Concurrency Patterns](https://blog.golang.org/go-concurrency-patterns-timing-out-and)
- [Advanced Go Concurrency Patterns](https://blog.golang.org/advanced-go-concurrency-patterns)
- [Effective Go - Concurrency](https://golang.org/doc/effective_go.html#concurrency)

## 🤝 Применение в реальных проектах

Данные паттерны широко используются в:

- **WebSocket серверах** - для отправки сообщений всем подключенным клиентам
- **Системах уведомлений** - рассылка push-уведомлений
- **Event-driven архитектурах** - распространение событий
- **Мониторинге и логировании** - отправка метрик множественным получателям
- **Game серверах** - синхронизация состояния игры между игроками 
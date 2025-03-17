#readYourWrites #consistency #distributedSystems #database #clientCentricConsistency #eventualConsistency #strongConsistency #replication #CAPtheorem

# Консистентность чтения своих записей

```table-of-contents
```

## Общее описание

Консистентность чтения своих записей (Read-your-writes Consistency, RYWC) — это один из видов [[консистентности]], ориентированный на клиента, в отличие от консистентности, ориентированной на данные. Она гарантирует, что если пользователь записывает данные, то последующие чтения этого пользователя будут отражать эту запись, даже если система распределенная и данные реплицированы на несколько узлов. Другими словами, пользователь всегда видит свои собственные изменения. Это более слабая гарантия, чем строгая консистентность, но более сильная, чем [[согласованность в конечном счёте]] (Eventual Consistency).

## Принцип работы

Рассмотрим принцип работы RYWC на примере распределенной системы с несколькими репликами данных.

1.  **Запись данных:** Клиент отправляет запрос на запись данных на один из узлов (реплик). Этот узел может быть как основным (primary/master), так и вторичным (secondary/replica).

2.  **Распространение записи (Replication):** После успешной записи на узле, изменения распространяются на другие реплики. Этот процесс может быть синхронным (запись считается завершенной только после подтверждения от всех или большинства реплик) или асинхронным (запись считается завершенной сразу после локальной записи, а репликация происходит в фоновом режиме). RYWC чаще всего подразумевает асинхронную репликацию, так как синхронная репликация ближе к строгой консистентности.

3.  **Чтение данных:** Когда клиент отправляет запрос на чтение, система должна гарантировать, что он увидит свои последние записи. Это достигается несколькими способами:
    *   **Sticky Sessions (прилипчивые сессии):** Клиент "привязывается" к определенной реплике на время сессии. Все его запросы (как на чтение, так и на запись) направляются на эту реплику. Это гарантирует RYWC, так как клиент всегда читает из той же реплики, куда он записывал. Недостаток: теряется преимущество распределения нагрузки, и в случае отказа узла, к которому привязан клиент, сессия может быть потеряна.
    *   **Version Vectors (векторные часы):** Каждая запись получает версию (например, используя [[векторные часы]]). Клиент отслеживает версию своих последних записей. При чтении клиент передает эту версию, и система ищет реплику, которая содержит данные с этой или более новой версией. Это более сложный, но и более гибкий подход.
    *   **Read from Primary (чтение с основного узла):** Все записи направляются на основной узел, а чтения могут выполняться как с основного, так и с вторичных узлов. При чтении с вторичного узла, система проверяет, была ли реплицирована последняя запись клиента на этот узел. Если нет, запрос перенаправляется на основной узел или другую реплику, где данные уже доступны.

## Пример

Предположим, у нас есть распределенная база данных с тремя репликами (A, B, C). Пользователь Alice делает следующие действия:

1.  Alice записывает новое сообщение в свой блог. Запись попадает на реплику A.
2.  Репликация на B и C происходит асинхронно.
3.  Alice обновляет страницу своего блога.

Если система *не* поддерживает RYWC, запрос на чтение может попасть на реплику B или C, которые еще не получили обновление. Alice не увидит свое новое сообщение.

Если система поддерживает RYWC (например, с использованием sticky sessions), запрос на чтение будет направлен на реплику A, где уже есть запись Alice. Она увидит свое новое сообщение.

## Сравнение с другими видами консистентности

*   **Строгая консистентность (Strong Consistency):** Гарантирует, что все клиенты видят одну и ту же версию данных в любой момент времени. Все операции выглядят так, как будто они выполняются последовательно на одном узле. RYWC - более слабая гарантия. RYWC гарантирует консистентность только для одного клиента, а строгая консистентность - для всех.
*   **Согласованность в конечном счёте (Eventual Consistency):** Самая слабая гарантия. Гарантирует, что в конечном итоге (после некоторого периода времени без новых записей) все реплики придут в согласованное состояние. RYWC - более сильная гарантия. RYWC гарантирует, что клиент увидит свои собственные записи сразу, а eventual consistency не дает таких гарантий.
*  **Монотонное чтение (Monotonic Reads):** Гарантирует, что если клиент прочитал определенную версию данных, то последующие чтения не вернут более старую версию. RYWC подразумевает monotonic reads, но не наоборот.
*  **Монотонные записи (Monotonic Writes):** Гарантирует, что записи клиента будут применяться в том же порядке, в котором они были сделаны. RYWC не обязательно гарантирует monotonic writes.

## Преимущества и недостатки

**Преимущества:**

*   **Улучшенный пользовательский опыт:** Пользователь всегда видит свои изменения, что создает ощущение целостности данных.
*   **Проще в реализации, чем строгая консистентность:** Не требует сложной синхронизации между репликами.
*    Баланс между консистентностью и производительностью.

**Недостатки:**

*   **Не гарантирует консистентность между разными клиентами:** Другие пользователи могут не сразу увидеть изменения, сделанные Alice.
*   **Сложность реализации зависит от выбранного механизма:** Version vectors требуют более сложной логики, чем sticky sessions.
*   Может привести к увеличению задержки, если необходимо перенаправить запрос на другую реплику.

## Реализация в Go

Рассмотрим пример реализации Read-your-writes consistency с использованием sticky sessions и in-memory хранилища для простоты.

```go
package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Replica представляет собой реплику данных.
type Replica struct {
	ID      string
	Data    map[string]string
	mu      sync.RWMutex
	Latency time.Duration // Имитация задержки репликации.
}

// NewReplica создает новую реплику.
func NewReplica(id string, latency time.Duration) *Replica {
	return &Replica{
		ID:      id,
		Data:    make(map[string]string),
		Latency: latency,
	}
}

// Write записывает данные в реплику.
func (r *Replica) Write(key, value string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Data[key] = value
	fmt.Printf("Replica %s: Wrote key '%s', value '%s'\n", r.ID, key, value)
}

// Read читает данные из реплики.
func (r *Replica) Read(key string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	value, ok := r.Data[key]
	fmt.Printf("Replica %s: Read key '%s', value '%s', found: %t\n", r.ID, key, value, ok)
	return value, ok
}

// Replicate имитирует репликацию данных из другой реплики.
func (r *Replica) Replicate(source *Replica) {
	time.Sleep(r.Latency) // Имитируем задержку.
	source.mu.RLock()
	defer source.mu.RUnlock()
	r.mu.Lock()
	defer r.mu.Unlock()

    // Копируем только те данные, которых нет в текущей реплике,
    // или если данные в источнике новее (в реальной системе нужна версионность).
	for key, value := range source.Data {
		if _, ok := r.Data[key]; !ok {
			r.Data[key] = value
			fmt.Printf("Replica %s: Replicated key '%s', value '%s' from Replica %s\n", r.ID, key, value, source.ID)
		}
	}
}

// Server представляет собой сервер, обслуживающий запросы.
type Server struct {
	Replicas []*Replica
	Sessions map[string]*Replica // Sticky sessions.
	mu       sync.Mutex
}

// NewServer создает новый сервер.
func NewServer(replicas []*Replica) *Server {
	return &Server{
		Replicas: replicas,
		Sessions: make(map[string]*Replica),
	}
}

// getReplicaForSession возвращает реплику для сессии.
func (s *Server) getReplicaForSession(sessionID string) *Replica {
	s.mu.Lock()
	defer s.mu.Unlock()

	replica, ok := s.Sessions[sessionID]
	if !ok {
		// Выбираем случайную реплику для новой сессии.
		replica = s.Replicas[0] // В реальном приложении нужен механизм выбора (round-robin, consistent hashing, ...).
		s.Sessions[sessionID] = replica
		fmt.Printf("New session %s assigned to Replica %s\n", sessionID, replica.ID)
	}
	return replica
}

// handleWrite обрабатывает запрос на запись.
func (s *Server) handleWrite(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		http.Error(w, "Missing session ID", http.StatusBadRequest)
		return
	}

	key := r.FormValue("key")
	value := r.FormValue("value")

	replica := s.getReplicaForSession(sessionID)
	replica.Write(key, value)

    // Запускаем асинхронную репликацию.
	for _, otherReplica := range s.Replicas {
		if otherReplica != replica {
			go otherReplica.Replicate(replica)
		}
	}

	fmt.Fprintf(w, "OK")
}

// handleRead обрабатывает запрос на чтение.
func (s *Server) handleRead(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		http.Error(w, "Missing session ID", http.StatusBadRequest)
		return
	}

	key := r.FormValue("key")

	replica := s.getReplicaForSession(sessionID)
	value, ok := replica.Read(key)
	if !ok {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, value)
}

func main() {
	replicaA := NewReplica("A", 1*time.Second)
	replicaB := NewReplica("B", 2*time.Second)
	replicaC := NewReplica("C", 3*time.Second)
	replicas := []*Replica{replicaA, replicaB, replicaC}

	server := NewServer(replicas)

	http.HandleFunc("/write", server.handleWrite)
	http.HandleFunc("/read", server.handleRead)

	fmt.Println("Server listening on port 8080")
	http.ListenAndServe(":8080", nil)
}

```

**Описание кода:**

*   **Replica:** Структура, представляющая реплику данных. Содержит `ID`, `Data` (in-memory хранилище), мьютекс `mu` для синхронизации доступа и `Latency` для имитации задержки репликации.
*   **NewReplica:** Конструктор для `Replica`.
*   **Write:** Метод для записи данных в реплику. Использует мьютекс для защиты от одновременного доступа.
*   **Read:** Метод для чтения данных из реплики. Также использует мьютекс.
*   **Replicate:** Метод, имитирующий репликацию данных. Принимает другую реплику в качестве источника. Копирует данные из источника в текущую реплику.
*   **Server:** Структура, представляющая сервер. Содержит список реплик `Replicas` и карту `Sessions` для реализации sticky sessions.
*   **NewServer:** Конструктор для `Server`.
*   **getReplicaForSession:** Метод, возвращающий реплику для заданной сессии. Если сессия новая, выбирает реплику (в данном примере - первую) и привязывает к ней сессию.
*   **handleWrite:** Обработчик HTTP запроса на запись. Получает `sessionID` из заголовка, ключ и значение из тела запроса. Записывает данные в реплику, привязанную к сессии, и запускает асинхронную репликацию на другие реплики.
*   **handleRead:** Обработчик HTTP запроса на чтение. Получает `sessionID` из заголовка и ключ из тела запроса. Читает данные из реплики, привязанной к сессии.
*   **main:** Создает три реплики с разной задержкой. Создает сервер. Регистрирует обработчики для `/write` и `/read`. Запускает HTTP сервер.

**Пример использования:**

1.  Запускаем сервер.
2.  Отправляем запрос на запись, указав `sessionID`:

    ```bash
    curl -X POST -d "key=mykey&value=myvalue" -H "X-Session-ID: 123" http://localhost:8080/write
    ```

    Этот запрос запишет данные (`mykey`, `myvalue`) в реплику, привязанную к сессии `123`. В данном случае, это будет реплика A.
3.  Отправляем запрос на чтение с тем же `sessionID`:

    ```bash
    curl -X GET -d "key=mykey" -H "X-Session-ID: 123" http://localhost:8080/read
    ```

    Этот запрос прочитает данные из реплики, привязанной к сессии `123` (реплика A), и вернет `myvalue`.

4.  Если мы отправим запрос на чтение с другим `sessionID` *до* завершения репликации, мы можем не получить данные:

    ```bash
    curl -X GET -d "key=mykey" -H "X-Session-ID: 456" http://localhost:8080/read
    ```
    Этот запрос может попасть на реплику B или C, которые еще не получили данные, и вернет ошибку "Not found".

5.  После завершения репликации (в данном примере, через 1, 2 и 3 секунды для реплик B и C соответственно), запрос на чтение с любым `sessionID` вернет данные.

Этот пример демонстрирует базовую реализацию Read-your-writes consistency с использованием sticky sessions. В реальных системах используются более сложные механизмы, такие как version vectors, quorum reads/writes, и более надежные способы выбора реплик. Кроме того, in-memory хранилище заменяется на настоящую базу данных.

```old
[[Консистентность чтения своих записей]] (Read-your-writes Consistency)
```
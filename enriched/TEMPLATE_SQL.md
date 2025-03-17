#go #testcontainers #postgres #golang #testing #database #docker #singleton #reuse #migrations

# Переиспользование контейнера PostgreSQL в тестах с Testcontainers-go

```table-of-contents
```

Задача состоит в том, чтобы оптимизировать выполнение тестов, использующих базу данных PostgreSQL, с помощью библиотеки `testcontainers-go`. Текущая реализация создает отдельную базу данных для каждого теста, используя шаблон (`TEMPLATE`) базовой базы данных с миграциями.  Необходимо реализовать переиспользование контейнера PostgreSQL между тестами для сокращения времени выполнения.

## Проблема

Изначально, для каждого теста создавался новый контейнер с PostgreSQL, что занимало значительное время.  Использование `TEMPLATE` в PostgreSQL позволяет быстро создавать копии базы данных, но запуск самого контейнера все еще является узким местом.  Флаг `Reuse` в `GenericContainerRequest` предназначен для переиспользования контейнеров, но он отсутствует в специфической реализации для PostgreSQL (`tcpg.PostgresContainer`).

## Решение

Предлагается использовать паттерн Singleton для управления экземпляром контейнера PostgreSQL.  Это гарантирует, что контейнер будет создан только один раз и переиспользован во всех тестах.  Для синхронизации доступа к контейнеру используется мьютекс (`sync.Mutex`).

### Шаг 1: Определение констант и глобальных переменных

Определяем константы для имени базы данных, пользователя, пароля, имени контейнера и образа Docker.  Также объявляем глобальные переменные для хранения экземпляра контейнера и мьютекса.

```go
const (
	baseDBName     = "base_testdb"
	testDBUser     = "testuser"
	testDBPass     = "testpass"
	containerName  = "reusable-postgres-container"
	containerImage = "postgres:13-alpine"
)

var (
	postgresContainer *tcpg.PostgresContainer
	containerLock     sync.Mutex
)
```

### Шаг 2: Реализация функции `getOrCreatePostgresContainer`

Эта функция отвечает за создание или получение существующего экземпляра контейнера.  Она использует мьютекс для защиты от одновременного доступа из разных горутин (тестов).

```go
func getOrCreatePostgresContainer(ctx context.Context) (*tcpg.PostgresContainer, error) {
	containerLock.Lock()
	defer containerLock.Unlock()

	if postgresContainer != nil {
		return postgresContainer, nil
	}

	req := tc.ContainerRequest{
		Image:        containerImage,
		ExposedPorts: []string{"5432/tcp"},
		Name:         containerName,
		Env: map[string]string{
			"POSTGRES_DB":       baseDBName,
			"POSTGRES_USER":     testDBUser,
			"POSTGRES_PASSWORD": testDBPass,
		},
		WaitingFor: tcwait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(5 * time.Second),
	}

	container, err := tcpg.RunContainer(ctx,
		tcpg.WithDatabase(baseDBName),
		tcpg.WithUsername(testDBUser),
		tcpg.WithPassword(testDBPass),
    tcpg.WithSSLMode("disable"),
		tc.WithContainerRequest(req),
	)

	if err != nil {
		return nil, err
	}

	postgresContainer = container
	return container, nil
}
```

- **`containerLock.Lock()` и `containerLock.Unlock()`:**  Гарантируют, что только одна горутина может получить доступ к коду создания контейнера в любой момент времени.
- **`if postgresContainer != nil`:** Проверяет, был ли контейнер уже создан. Если да, функция возвращает существующий экземпляр.
- **`tc.ContainerRequest`:**  Определяет параметры запуска контейнера, такие как образ, порты, переменные окружения и стратегия ожидания готовности.
- **`tcpg.RunContainer`:**  Запускает контейнер с указанными параметрами.
- **`postgresContainer = container`:**  Сохраняет созданный экземпляр контейнера в глобальной переменной.

### Шаг 3: Использование `getOrCreatePostgresContainer` в тестах

Вместо прямого вызова `tcpg.RunContainer` в каждом тесте, теперь используется функция `getOrCreatePostgresContainer`.

```go
func TestExample(t *testing.T) {
	ctx := context.Background()

	postgres, err := getOrCreatePostgresContainer(ctx)
	require.NoError(t, err, "failed to start or get postgres container")

  // ... остальная часть теста ...
}
```
- **`postgres, err := getOrCreatePostgresContainer(ctx)`:** Получаем (или создаем) контейнер.

### Шаг 4: Использование `pgxpool`

Вместо `pgx` используется `pgxpool` для управления пулом соединений к базе данных, что улучшает производительность при работе с базой данных.

```go
	connString, err := postgres.ConnectionString(ctx)
	require.NoError(t, err, "failed to get connection string")

	basePool, err := pgxpool.New(ctx, connString+" sslmode=disable")
	require.NoError(t, err, "failed to create connection pool")
	defer basePool.Close()
```

### Шаг 5: Использование `TestMain` для управления жизненным циклом контейнера

Функция `TestMain` позволяет выполнить код до и после выполнения всех тестов в пакете.  Это идеальное место для запуска и остановки контейнера.

```go
func TestMain(m *testing.M) {
	ctx := context.Background()

	_, err := getOrCreatePostgresContainer(ctx)
	if err != nil {
		log.Fatalf("Could not start container: %s", err)
	}

	code := m.Run()

	if postgresContainer != nil {
		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Fatalf("Could not stop container: %s", err)
		}
	}

	os.Exit(code)
}
```

- **`m.Run()`:** Запускает все тесты в пакете.
- **`postgresContainer.Terminate(ctx)`:** Останавливает контейнер после выполнения всех тестов.

### Шаг 6:  Настройка базовой базы данных (миграции)

Функция `setupBaseDatabase` создает таблицу `users`, если она еще не существует.

```go
func setupBaseDatabase(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	return nil
}
```

### Шаг 7: Создание тестовой базы данных

Функция `createTestDatabase` создает новую базу данных, используя `TEMPLATE` базовой базы данных. Она также удаляет существующую тестовую базу данных перед созданием новой.

```go
func createTestDatabase(ctx context.Context, basePool *pgxpool.Pool, testDBName string) (*pgxpool.Pool, error) {
	_, err := basePool.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
	if err != nil {
		return nil, fmt.Errorf("failed to drop existing test database: %w", err)
	}

	_, err = basePool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s WITH TEMPLATE %s", testDBName, baseDBName))
	if err != nil {
		return nil, fmt.Errorf("failed to create test database: %w", err)
	}

	connConfig, err := pgxpool.ParseConfig(basePool.Config().ConnString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	connConfig.ConnConfig.Database = testDBName
	testPool, err := pgxpool.ConnectConfig(ctx, connConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	return testPool, nil
}
```

-  **`DROP DATABASE IF EXISTS %s`**:  Удаляет базу данных с таким же именем, если она существует.  Это необходимо для обеспечения чистоты тестовой среды.
-  **`CREATE DATABASE %s WITH TEMPLATE %s`**: Создает новую базу данных, используя базовую базу данных в качестве шаблона.
- **`pgxpool.ParseConfig` and `pgxpool.ConnectConfig`**:  Создает конфигурацию и подключается к новой базе данных.

### Шаг 8:  Использование `require` из `testify`

Библиотека `testify/require` используется для упрощения обработки ошибок в тестах. Вместо `t.Fatalf` используется `require.NoError`, что делает код более читаемым.

### Шаг 9: Упрощение настройки миграций (предложение)

Вместо ручного создания таблицы `users` в функции `setupBaseDatabase`, можно использовать параметр `tcpg.WithInitScripts`. Он позволяет указать SQL-скрипт, который будет выполнен при создании контейнера. Это упрощает настройку базы данных и делает код более декларативным.

Пример:

1.  Создайте файл `testdata/dev-db.sql` со следующим содержимым:

    ```sql
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        name TEXT NOT NULL,
        email TEXT UNIQUE NOT NULL
    );
    ```

2.  Измените функцию `getOrCreatePostgresContainer`, добавив `tcpg.WithInitScripts`:

    ```go
    container, err := tcpg.RunContainer(ctx,
      tcpg.WithDatabase(baseDBName),
      tcpg.WithUsername(testDBUser),
      tcpg.WithPassword(testDBPass),
      tcpg.WithSSLMode("disable"),
      tc.WithContainerRequest(req),
      tcpg.WithInitScripts(filepath.Join(".", "testdata", "dev-db.sql")), // Добавлено
    )
    ```

3.  Удалите функцию `setupBaseDatabase` и ее вызов из `TestExample`.

### Шаг 10: Использование tmpfs и отключение fsync (упомянуто в вопросе, но не реализовано в коде)

Для дальнейшего ускорения тестов можно использовать `tmpfs` и отключить `fsync`.

- **`tmpfs`:**  Монтирует директорию базы данных в оперативную память, что значительно ускоряет операции ввода-вывода.
- **`fsync`:**  Отключение `fsync` предотвращает сброс данных на диск после каждой транзакции.  Это рискованно для production-сред, но может быть приемлемо для тестов, где потеря данных не критична.

Эти опции можно настроить в конфигурации PostgreSQL, но это выходит за рамки текущей задачи.

## Полный код

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcwait "github.com/testcontainers/testcontainers-go/wait"
)

const (
	baseDBName     = "base_testdb"
	testDBUser     = "testuser"
	testDBPass     = "testpass"
	containerName  = "reusable-postgres-container"
	containerImage = "postgres:13-alpine"
)

var (
	postgresContainer *tcpg.PostgresContainer
	containerLock     sync.Mutex
)

func getOrCreatePostgresContainer(ctx context.Context) (*tcpg.PostgresContainer, error) {
	containerLock.Lock()
	defer containerLock.Unlock()

	if postgresContainer != nil {
		return postgresContainer, nil
	}

	req := tc.ContainerRequest{
		Image:        containerImage,
		ExposedPorts: []string{"5432/tcp"},
		Name:         containerName,
		Env: map[string]string{
			"POSTGRES_DB":       baseDBName,
			"POSTGRES_USER":     testDBUser,
			"POSTGRES_PASSWORD": testDBPass,
		},
		WaitingFor: tcwait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(5 * time.Second),
	}

	container, err := tcpg.RunContainer(ctx,
		tcpg.WithDatabase(baseDBName),
		tcpg.WithUsername(testDBUser),
		tcpg.WithPassword(testDBPass),
		tcpg.WithSSLMode("disable"),
		tc.WithContainerRequest(req),
        // Раскомментируйте следующую строку и создайте файл testdata/dev-db.sql, чтобы использовать WithInitScripts.
        // tcpg.WithInitScripts(filepath.Join(".", "testdata", "dev-db.sql")),
	)

	if err != nil {
		return nil, err
	}

	postgresContainer = container
	return container, nil
}

func setupBaseDatabase(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	return nil
}

func createTestDatabase(ctx context.Context, basePool *pgxpool.Pool, testDBName string) (*pgxpool.Pool, error) {
	_, err := basePool.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
	if err != nil {
		return nil, fmt.Errorf("failed to drop existing test database: %w", err)
	}

	_, err = basePool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s WITH TEMPLATE %s", testDBName, baseDBName))
	if err != nil {
		return nil, fmt.Errorf("failed to create test database: %w", err)
	}

	connConfig, err := pgxpool.ParseConfig(basePool.Config().ConnString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	connConfig.ConnConfig.Database = testDBName
	testPool, err := pgxpool.ConnectConfig(ctx, connConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	return testPool, nil
}

func TestExample(t *testing.T) {
	ctx := context.Background()

	postgres, err := getOrCreatePostgresContainer(ctx)
	require.NoError(t, err, "failed to start or get postgres container")

	connString, err := postgres.ConnectionString(ctx)
	require.NoError(t, err, "failed to get connection string")

	basePool, err := pgxpool.New(ctx, connString+" sslmode=disable")
	require.NoError(t, err, "failed to create connection pool")
	defer basePool.Close()

    // Раскомментируйте следующую строку, если используете WithInitScripts.
	// err = setupBaseDatabase(ctx, basePool)
	// require.NoError(t, err, "failed to setup base database")

	testDBName := fmt.Sprintf("%s_%s", baseDBName, t.Name())
	testPool, err := createTestDatabase(ctx, basePool, testDBName)
	require.NoError(t, err, "failed to create test database")
	defer testPool.Close()

	_, err = testPool.Exec(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)", "John Doe", "john@example.com")
	require.NoError(t, err, "failed to insert user")

	var count int
	err = testPool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	require.NoError(t, err, "failed to count users")

	require.Equal(t, 1, count, "Expected 1 user, got %d", count)
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	_, err := getOrCreatePostgresContainer(ctx)
	if err != nil {
		log.Fatalf("Could not start container: %s", err)
	}

	code := m.Run()

	if postgresContainer != nil {
		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Fatalf("Could not stop container: %s", err)
		}
	}

	os.Exit(code)
}

func main() {
	log.Println("This is a test file. Run 'go test' to execute the tests.")
}
```

## Выводы

Реализовано переиспользование контейнера PostgreSQL между тестами с помощью паттерна Singleton и мьютекса.  Это значительно сокращает время выполнения тестов, так как контейнер запускается только один раз.  Использование `pgxpool` улучшает производительность работы с базой данных.  Предложено упрощение настройки миграций с помощью `tcpg.WithInitScripts`.  Упомянуты дополнительные возможности оптимизации (`tmpfs` и `fsync`), которые могут быть реализованы в будущем.  Использование  `testify/require`  делает код тестов более читаемым.

```old
Мы в тестконтейнерах поднимаем субд, создаем одну бд с миграцичми, а потом на каждый тест создаём свою бд, используя TEMPLATE в постгре и все тесты параллельно гоняем на изолированных бд. Считай юнит тесты

Еще можно юзать темпфс, чтобы не тратить время для записи на диск, и отключить у постгреса fsync, что мы нашли допустимым для тестов, и тогда время выполнения тестов, почти как с моками, а качество и надежность — в разы выше

\`\`\`go
package main

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	tc "github.com/testcontainers/testcontainers-go"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcwait "github.com/testcontainers/testcontainers-go/wait"
)

const (
	baseDBName = "base_testdb"
	testDBUser = "testuser"
	testDBPass = "testpass"
)

func runContainer(ctx context.Context) (*tcpg.PostgresContainer, error) {
	return tcpg.RunContainer(ctx,
		tcpg.WithDatabase(baseDBName),
		tcpg.WithUsername(testDBUser),
		tcpg.WithPassword(testDBPass),
		tc.WithWaitStrategy(
			tcwait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
}

func setupBaseDatabase(ctx context.Context, connString string) error {
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close(ctx)

	// Выполняем миграции или создаем необходимую структуру в базовой БД
	_, err = conn.Exec(ctx, `
		CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	return nil
}

func createTestDatabase(ctx context.Context, baseConnString, testDBName string) (string, error) {
	conn, err := pgx.Connect(ctx, baseConnString)
	if err != nil {
		return "", fmt.Errorf("failed to connect to base database: %w", err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s WITH TEMPLATE %s", testDBName, baseDBName))
	if err != nil {
		return "", fmt.Errorf("failed to create test database: %w", err)
	}

	// Формируем новую строку подключения для тестовой БД
	testConnString := fmt.Sprintf("%s dbname=%s", baseConnString[:len(baseConnString)-len(baseDBName)], testDBName)
	return testConnString, nil
}

func TestExample(t *testing.T) {
	ctx := context.Background()

	postgres, err := runContainer(ctx)
	if err != nil {
		t.Fatalf("failed to start postgres container: %s", err)
	}
	defer func() {
		if err := postgres.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	connString, err := postgres.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get connection string: %s", err)
	}

	if err := setupBaseDatabase(ctx, connString); err != nil {
		t.Fatalf("failed to setup base database: %s", err)
	}

	// Создаем тестовую базу данных
	testDBName := fmt.Sprintf("%s_%s", baseDBName, t.Name())
	testConnString, err := createTestDatabase(ctx, connString, testDBName)
	if err != nil {
		t.Fatalf("failed to create test database: %s", err)
	}

	// Подключаемся к тестовой базе данных
	conn, err := pgx.Connect(ctx, testConnString)
	if err != nil {
		t.Fatalf("failed to connect to test database: %s", err)
	}
	defer conn.Close(ctx)

	// Выполняем тестовые операции
	_, err = conn.Exec(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)", "John Doe", "john@example.com")
	if err != nil {
		t.Fatalf("failed to insert user: %s", err)
	}

	var count int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count users: %s", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 user, got %d", count)
	}
}

func main() {
	log.Println("This is a test file. Run 'go test' to execute the tests.")
}
\`\`\`
но как нам переиспользовать контейнер?

\`\`\`go
genericContainerReq := testcontainers.GenericContainerRequest{
  ContainerRequest: req,
  Started:          true,
  Logger:           logger,
  Reuse:            true, // правильный ответ
}
\`\`\`

только флага `Reuse` нет в модуле Postgres, приходится изобретать синглтон:

\`\`\`go
const (
	baseDBName     = "base_testdb"
	testDBUser     = "testuser"
	testDBPass     = "testpass"
	containerName  = "reusable-postgres-container"
	containerImage = "postgres:13-alpine"
)

var (
	postgresContainer *tcpg.PostgresContainer
	containerLock     sync.Mutex
)

func getOrCreatePostgresContainer(ctx context.Context) (*tcpg.PostgresContainer, error) {
	containerLock.Lock()
	defer containerLock.Unlock()

	if postgresContainer != nil {
		return postgresContainer, nil
	}

	req := tc.ContainerRequest{
		Image:        containerImage,
		ExposedPorts: []string{"5432/tcp"},
		Name:         containerName,
		Env: map[string]string{
			"POSTGRES_DB":       baseDBName,
			"POSTGRES_USER":     testDBUser,
			"POSTGRES_PASSWORD": testDBPass,
		},
		WaitingFor: tcwait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(5 * time.Second),
	}

	container, err := tcpg.RunContainer(ctx,
		tcpg.WithDatabase(baseDBName),
		tcpg.WithUsername(testDBUser),
		tcpg.WithPassword(testDBPass),
    tcpg.WithSSLMode("disable"),
		tc.WithContainerRequest(req),
	)

	if err != nil {
		return nil, err
	}

	postgresContainer = container
	return container, nil
}
\`\`\`

Ожидаю: https://github.com/testcontainers/testcontainers-go/issues/2726

Пока так:

\`\`\`go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcwait "github.com/testcontainers/testcontainers-go/wait"
)

const (
	baseDBName     = "base_testdb"
	testDBUser     = "testuser"
	testDBPass     = "testpass"
	containerName  = "reusable-postgres-container"
	containerImage = "postgres:13-alpine"
)

var (
	postgresContainer *tcpg.PostgresContainer
	containerLock     sync.Mutex
)

func getOrCreatePostgresContainer(ctx context.Context) (*tcpg.PostgresContainer, error) {
	containerLock.Lock()
	defer containerLock.Unlock()

	if postgresContainer != nil {
		return postgresContainer, nil
	}

	req := tc.ContainerRequest{
		Image:        containerImage,
		ExposedPorts: []string{"5432/tcp"},
		Name:         containerName,
		Env: map[string]string{
			"POSTGRES_DB":       baseDBName,
			"POSTGRES_USER":     testDBUser,
			"POSTGRES_PASSWORD": testDBPass,
		},
		WaitingFor: tcwait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(5 * time.Second),
	}

	container, err := tcpg.RunContainer(ctx,
		tcpg.WithDatabase(baseDBName),
		tcpg.WithUsername(testDBUser),
		tcpg.WithPassword(testDBPass),
    tcpg.WithSSLMode("disable"),
    tc.WithContainerRequest(req),
	)

	if err != nil {
		return nil, err
	}

	postgresContainer = container
	return container, nil
}

func setupBaseDatabase(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	return nil
}

func createTestDatabase(ctx context.Context, basePool *pgxpool.Pool, testDBName string) (*pgxpool.Pool, error) {
	_, err := basePool.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
	if err != nil {
		return nil, fmt.Errorf("failed to drop existing test database: %w", err)
	}

	_, err = basePool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s WITH TEMPLATE %s", testDBName, baseDBName))
	if err != nil {
		return nil, fmt.Errorf("failed to create test database: %w", err)
	}

	connConfig, err := pgxpool.ParseConfig(basePool.Config().ConnString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	connConfig.ConnConfig.Database = testDBName
	testPool, err := pgxpool.ConnectConfig(ctx, connConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	return testPool, nil
}

func TestExample(t *testing.T) {
	ctx := context.Background()

	postgres, err := getOrCreatePostgresContainer(ctx)
	require.NoError(t, err, "failed to start or get postgres container")

	connString, err := postgres.ConnectionString(ctx)
	require.NoError(t, err, "failed to get connection string")

	basePool, err := pgxpool.New(ctx, connString+" sslmode=disable")
	require.NoError(t, err, "failed to create connection pool")
	defer basePool.Close()

	err = setupBaseDatabase(ctx, basePool)
	require.NoError(t, err, "failed to setup base database")

	testDBName := fmt.Sprintf("%s_%s", baseDBName, t.Name())
	testPool, err := createTestDatabase(ctx, basePool, testDBName)
	require.NoError(t, err, "failed to create test database")
	defer testPool.Close()

	_, err = testPool.Exec(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)", "John Doe", "john@example.com")
	require.NoError(t, err, "failed to insert user")

	var count int
	err = testPool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	require.NoError(t, err, "failed to count users")

	require.Equal(t, 1, count, "Expected 1 user, got %d", count)
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	
	_, err := getOrCreatePostgresContainer(ctx)
	if err != nil {
		log.Fatalf("Could not start container: %s", err)
	}

	code := m.Run()

	if postgresContainer != nil {
		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Fatalf("Could not stop container: %s", err)
		}
	}

	os.Exit(code)
}

func main() {
	log.Println("This is a test file. Run 'go test' to execute the tests.")
}
\`\`\`

у твоего tcpg есть такая настройка `tcpg.WithInitScripts(filepath.Join(".", "testdata", "dev-db.sql"))`, Чтобы руками в тестах не делать setupBaseDatabase
```
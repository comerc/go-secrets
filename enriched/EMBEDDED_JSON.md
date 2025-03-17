#golang #go #json #marshal #unmarshal #embedded #fields #anonymous #struct #workaround #time

# Обход встраивания полей с реализацией Marshaler/Unmarshaler

```table-of-contents
```

## Постановка задачи

Рассмотрим проблему обработки встроенных полей в Go, особенно когда эти поля реализуют интерфейсы `Marshaler` и `Unmarshaler`. Встроенные поля (anonymous fields) в Go предоставляют удобный способ наследования поведения и данных, но могут создавать сложности при сериализации/десериализации данных, если логика сериализации/десериализации по умолчанию не подходит.

Пример кода демонстрирует эту проблему:

```go
package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Event struct {
	ID        int
	time.Time // "embedded"
}

// type Marshaler interface {
// 	MarshalJSON() ([]byte, error)
// }

func (e Event) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		struct {
			ID   int
			Time time.Time
		}{
			ID:   e.ID,
			Time: e.Time,
		},
	)
}

func main() {
	event := Event{ID: 1234, Time: time.Now()}
	fmt.Printf("%+v", event)
	b, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
```

В этом коде структура `Event` содержит встроенное поле `time.Time`. `time.Time` по умолчанию сериализуется в формате RFC3339Nano. Однако, если мы хотим управлять сериализацией `Event`, возникает проблема - переопределение `MarshalJSON` для `Event` не переопределяет сериализацию встроенного поля `time.Time`. В примере приведён способ решения, но он не оптимален.

## Анализ проблемы и подходы к решению

Проблема заключается в том, что при маршалинге в JSON встроенного поля, Go в первую очередь проверяет, реализует ли тип этого поля интерфейс `Marshaler`. Если да, то используется метод `MarshalJSON` этого типа. Если нет, то Go рекурсивно обрабатывает поля структуры.

Существует несколько подходов к решению этой проблемы, каждый из которых имеет свои преимущества и недостатки:

1.  **Использование анонимной структуры внутри `MarshalJSON` (и `UnmarshalJSON`):**  Этот подход показан в исходном примере. Он заключается в создании анонимной структуры внутри методов `MarshalJSON` и `UnmarshalJSON`, которая содержит те же поля, что и исходная структура, но с явным указанием имени для поля `time.Time`.

    *   **Преимущества:**  Полный контроль над процессом сериализации/десериализации.
    *   **Недостатки:**  Дублирование кода (определение полей структуры), что усложняет поддержку.  Если в структуру добавляется новое поле, его нужно добавить и в анонимную структуру в методах `MarshalJSON` и `UnmarshalJSON`.

2.  **Переименование встроенного поля:**  Простейший способ - дать встроенному полю имя.  В этом случае поле `time.Time` перестает быть встроенным, и его сериализация/десериализация будет контролироваться методами `MarshalJSON` и `UnmarshalJSON` структуры `Event`.

    ```go
    type Event struct {
    	ID   int
    	Time time.Time // Явное имя поля
    }
    ```

    *   **Преимущества:**  Простота реализации.
    *   **Недостатки:**  Изменяется структура данных.  Если важно сохранить именно встроенное поле (например, для совместимости с другими частями кода), этот подход не подойдет.

3.  **Создание пользовательского типа для `time.Time`:** Можно создать новый тип, обертывающий `time.Time`, и реализовать для него методы `MarshalJSON` и `UnmarshalJSON`.

    ```go
    package main

    import (
    	"encoding/json"
    	"fmt"
    	"time"
    )

    type CustomTime struct {
    	time.Time
    }

    func (ct CustomTime) MarshalJSON() ([]byte, error) {
    	// Своя логика сериализации
    	return json.Marshal(ct.Time.Format(time.RFC3339))
    }

    func (ct *CustomTime) UnmarshalJSON(data []byte) error {
    	// Своя логика десериализации
        var s string
        if err := json.Unmarshal(data, &s); err != nil {
            return err
        }
        t, err := time.Parse(time.RFC3339, s)
        if err != nil {
            return err
        }
        ct.Time = t
        return nil
    }

    type Event struct {
    	ID   int
    	Time CustomTime // Используем свой тип
    }

    func main() {
    	event := Event{ID: 1234, Time: CustomTime{time.Now()}}
    	fmt.Printf("%+v\n", event)
    	b, err := json.Marshal(event)
    	if err != nil {
    		panic(err)
    	}
    	fmt.Println(string(b))

        var event2 Event
        err = json.Unmarshal(b, &event2)
        if err != nil {
            panic(err)
        }
        fmt.Printf("%+v\n", event2)
    }
    ```

    *   **Преимущества:**  Более чистое решение с точки зрения разделения ответственности.  Логика сериализации/десериализации `time.Time` инкапсулирована в отдельном типе.
    *   **Недостатки:**  Необходимость создания нового типа и, возможно, конвертации между `time.Time` и `CustomTime` в других частях кода.

4. **Использование указателя на `time.Time`:** Если поле является указателем, то при маршалинге будет проверено, реализует ли *тип, на который указывает указатель,* интерфейс `Marshaler`. Если указатель равен `nil`, то будет сериализовано `null`.

    ```go
    type Event struct {
        ID   int
        Time *time.Time // Используем указатель
    }

    func (e Event) MarshalJSON() ([]byte, error) {
        if e.Time == nil {
            return json.Marshal(struct {
                ID   int
                Time *time.Time
            }{
                ID: e.ID,
                Time: nil,
            })
        }
        return json.Marshal(struct {
            ID   int
            Time string // Сериализуем как строку в нужном формате
        }{
            ID:   e.ID,
            Time: e.Time.Format(time.RFC3339),
        })
    }

    ```

    *  **Преимущества:** Позволяет сериализовать `null`, если время не установлено.
    *  **Недостатки:** Необходимость работы с указателями, что может усложнить код.

## Подробное описание решения с использованием анонимной структуры

Рассмотрим более детально первый подход с использованием анонимной структуры, так как он предоставляет наибольший контроль и не требует изменения структуры `Event`.

```go
package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Event struct {
	ID        int
	time.Time // "embedded"
}

func (e Event) MarshalJSON() ([]byte, error) {
	// Создаем анонимную структуру с явным именем поля Time.
	aux := struct {
		ID   int       `json:"id"`
		Time time.Time `json:"time"` // Явное имя поля
	}{
		ID:   e.ID,
		Time: e.Time,
	}
	return json.Marshal(aux)
}

func (e *Event) UnmarshalJSON(data []byte) error {
	// Создаем анонимную структуру с явным именем поля Time.
	aux := struct {
		ID   int       `json:"id"`
		Time time.Time `json:"time"` // Явное имя поля
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	e.ID = aux.ID
	e.Time = aux.Time
	return nil
}

func main() {
	event := Event{ID: 1234, Time: time.Now()}
	fmt.Printf("%+v\n", event)
	b, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))

	var event2 Event
	err = json.Unmarshal(b, &event2)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", event2)
}
```

**Шаг 1: Определение структуры `Event`**

Определяем структуру `Event` со встроенным полем `time.Time`.

**Шаг 2: Реализация метода `MarshalJSON`**

Реализуем метод `MarshalJSON` для структуры `Event`. Внутри метода:

1.  Создаем анонимную структуру `aux` с теми же полями, что и `Event`, но с явным именем для поля `Time`. Это ключевой момент, позволяющий обойти стандартное поведение маршалинга для встроенного поля `time.Time`.
2.  Присваиваем полям структуры `aux` значения соответствующих полей структуры `e`.
3.  Используем стандартную функцию `json.Marshal` для сериализации структуры `aux`.

**Шаг 3: Реализация метода `UnmarshalJSON`**

Реализуем метод `UnmarshalJSON` для структуры `Event`. Внутри метода:

1.  Создаем анонимную структуру `aux`, аналогичную той, что используется в `MarshalJSON`.
2.  Используем стандартную функцию `json.Unmarshal` для десериализации данных в структуру `aux`.
3.  Присваиваем полям структуры `e` значения соответствующих полей структуры `aux`.

**Шаг 4: Использование**

В функции `main` создаем экземпляр `Event`, сериализуем его в JSON, выводим результат, затем десериализуем обратно и снова выводим.

## Заключение

Выбор конкретного подхода зависит от требований к коду и ограничений. Если необходимо сохранить структуру `Event` неизменной и при этом полностью контролировать процесс сериализации/десериализации, то лучшим выбором будет использование анонимной структуры внутри методов `MarshalJSON` и `UnmarshalJSON`. Если же структура данных может быть изменена, то более простым и понятным решением будет переименование встроенного поля. Создание пользовательского типа для `time.Time` является компромиссным решением, обеспечивающим баланс между контролем и чистотой кода. Использование указателя на `time.Time` полезно, когда необходимо иметь возможность представлять отсутствие времени как `null`.

```old
как обойти вопрос встроенного поля, в котором реализован Marshaler (Unmrshaler):

\`\`\`go
package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Event struct {
	ID        int
	time.Time // "embedded"
}

// type Marshaler interface {
// 	MarshalJSON() ([]byte, error)
// }

func (e Event) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		struct {
			ID   int
			Time time.Time
		}{
			ID:   e.ID,
			Time: e.Time,
		},
	)
}

func main() {
	event := Event{ID: 1234, Time: time.Now()}
	fmt.Printf("%+v", event)
	b, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
\`\`\`

такое же решение может быть с форматированием строк:

\`\`\`go
func (e Event) String() (string, error) {
	return fmt.Sprint(
		struct {
			ID   int
			Time time.Time
		}{
			ID:   e.ID,
			Time: e.Time,
		},
	), nil
}
\`\`\`

или можно просто добавить имя, чтобы поле time.Time больше не было встроенным:

\`\`\`go
type Event struct {
	ID        int
	Time time.Time // !!!
}
\`\`\`

```
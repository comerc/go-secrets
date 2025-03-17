#recovery #go #testing #panic #recover #defer #assert #testify #tdd #programming

# Обработка паники и восстановление в Go

```table-of-contents
```

Go предоставляет механизм `panic` и `recover` для обработки исключительных ситуаций, которые не могут быть обработаны обычным способом. `panic` останавливает нормальное выполнение функции, а `recover` позволяет перехватить панику и восстановить управление. В этом ответе подробно рассматривается, как использовать `panic` и `recover`, а также как тестировать код, который может вызвать панику, с использованием пакета `testify`.

## Паника (Panic)

`panic` - это встроенная функция, которая останавливает нормальное выполнение текущей горутины. Когда функция вызывает `panic`, ее выполнение немедленно прекращается, любые отложенные функции (`defer`) выполняются, и управление возвращается к вызывающей функции. Этот процесс продолжается вверх по стеку вызовов, пока не будет найдена функция `recover` или пока программа не завершится аварийно.

`panic` обычно используется в ситуациях, когда программа не может продолжать работу из-за неустранимой ошибки. Например, при попытке доступа к несуществующему элементу массива или при возникновении ошибки ввода-вывода, которую невозможно обработать.

Синтаксис:

```go
panic(v interface{})
```

`v` - это значение любого типа, которое описывает причину паники. Обычно это строка с сообщением об ошибке, но может быть и любым другим значением.

Пример:

```go
package main

import "fmt"

func divide(x, y int) int {
	if y == 0 {
		panic("division by zero")
	}
	return x / y
}

func main() {
	fmt.Println(divide(10, 2))
	fmt.Println(divide(10, 0)) // Вызовет панику
	fmt.Println(divide(10, 5)) // Этот код не выполнится
}
```

В этом примере функция `divide` вызывает `panic`, если делитель равен нулю. При вызове `divide(10, 0)` программа завершится с сообщением об ошибке "division by zero".

## Восстановление (Recover)

`recover` - это встроенная функция, которая позволяет перехватить панику и восстановить нормальное выполнение программы. `recover` следует использовать только внутри отложенных функций (`defer`). Если `recover` вызывается вне отложенной функции или если паники не было, `recover` возвращает `nil`. Если паника была перехвачена, `recover` возвращает значение, которое было передано в `panic`.

Синтаксис:

```go
recover() interface{}
```

Пример:

```go
package main

import "fmt"

func safeDivide(x, y int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occurred: %v", r)
		}
	}()

	if y == 0 {
		panic("division by zero")
	}
	result = x / y
	return result, nil
}

func main() {
	result, err := safeDivide(10, 2)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Result:", result)
	}

	result, err = safeDivide(10, 0)
	if err != nil {
		fmt.Println(err) // Выведет "panic occurred: division by zero"
	} else {
		fmt.Println("Result:", result)
	}
}
```

В этом примере функция `safeDivide` использует `defer` и `recover` для перехвата паники, вызванной делением на ноль.  Если паника происходит, отложенная функция перехватывает ее и возвращает ошибку. Это позволяет вызывающей функции `main` корректно обработать ошибку без аварийного завершения программы.

## Тестирование паники с помощью testify/assert

Пакет `testify/assert` предоставляет удобные функции для тестирования кода, который может вызвать панику.

### `assert.NotPanics`

Эта функция проверяет, что переданная функция *не* вызывает панику.

```go
func TestSomethingThatMightPanic(t *testing.T) {
    assert.NotPanics(t, func() {
        // код, который не должен вызывать панику
        // Например:
        result := 10 / 2
        if result != 5 {
            panic("Unexpected result")
        }
    }, "This code should not panic")
}
```

### `assert.Panics`

Эта функция проверяет, что переданная функция *вызывает* панику.

```go
func TestSomethingThatShouldPanic(t *testing.T) {
    assert.Panics(t, func() {
        // код, который должен вызвать панику
		panic("expected panic")
    }, "This code should panic")
}
```

### `assert.PanicsWithValue`

Эта функция проверяет, что переданная функция вызывает панику с *определенным значением*.

```go
func TestPanicMessage(t *testing.T) {
    assert.PanicsWithValue(t, "expected message", func() {
        panic("expected message")
    }, "The panic message should be 'expected message'")
}
```
Эта функция принимает ожидаемое значение паники, функцию, которая должна вызвать панику, и необязательное сообщение об ошибке, которое выводится, если тест не пройден.

### Более сложный пример с использованием `assert.PanicsWithError`

Существует также функция `assert.PanicsWithError`, которая проверяет, что паника произошла с ожидаемой *ошибкой*. Это полезно, когда в качестве значения паники используется ошибка (`error`).

```go
import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPanicsWithError(t *testing.T) {
	expectedErr := errors.New("expected error")

	assert.PanicsWithError(t, "expected error", func() {
		panic(expectedErr)
	}, "The panic error should be 'expected error'")
}

```

В этом примере мы создаем ожидаемую ошибку `expectedErr` и проверяем, что функция, переданная в `assert.PanicsWithError`, вызывает панику именно с этой ошибкой.

## Подробный пример использования `panic` и `recover`

Рассмотрим пример обработки ошибок при работе с файлами.

```go
package main

import (
	"fmt"
	"os"
)

func readFile(filename string) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from panic: %v\n", r)
		}
	}()

	file, err := os.Open(filename)
	if err != nil {
		panic(fmt.Sprintf("failed to open file: %v", err)) // Паника с подробным сообщением
	}
	defer file.Close()

	data := make([]byte, 100) // Предполагаем, что файл не больше 100 байт
	n, err := file.Read(data)
	if err != nil && err != os.ErrClosed { // Игнорируем ошибку, если файл уже закрыт.
        panic(fmt.Sprintf("Failed to read file: %v", err))
    }

	return data[:n], nil
}

func main() {
	content, err := readFile("nonexistent.txt")
	if err != nil {
		fmt.Println("Error:", err) // Это не сработает, т.к. ошибка обрабатывается через panic
	}
	fmt.Println(content)

	content, err = readFile("existing.txt") // Предположим, что такой файл существует
	if err != nil {
		fmt.Println("Error:", err)  // Это не сработает, т.к. ошибка обрабатывается через panic
	}
	fmt.Println(string(content))
}
```

В этом примере:

1.  Функция `readFile` пытается открыть файл и прочитать его содержимое.
2.  Если возникают ошибки при открытии или чтении файла, вызывается `panic` с подробным сообщением об ошибке.
3.  Отложенная функция с помощью `recover` перехватывает панику и выводит сообщение в консоль.
4.  Функция `main` вызывает `readFile` дважды: с несуществующим файлом и с существующим. Обратите внимание, что стандартный способ обработки ошибок с помощью `if err != nil` не работает, если ошибка "возвращается" с помощью `panic`.

Этот пример демонстрирует, что `panic` и `recover` могут использоваться для обработки ошибок, но это не самый лучший способ.  В большинстве случаев лучше возвращать ошибки явно с помощью `error`.  `panic` следует использовать только в действительно исключительных ситуациях, когда продолжение работы программы невозможно.

## Заключение

`panic` и `recover` - мощный, но опасный механизм. Его следует использовать с осторожностью, только в ситуациях, когда программа не может продолжать работу. Тестирование кода, который может вызвать панику, важно для обеспечения надежности приложения. Пакет `testify/assert` предоставляет удобные средства для такого тестирования. Важно помнить, что `recover` работает только внутри отложенных функций (`defer`).

```old
\`\`\`go
// Пример тестирования recovery
func TestSomethingThatMightPanic(t *testing.T) {
    assert.NotPanics(t, func() {
        // код, который может вызвать панику
    })

    // или для проверки конкретной паники
    assert.Panics(t, func() {
        panic("ожидаемая паника")
    })
}

// Более сложный пример с проверкой сообщения паники
func TestPanicMessage(t *testing.T) {
    assert.PanicsWithValue(t, "ожидаемое сообщение", func() {
        panic("ожидаемое сообщение")
    })
}
\`\`\`
```
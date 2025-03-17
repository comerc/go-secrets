#go #errors #error_handling #comparison #type_assertion #golang

# Разница между errors.Is() и errors.As() в Go

```table-of-contents
```

Функции `errors.Is()` и `errors.As()` появились в Go 1.13 как часть нового подхода к обработке ошибок. Они предоставляют более гибкие способы проверки и извлечения информации из ошибок по сравнению с прямым сравнением (`==`) или приведением типов. Рассмотрим каждую из них подробнее, а затем сравним.

## errors.Is()

Функция `errors.Is()` используется для проверки, присутствует ли конкретная ошибка в цепочке ошибок. Это важно, потому что ошибки в Go могут быть обёрнуты (wrapped) с помощью `%w` в `fmt.Errorf`, создавая иерархию. Простое сравнение `==` не будет работать с обёрнутыми ошибками.

**Сигнатура:**

```go
func Is(err, target error) bool
```

*   `err`: Ошибка, которую мы проверяем.
*   `target`: Целевая ошибка, наличие которой мы ищем в цепочке `err`.

**Принцип работы:**

`errors.Is()` работает следующим образом:

1.  Сначала проверяет, равны ли `err` и `target` с помощью `==`. Если равны, возвращает `true`.
2.  Если `err` реализует интерфейс `Is(error) bool`, то вызывает этот метод `err.Is(target)` и возвращает результат. Этот интерфейс позволяет пользовательским типам ошибок определять собственную логику сравнения.
3.  Если `err` является обёрнутой ошибкой (wrapped error), то есть реализует метод `Unwrap() error`, то `errors.Is()` рекурсивно вызывает себя для развёрнутой (unwrapped) ошибки: `errors.Is(err.Unwrap(), target)`.
4.  Если ни одно из условий не выполнено, возвращает `false`.

**Пример:**

```go
package main

import (
	"errors"
	"fmt"
)

var ErrSentinel = errors.New("sentinel error")

func main() {
	err1 := ErrSentinel
	err2 := fmt.Errorf("wrapped: %w", ErrSentinel)

	// Прямое сравнение работает только для err1
	fmt.Println("err1 == ErrSentinel:", err1 == ErrSentinel) // true
	fmt.Println("err2 == ErrSentinel:", err2 == ErrSentinel) // false

	// errors.Is() работает и для обёрнутой ошибки
	fmt.Println("errors.Is(err1, ErrSentinel):", errors.Is(err1, ErrSentinel)) // true
	fmt.Println("errors.Is(err2, ErrSentinel):", errors.Is(err2, ErrSentinel)) // true
}
```

В этом примере `err2` обёртывает `ErrSentinel`. Прямое сравнение `err2 == ErrSentinel` даёт `false`, в то время как `errors.Is(err2, ErrSentinel)` возвращает `true`, потому что `errors.Is` разворачивает ошибку и находит `ErrSentinel` внутри.

## errors.As()

Функция `errors.As()` используется для извлечения конкретного типа ошибки из цепочки ошибок.  Это аналог приведения типа, но работающий с обёрнутыми ошибками.

**Сигнатура:**

```go
func As(err error, target any) bool
```

*   `err`: Ошибка, которую мы проверяем.
*   `target`: Указатель на переменную, в которую мы хотим извлечь ошибку определённого типа.  Тип `target` должен быть указателем на интерфейс или указателем на конкретный тип ошибки.

**Принцип работы:**

1.  Проверяет, можно ли привести `err` к типу, на который указывает `target`. Если да, то присваивает значение `err` переменной, на которую указывает `target`, и возвращает `true`.
2.  Если `err` реализует интерфейс `As(any) bool`, то вызывает этот метод `err.As(target)` и возвращает результат.  Пользовательские типы ошибок могут реализовать этот интерфейс, чтобы предоставить свою логику извлечения.
3.  Если `err` является обёрнутой ошибкой, то рекурсивно вызывает `errors.As(err.Unwrap(), target)`.
4.  Если ни одно из условий не выполнено, возвращает `false`.

**Важно:** `target` должен быть указателем.

**Пример:**

```go
package main

import (
	"errors"
	"fmt"
	"io"
)

type MyError struct {
	Code int
	Msg  string
}

func (e *MyError) Error() string {
	return fmt.Sprintf("MyError: code=%d, msg=%s", e.Code, e.Msg)
}

func main() {
	err1 := &MyError{Code: 404, Msg: "Not Found"}
	err2 := fmt.Errorf("wrapped: %w", err1)

	// Попытка извлечь *MyError из err1
	var myErr *MyError
	if errors.As(err1, &myErr) {
		fmt.Println("Extracted from err1:", myErr.Code, myErr.Msg) // Extracted from err1: 404 Not Found
	}

	// Попытка извлечь *MyError из err2 (обёрнутая ошибка)
	if errors.As(err2, &myErr) {
		fmt.Println("Extracted from err2:", myErr.Code, myErr.Msg) // Extracted from err2: 404 Not Found
	}
    // Попытка извлечь io.EOF
    err3 := fmt.Errorf("wrapped io.EOF: %w", io.EOF)
    var eof error
    if errors.As(err3, &eof){
        fmt.Println("Extracted io.EOF")
    }
}
```

В этом примере `errors.As()` успешно извлекает `*MyError` как из `err1`, так и из обёрнутой ошибки `err2`.  Заметьте, что мы передаём `&myErr` (указатель на `myErr`), а не `myErr`.

## Сравнение errors.Is() и errors.As()

| Характеристика    | errors.Is()                                                                                                                                                                                                                                      | errors.As()                                                                                                                                                                                                                                                                                                                   |
| :---------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Цель**          | Проверить, *содержится* ли целевая ошибка в цепочке.                                                                                                                                                                                               | *Извлечь* ошибку определённого типа из цепочки.                                                                                                                                                                                                                                                                             |
| **Возвращаемое значение** | `bool` (true/false) - присутствует ли ошибка.                                                                                                                                                                                                  | `bool` (true/false) - удалось ли извлечь ошибку.                                                                                                                                                                                                                                                                          |
| **Аргументы**     | `err error, target error` - проверяемая ошибка и целевая ошибка.                                                                                                                                                                                    | `err error, target any` - проверяемая ошибка и *указатель* на переменную, куда извлекать ошибку.                                                                                                                                                                                                                             |
| **Использование**   | Когда нужно просто узнать, есть ли ошибка определённого типа. Например, проверить, является ли ошибка `ErrNotFound`.                                                                                                                                   | Когда нужно получить доступ к полям или методам конкретного типа ошибки. Например, если ошибка имеет поле `Code` (как `MyError` в примере выше), и нужно его прочитать.                                                                                                                                                   |
| **Рекурсия**       | Рекурсивно разворачивает обёрнутые ошибки.                                                                                                                                                                                                        | Рекурсивно разворачивает обёрнутые ошибки.                                                                                                                                                                                                                                                                      |
| **Пользовательские типы** | Может использовать метод `Is(error) bool` пользовательского типа ошибки, если он реализован.                                                                                                                                                 | Может использовать метод `As(any) bool` пользовательского типа ошибки, если он реализован.                                                                                                                                                                                                                                     |
| **Пример** |`if errors.Is(err, os.ErrNotExist)`  Проверка на ошибку "файл не найден" | `var pathError *os.PathError; if errors.As(err, &pathError)` Извлечение ошибки `os.PathError` и затем, например, доступ к `pathError.Path`. |

**Ключевые отличия и когда что использовать:**

1.  **Цель:** `errors.Is()` проверяет наличие, `errors.As()` извлекает.
2.  **Тип `target`:** У `errors.Is()` `target` - это значение ошибки (или интерфейс `error`), у `errors.As()` `target` - это *указатель* на переменную, в которую нужно извлечь ошибку.
3.  **Когда использовать `errors.Is()`:**
    *   Когда вам не нужны детали ошибки, а только факт её наличия определённого типа.
    *   Когда вы работаете со стандартными ошибками, такими как `io.EOF`, `os.ErrNotExist`, и т.д., или со своими sentinel errors.
    *   Когда вам нужно проверить, что ошибка относится к определённой категории (например, ошибка ввода-вывода, ошибка сети и т.д.), и вы определили для этих категорий свои sentinel errors.

4.  **Когда использовать `errors.As()`:**
    *   Когда вам нужно получить доступ к специфическим полям или методам ошибки, которые определены в её типе.
    *   Когда вы работаете со своими собственными типами ошибок, которые содержат дополнительную информацию (например, код ошибки, сообщение, контекст и т.д.).
    *   Когда вам нужно выполнить разное поведение в зависимости от конкретного типа ошибки, а не просто проверить её наличие.

**Пример, объединяющий `errors.Is()` и `errors.As()`:**

```go
package main

import (
	"errors"
	"fmt"
	"os"
)

type MyCustomError struct {
	FilePath string
	Op       string
}

func (e *MyCustomError) Error() string {
	return fmt.Sprintf("operation %s on file %s failed", e.Op, e.FilePath)
}


var ErrCustom = errors.New("custom error")

func OpenAndProcess(filepath string) error {
    //...
    _, err := os.Open(filepath) //Предпологаем что здесь может быть ошибка
    if err != nil {
        if errors.Is(err, os.ErrNotExist)
        {
            return fmt.Errorf("OpenAndProcess: %w", err) // Оборачиваем, но тип ошибки сохраняется
        }
        return &MyCustomError{filepath, "open"} // Возвращаем свой тип ошибки
    }
    return nil
}

func main() {
	err := OpenAndProcess("nonexistent.txt")
	if err != nil {
        //Сначала проверяем, является ли ошибка нашей кастомной
        var myErr *MyCustomError
        if errors.As(err, &myErr){
            fmt.Printf("Custom error occurred: Op: %s, File: %s\n", myErr.Op, myErr.FilePath)
            return
        }

        //Если не является нашей кастомной, проверяем на стандартную ошибку "файл не существует"
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("File not found error:", err)
            return
		}

		fmt.Println("An unexpected error occurred:", err)
	}
}
```

В этом примере `OpenAndProcess` может возвращать как ошибку `os.ErrNotExist` (обёрнутую), так и `MyCustomError`. В `main` мы сначала используем `errors.As()`, чтобы попытаться извлечь `MyCustomError` и получить доступ к её полям. Если это не удаётся, мы используем `errors.Is()`, чтобы проверить, является ли ошибка `os.ErrNotExist`.

Таким образом, `errors.Is()` и `errors.As()` дополняют друг друга, предоставляя полный набор инструментов для работы с иерархиями ошибок в Go. `errors.Is()` для проверки наличия, `errors.As()` для извлечения и доступа к деталям.

```old
В чём разница errors.Is() и errors.As()

\`\`\`go
package main

import "errors"

type MyError struct {
  message string
}

func (e MyError) Error() string {
  return e.message
}

func main() {
  err := MyError{"My custom error"}
  // сравнение со значением:
  println(errors.Is(err, MyError{"My custom error"})) // true
  // сравнение с типом:
  println(errors.As(err, &MyError{})) // true
}
\`\`\`

```
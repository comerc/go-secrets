#goErrorHandling

#go #errors #wrapping #unwrapping #fmt #errorHandling #golang #programming #example #error

# Обработка ошибок в Go: Wrapping и Unwrapping

```table-of-contents
```

## Введение в обработку ошибок

Обработка ошибок является критически важной частью разработки программного обеспечения. В Go, ошибки представлены значениями типа `error`. Этот интерфейс предоставляет простой, но мощный механизм для сообщения о сбоях в работе функций. В Go 1.13 была представлена концепция "оборачивания" (wrapping) ошибок, которая позволяет добавлять контекст к ошибке, сохраняя при этом исходную ошибку.

## Оборачивание ошибок (Error Wrapping)

Оборачивание ошибок позволяет создавать цепочки ошибок, где каждая последующая ошибка содержит информацию о предыдущей и добавляет свой контекст. Это достигается с помощью директивы `%w` в функции `fmt.Errorf`.

Рассмотрим пример:

```go
package main

import (
	"errors"
	"fmt"
)

func main() {
	err1 := errors.New("foo") // Создание базовой ошибки
	err2 := fmt.Errorf("baz %w", err1) // Оборачивание err1 в err2 с добавлением контекста "baz"
	fmt.Println(err2)                // Вывод: baz foo
	fmt.Println(err1)                // Вывод: foo
	fmt.Println(errors.Unwrap(err2)) // Вывод: foo - извлечение оригинальной ошибки
}
```

В этом примере:

1.  `err1 := errors.New("foo")`: Создается новая ошибка с сообщением "foo". Это базовая ошибка, с которой мы начинаем.
2.  `err2 := fmt.Errorf("baz %w", err1)`: Здесь происходит оборачивание. Функция `fmt.Errorf` с директивой `%w` создает новую ошибку, которая "оборачивает" `err1`.  Сообщение новой ошибки - "baz", а `%w` указывает, что `err1` должна быть обернута. Таким образом, `err2` содержит и свое сообщение ("baz"), и ссылку на исходную ошибку `err1`.
3. `fmt.Println(err2)`: Выводит сообщение составной ошибки "baz foo".
4. `fmt.Println(errors.Unwrap(err2))`: Функция `errors.Unwrap` извлекает обернутую ошибку из `err2`. В данном случае, она возвращает `err1`.

## Разворачивание ошибок (Error Unwrapping)

Функция `errors.Unwrap` используется, когда требуется получить доступ к исходной ошибке, которая была обернута. Она возвращает обернутую ошибку, если она существует, или `nil`, если обернутой ошибки нет.

Рассмотрим более сложный пример с несколькими уровнями обертывания:

```go
package main

import (
	"errors"
	"fmt"
)

func doSomething() error {
	return errors.New("initial error")
}

func wrapError() error {
	err := doSomething()
	if err != nil {
		return fmt.Errorf("wrapped: %w", err)
	}
	return nil
}

func doubleWrapError() error {
	err := wrapError()
	if err != nil {
		return fmt.Errorf("double wrapped: %w", err)
	}
	return nil
}

func main() {
	err := doubleWrapError()
	fmt.Println(err) // double wrapped: wrapped: initial error

	unwrapped1 := errors.Unwrap(err)
	fmt.Println(unwrapped1) // wrapped: initial error

	unwrapped2 := errors.Unwrap(unwrapped1)
	fmt.Println(unwrapped2) // initial error

	unwrapped3 := errors.Unwrap(unwrapped2)
	fmt.Println(unwrapped3) // <nil>
}
```

В этом примере:

1.  `doSomething()` возвращает исходную ошибку.
2.  `wrapError()` оборачивает эту ошибку.
3.  `doubleWrapError()` оборачивает ошибку, возвращенную `wrapError()`.
4.  В `main()`, мы последовательно разворачиваем ошибку с помощью `errors.Unwrap`.

## Преимущества оборачивания ошибок

1.  **Сохранение контекста:** Оборачивание позволяет добавлять контекст к ошибке на каждом уровне, делая отладку проще. Можно увидеть всю цепочку вызовов, приведшую к ошибке.
2.  **Сохранение исходной ошибки:** Исходная ошибка не теряется, и к ней можно получить доступ с помощью `errors.Unwrap`.
3.  **Улучшенная обработка ошибок:** Можно проверять как конкретные типы ошибок, так и наличие определенных ошибок в цепочке с помощью `errors.Is` и `errors.As`.

## `errors.Is` и `errors.As`

Функции `errors.Is` и `errors.As` предоставляют более гибкие способы проверки ошибок в цепочке.

*   `errors.Is(err, target)`: Проверяет, является ли ошибка `err` (или любая ошибка в ее цепочке обертывания) ошибкой `target`.

*   `errors.As(err, target)`: Проверяет, может ли ошибка `err` (или любая ошибка в ее цепочке) быть приведена к типу, на который указывает `target`. Если да, то `target` присваивается значение этой ошибки.

Пример:

```go
package main

import (
	"errors"
	"fmt"
	"io"
)

func main() {
	err1 := io.EOF
	err2 := fmt.Errorf("read error: %w", err1)

	// errors.Is
	if errors.Is(err2, io.EOF) {
		fmt.Println("errors.Is: Found io.EOF") // Этот блок выполнится
	}

	// errors.As
	var eofError *io.EOFError // io.EOF is a type (struct) and also has associated value (singleton)
	if errors.As(err2, &eofError) {
		fmt.Printf("errors.As: Type is %T\n", eofError) // Этот блок не выполнится, т.к. io.EOF - значение (переменная), а не тип
	}
	var myErr *MyError
	err3 := fmt.Errorf("wrap: %w", NewMyError())

	if errors.As(err3, &myErr) {
		fmt.Println("errors.As: Found MyError") // Вывод: errors.As: Found MyError
	}

}

type MyError struct{}

func (e *MyError) Error() string {
	return "my error"
}
func NewMyError() error {
	return &MyError{}
}

```
В этом примере:

* `errors.Is(err2, io.EOF)` возвращает `true`, потому что `err2` оборачивает `io.EOF`.
* `errors.As(err2, &eofError)` не присваивает значение `eofError`, потому что `io.EOF` - это *значение*, а не тип.
* Пример с `MyError` демонстрирует, как использовать `errors.As` для извлечения ошибки определенного пользовательского типа.

## Заключение

Оборачивание ошибок в Go - мощный механизм для создания информативных и удобных для отладки ошибок. Использование `fmt.Errorf` с `%w`, `errors.Unwrap`, `errors.Is` и `errors.As` позволяет строить сложные цепочки ошибок, сохранять контекст и упрощает обработку различных типов ошибок.

```old
\`\`\`go
package main

import (
	"errors"
	"fmt"
)

func main() {
	err1 := errors.New("foo")
	err2 := fmt.Errorf("baz %w", err1)
	fmt.Println(err2)                // baz foo
	fmt.Println(err1)                // foo
	fmt.Println(errors.Unwrap(err2)) // foo
}
\`\`\`

```
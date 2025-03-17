#go_1_24 #reflection #unsafe #pointer #interface #comparator #address #comparison #panic

# Разбор кода с паникой и исправление Comparator

```table-of-contents
```

## Пошаговый разбор кода и причин паники

Предоставленный код демонстрирует сравнение указателей разных типов и паникует из-за некорректной логики в `main` и реализации `Comparator`. Разберем детально, почему возникает паника и как исправить ситуацию.

### 1. Почему `false`?

В коде определены два типа: `T0` и `T1`, оба являются псевдонимами для `int`. Создаются переменные `value0` типа `T0` и `value1` типа `T1`. Затем создаются указатели разных типов, но указывающие на одни и те же области памяти:

-   `ptr0a` (`*T0`) и `ptr1a` (`*T1`) указывают на `value0`.
-   `ptr0b` (`*T0`) и `ptr1b` (`*T1`) указывают на `value1`.

Функция `wrapper` вызывает метод `Compare` интерфейса `Comparator`, передавая ему два интерфейсных значения (`interface{}`).  Текущая реализация `Compare` (`ci` struct) просто сравнивает интерфейсные значения с помощью оператора `==`.

Рассмотрим результаты вызовов `wrapper`:

1.  `v0 := wrapper(f, ptr0a, ptr1a)`:  `ptr0a` и `ptr1a` имеют разные типы (`*T0` и `*T1`), несмотря на то, что указывают на одну и ту же область памяти.  При сравнении интерфейсных значений, содержащих указатели разных типов, оператор `==` сравнивает *типы* и *значения*. Поскольку типы разные, результат сравнения - `false`.

2.  `v1 := wrapper(f, ptr0b, ptr1b)`: Аналогично `v0`, `ptr0b` (`*T0`) и `ptr1b` (`*T1`) имеют разные типы, поэтому результат - `false`.

3.  `v2 := wrapper(f, ptr0a, ptr0b)`: `ptr0a` и `ptr0b` имеют одинаковый тип (`*T0`), но указывают на *разные* переменные (`value0` и `value1`).  Оператор `==` при сравнении указателей проверяет, указывают ли они на одну и ту же область памяти. В данном случае они указывают на разные области, поэтому результат - `false`.

4.  `v3 := wrapper(f, ptr1a, ptr1b)`: Аналогично `v2`, `ptr1a` и `ptr1b` имеют одинаковый тип (`*T1`), но указывают на разные переменные, поэтому результат - `false`.

### 2. Условие паники

Условие `if !(v0 || v1) || (v2 || v3)` вызывает панику. Разберем его:

-   `!(v0 || v1)`:  Поскольку `v0` и `v1` оба `false`, `v0 || v1` равно `false`. Отрицание `!(false)` дает `true`.
-   `(v2 || v3)`:  Поскольку `v2` и `v3` оба `false`, `v2 || v3` равно `false`.
-   `true || false`:  Результат - `true`.  Поскольку условие истинно, вызывается `panic("failed")`.

### 3. Исправление `Comparator`

Цель, как я понимаю, в том чтобы `Compare` возвращал `true` если указатели указывают на одну и ту же область памяти *независимо от их типа*.  Есть два основных подхода: с использованием рефлексии и без нее (с использованием `unsafe`).

#### 3.1. Решение с рефлексией

```go
package main

import (
	"fmt"
	"reflect"
)

// Comparator - интерфейс компаратора, модифицировать нельзя
type Comparator interface {
	Compare(a, b interface{}) bool
}

func wrapper(c Comparator, a, b interface{}) bool {
	verdict := c.Compare(a, b)
	fmt.Println(a, b, verdict)

	return verdict
}

type ci struct{}

// Compare - реализация компаратора, которую надо поправить
func (c *ci) Compare(a, b interface{}) bool {
	if reflect.ValueOf(a).Kind() != reflect.Ptr || reflect.ValueOf(b).Kind() != reflect.Ptr {
		return false // Or panic, depending on desired behavior for non-pointers.
	}
	return reflect.ValueOf(a).Pointer() == reflect.ValueOf(b).Pointer()
}

type T0 int
type T1 int

func main() {
	var value0 T0
	var value1 T1

	ptr0a := (*T0)(&value0)
	ptr1a := (*T1)(ptr0a)

	ptr0b := (*T0)(&value1)
	ptr1b := (*T1)(ptr0b)

	f := &ci{}

	v0 := wrapper(f, ptr0a, ptr1a) // true
	v1 := wrapper(f, ptr0b, ptr1b) // true
	v2 := wrapper(f, ptr0a, ptr0b) // false
	v3 := wrapper(f, ptr1a, ptr1b) // false

	if !(v0 || v1) || (v2 || v3) {
		panic("failed")
	}
}
```

**Плюсы:**

-   Работает с любыми типами указателей.
-   Более читаемый код (по сравнению с `unsafe`).

**Минусы:**

-   Использование рефлексии может быть медленнее, чем прямой доступ к памяти.
-   Рефлексия может усложнить отладку.
-   Нужно обрабатывать случаи, когда на вход подаются не указатели.

**Пояснение:**

1.  `reflect.ValueOf(a).Kind() != reflect.Ptr`: Проверяем, что оба аргумента являются указателями.  Если нет, возвращаем `false` (или паникуем, в зависимости от желаемого поведения).

2.  `reflect.ValueOf(a).Pointer()`: Получаем числовое значение указателя (адрес в памяти) как `uintptr`.

3.  Сравниваем числовые значения указателей.

#### 3.2. Решение без рефлексии (с использованием `unsafe`)

```go
package main

import (
	"fmt"
	"unsafe"
)

// Comparator - интерфейс компаратора, модифицировать нельзя
type Comparator interface {
	Compare(a, b interface{}) bool
}

func wrapper(c Comparator, a, b interface{}) bool {
	verdict := c.Compare(a, b)
	fmt.Println(a, b, verdict)

	return verdict
}

type ci struct{}

// Compare - реализация компаратора, которую надо поправить
func (c *ci) Compare(a, b interface{}) bool {
	return getAddr(a) == getAddr(b)
}

func getAddr(a interface{}) uintptr {
	return (*[2]uintptr)(unsafe.Pointer(&a))[1]
}

type T0 int
type T1 int

func main() {
	var value0 T0
	var value1 T1

	ptr0a := (*T0)(&value0)
	ptr1a := (*T1)(ptr0a)

	ptr0b := (*T0)(&value1)
	ptr1b := (*T1)(ptr0b)

	f := &ci{}

	v0 := wrapper(f, ptr0a, ptr1a) // true
	v1 := wrapper(f, ptr0b, ptr1b) // true
	v2 := wrapper(f, ptr0a, ptr0b) // false
	v3 := wrapper(f, ptr1a, ptr1b) // false

  if !(v0 || v1) || (v2 || v3) {
    panic("failed")
  }
}
```

**Плюсы:**

-   Максимальная производительность (нет накладных расходов на рефлексию).

**Минусы:**

-   Использование `unsafe` потенциально опасно.  Неправильное использование может привести к непредсказуемому поведению и ошибкам памяти.
-   Код менее читаемый и более сложный для понимания.
-   Менее переносимый код (может зависеть от особенностей реализации интерфейсов в Go).

**Пояснение:**

1.  `unsafe.Pointer(&a)`:  Получаем `unsafe.Pointer` на интерфейсное значение `a`.  `unsafe.Pointer` - это специальный тип указателя, который может указывать на данные любого типа.

2.  `(*[2]uintptr)(unsafe.Pointer(&a))`:  Преобразуем `unsafe.Pointer` к указателю на массив из двух `uintptr`.  Это работает, потому что интерфейсное значение в Go внутренне представлено как структура из двух слов: тип и указатель на данные.  Мы "обманываем" компилятор, говоря, что по адресу интерфейсной переменной находится массив из двух `uintptr`.

3.  `[1]`:  Обращаемся ко второму элементу массива (индекс 1).  Второе слово интерфейсного значения - это и есть указатель на данные.

4. `getAddr` возвращает адрес как uintptr

5. Compare сравнивает адреса.

#### 3.3. Решение с форматированием (крайне не рекомендуется)

```go
func (c *ci) Compare(a, b interface{}) bool {
  return fmt.Sprintf("%p", a) == fmt.Sprintf("%p", b)
}
```
Это решение использует `fmt.Sprintf("%p", a)`, чтобы получить строковое представление адреса указателя. Хотя оно может работать в некоторых простых случаях, оно крайне не рекомендуется по следующим причинам:

-   **Ненадежность:** Формат вывода `%p` не гарантируется спецификацией Go и может меняться в разных версиях Go, на разных платформах или даже при разных запусках программы.
-   **Низкая производительность:**  Преобразование в строку и сравнение строк гораздо медленнее, чем сравнение числовых значений указателей.
-   **Скрытая логика:**  Сравнение строк маскирует истинную цель - сравнение адресов.

**Вывод:**  Использовать `fmt.Sprintf("%p", a)` для сравнения указателей - *очень плохая практика*.

## Финальное решение и предотвращение паники

Выбираем решение с рефлексией как более безопасное и читаемое.  Чтобы предотвратить панику, нужно изменить логику в `main`:

```go
package main

import (
	"fmt"
	"reflect"
)

// Comparator - интерфейс компаратора, модифицировать нельзя
type Comparator interface {
	Compare(a, b interface{}) bool
}

func wrapper(c Comparator, a, b interface{}) bool {
	verdict := c.Compare(a, b)
	fmt.Println(a, b, verdict)

	return verdict
}

type ci struct{}

// Compare - реализация компаратора, которую надо поправить
func (c *ci) Compare(a, b interface{}) bool {
	if reflect.ValueOf(a).Kind() != reflect.Ptr || reflect.ValueOf(b).Kind() != reflect.Ptr {
		return false // Or panic, depending on desired behavior for non-pointers.
	}
	return reflect.ValueOf(a).Pointer() == reflect.ValueOf(b).Pointer()
}

type T0 int
type T1 int

func main() {
	var value0 T0
	var value1 T1

	ptr0a := (*T0)(&value0)
	ptr1a := (*T1)(ptr0a)

	ptr0b := (*T0)(&value1)
	ptr1b := (*T1)(ptr0b)

	f := &ci{}

	v0 := wrapper(f, ptr0a, ptr1a) // true
	v1 := wrapper(f, ptr0b, ptr1b) // true
	v2 := wrapper(f, ptr0a, ptr0b) // false
	v3 := wrapper(f, ptr1a, ptr1b) // false

	// Корректное условие: паникуем, если *не* выполняется ожидаемое поведение.
	if !((v0 && v1) && !(v2 || v3)) {
		panic("failed")
	}
}
```

Измененное условие: `!((v0 && v1) && !(v2 || v3))`. Оно проверяет, что:

-   `v0` и `v1` истинны (указатели на одну и ту же область памяти).
-   И `v2` и `v3` ложны (указатели на разные области памяти).

Если *хотя бы одно* из этих условий не выполняется, вызывается паника.  Это более логичное и предсказуемое поведение.

```old
1. Почему false?
2. Что надо поправить в реализации Comparator чтобы код перестал паниковать?

\`\`\`go
package main

import (
  "fmt"
)

// Comparator - интерфейс компаратора, модифицировать нельзя
type Comparator interface {
  Compare(a, b interface{}) bool
}

func wrapper(c Comparator, a, b interface{}) bool {
  verdict := c.Compare(a, b)
  fmt.Println(a, b, verdict)

  return verdict
}

type ci struct{}

// Compare - реализация компаратора, которую надо поправить
func (c *ci) Compare(a, b interface{}) bool {
  return a == b
}

type T0 int
type T1 int

func main() {
  var value0 T0
  var value1 T1

  ptr0a := (*T0)(&value0)
  ptr1a := (*T1)(ptr0a)

  ptr0b := (*T0)(&value1)
  ptr1b := (*T1)(ptr0b)

  f := &ci{}

  v0 := wrapper(f, ptr0a, ptr1a) // true
  v1 := wrapper(f, ptr0b, ptr1b) // true
  v2 := wrapper(f, ptr0a, ptr0b) // false
  v3 := wrapper(f, ptr1a, ptr1b) // false

  if !(v0 || v1) || (v2 || v3) {
    panic("failed")
  }
}
\`\`\`

Давайте разберем каждый из случаев:

v0: Сравнивает ptr0a и ptr1a, которые оба указывают на value0. Однако, ptr0a является указателем на T0, а ptr1a является указателем на T1. В Go, указатели на разные типы считаются разными, даже если они указывают на одно и то же значение. Поэтому v0 возвращает false.

v1: Сравнивает ptr0b и ptr1b, которые оба указывают на value1. Опять же, ptr0b является указателем на T0, а ptr1b является указателем на T1. Поэтому v1 также возвращает false.

v2: Сравнивает ptr0a и ptr0b. Оба являются указателями на T0, но они указывают на разные значения (value0 и value1 соответственно). Поэтому v2 возвращает false.

v3: Сравнивает ptr1a и ptr1b. Оба являются указателями на T1, но они указывают на разные значения (value0 и value1 соответственно). Поэтому v3 возвращает false.

Ваш код вызывает панику, если (v0 v1) возвращает false или (v2 v3) возвращает true. Поскольку все четыре значения возвращают false, ваш код вызывает панику.

- решение с рефлексией:

\`\`\`go
func (c *ci) Compare(a, b interface{}) bool {
  return fmt.Sprintf("%p", a) == fmt.Sprintf("%p", b)
}
\`\`\`

- решение без рефлексии:

\`\`\`go
func (c *ci) Compare(a, b interface{}) bool {
  return getAddr(a) == getAddr(b)
}

func getAddr(a interface{}) uintptr {
  return (*[2]uintptr)(unsafe.Pointer(&a))[1]
}
\`\`\`

```
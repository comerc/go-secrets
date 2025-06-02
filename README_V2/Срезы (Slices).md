#golang #slices #memory #functions #operations

# Слайсы в Go: передача, операции и поведение

```table-of-contents
```

## Передача слайса как аргумента функции

При передаче слайса в функцию происходит следующее: длина и вместимость передаются по значению, а сам массив значений передается по указателю. Это создает интересное поведение: изменения существующих элементов сохраняются в исходном слайсе, но добавление новых элементов через `append()` не влияет на исходный слайс.

```go
package main

import "fmt"

func main() {
    cap := 4 // если установить 3, то результаты будут разными; при 4 - одинаковые
    var a = make([]int, 0, cap)
    a = append(a, 111, 222, 333)

    fmt.Printf("%#v\n", getArray(a))
    fmt.Printf("%#v\n", a)
}

func remove(slice []int, s int) []int {
    return append(slice[:s], slice[s+1:]...)
}

func getArray(a []int) []int {
    a = append(a, 444)
    a = remove(a, 0)
    return a
}
```

Когда мы устанавливаем емкость равной 4, оба слайса будут содержать одинаковые значения, потому что при добавлении элемента в `getArray()` не происходит выделения нового массива. Если же установить емкость 3, то при добавлении элемента в `getArray()` произойдет выделение нового массива, и изменения не повлияют на исходный слайс.

## Операции со слайсами

Общий формат операций со слайсом: `a[начало:конец:шаг]`. Если параметры не указаны, используются значения по умолчанию:
- Начало: 0
- Конец: длина слайса
- Шаг: 1

Важно понимать, как меняется емкость при операциях со слайсами:

1. При отрезании слайса с начала (`a[1:]`) емкость уменьшается на количество отрезанных элементов.
2. При отрезании с конца (`a[:n]`) емкость остается равной исходной.
3. Третий параметр позволяет явно указать емкость (не больше исходной): `a[начало:конец:емкость]`.

Поведение `append()` при превышении емкости:
- Если текущая емкость не позволяет добавить новый элемент, то создается новый массив с емкостью в два раза больше исходной.
- Если за один раз добавляется несколько элементов (больше чем в два раза от исходной емкости), то дальнейшее увеличение емкости происходит с шагом 2 (это упрощенное описание механики увеличения ёмкости слайса, см. текущую реализацию в исходном коде).

## Примеры работы со слайсами

### Пример 1: Создание и расширение слайса

```go
func example1Slice() {
    var slice []int
    fmt.Printf("slice is nil %t\n", slice == nil) // true (!)
    slice2 := []int{}
    fmt.Printf("slice2 is nil %t\n", slice2 == nil) // false
    
    // append() увеличивает емкость среза в два раза:
    slice = append(slice, 1)
    fmt.Printf("slise = %+v len = %d; cap = %d;\n", slice, len(slice), cap(slice))
    // slise = [1] len = 1 cap = 1
    
    slice = append(slice, 2)
    fmt.Printf("slice = %+v len = %d; cap = %d;\n", slice, len(slice), cap(slice))
    // slise = [1, 2] len = 2 cap = 2
    
    slice = append(slice, 3)
    fmt.Printf("slice = %+v len = %d; cap = %d;\n", slice, len(slice), cap(slice))
    // slise = [1, 2, 3] len = 3 cap = 4 (!)
}
```

Этот пример демонстрирует важное различие между `nil` слайсом (`var slice []int`) и пустым слайсом (`slice2 := []int{}`). Также показано, как увеличивается емкость слайса при добавлении элементов: после добавления третьего элемента емкость увеличивается до 4.

### Пример 2: Работа с подслайсами

```go
func example2Slice() {
    sl := []int{1, 2, 3, 4, 5, 6}
    sl1 := sl[:3]
    sl2 := sl[1:3:4]

    fmt.Printf("sl1 = %+v len = %d; cap = %d;\n", sl1, len(sl1), cap(sl1))
    // sl1 = [1, 2, 3] len = 3 cap = 6
    fmt.Printf("sl2 = %+v len = %d; cap = %d;\n", sl2, len(sl2), cap(sl2))
    // sl2 = [2, 3] len = 2 cap = 3

    sl2 = append(sl2, 9)
    sl1 = sl1[:4]

    fmt.Printf("sl = %+v len = %d; cap = %d;\n", sl, len(sl), cap(sl))
    // sl = [1, 2, 3, 9, 5, 6] len = 6 cap = 6
    fmt.Printf("sl1 = %+v len = %d; cap = %d;\n", sl1, len(sl1), cap(sl1))
    // sl1 = [1, 2, 3, 9] len = 4 cap = 6
    fmt.Printf("sl2 = %+v len = %d; cap = %d;\n", sl2, len(sl2), cap(sl2))
    // sl2 = [2, 3, 9] len = 3 cap = 3

    add(sl1, 8)
    fmt.Printf("sl = %+v len = %d; cap = %d;\n", sl, len(sl), cap(sl))
    // sl = [1, 2, 3, 9, 8, 6] len = 6 cap = 6
    fmt.Printf("sl1 = %+v len = %d; cap = %d;\n", sl1, len(sl1), cap(sl1))
    // sl1 = [1, 2, 3, 9] len = 4 cap = 6
    fmt.Printf("sl2 = %+v len = %d; cap = %d;\n", sl2, len(sl2), cap(sl2))
    // sl2 = [2, 3, 9] len = 3 cap = 3

    changeSlice(sl, 5, 20)
    fmt.Printf("sl = %+v len = %d; cap = %d;\n", sl, len(sl), cap(sl))
    // sl = [1, 2, 3, 9, 8, 20] len = 6 cap = 6

    sl = append(sl, 7)
    fmt.Printf("sl = %+v len = %d; cap = %d;\n", sl, len(sl), cap(sl))
    // sl = [1, 2, 3, 9, 8, 20, 7] len = 7 cap = 12

    // sl1 = sl1[:7] - panic, cap = 6
    // fmt.Printf("sl1 = %+v len = %d; cap = %d;\n", sl1, len(sl1), cap(sl1))
    // sl1 = [1, 2, 3, 9] len = 4 cap = 6
}

func add(sl []int, val int) {
    sl = append(sl, val)
}

func changeSlice(sl []int, idx int, val int) {
    if 0 <= idx && idx < len(sl) {
        sl[idx] = val
    }
}
```

Этот пример демонстрирует:
1. Как создаются подслайсы и как ограничивается их емкость (с помощью третьего параметра)
2. Как изменения в подслайсах могут влиять на исходный слайс, если они используют общий базовый массив
3. Почему функция `add()` не изменяет исходный слайс (из-за локального `append()`)
4. Почему функция `changeSlice()` изменяет исходный слайс (прямое изменение элемента)
5. Как происходит увеличение емкости при добавлении элемента, когда текущая емкость исчерпана

### Пример 3: Работа с картами (maps)

```go
func example3Map() {
    var myMap map[int]int
    fmt.Printf("myMap is nil %t len = %d;\n", myMap == nil, len(myMap)) // true, len = 0
    // myMap[5] = 55 // panic
    // fmt.Printf("myMap is nil %t len = %d;\n", myMap == nil, len(myMap))

    myMap = map[int]int{}
    fmt.Printf("myMap is nil %t len = %d;\n", myMap == nil, len(myMap)) // false, len = 0
    changeMap(myMap, 6, 66)
    fmt.Printf("myMap is nil %t len = %d;\n", myMap == nil, len(myMap)) // false, len = 1
}

func changeMap(myMap map[int]int, key int, val int) {
    myMap[key] = val
}
```

Этот пример показывает различие между `nil` картой и пустой картой. Также демонстрируется, что карты всегда передаются по ссылке, поэтому функция `changeMap()` изменяет исходную карту.

## Определение общего базового массива для слайсов

Иногда необходимо определить, используют ли два слайса один и тот же базовый массив. Для этого можно использовать пакет `unsafe`:

```go
package main

import "unsafe"

func slicesShareMemory[T any](inner, outer []T) bool {
    if len(inner) == 0 || len(outer) == 0 {
        return false
    }

    aFirstAddr := unsafe.Pointer(&inner[0])
    bFirstAddr := unsafe.Pointer(&outer[0])
    aLastAddr := unsafe.Add(aFirstAddr, uintptr(cap(inner)-1)*unsafe.Sizeof(inner[0]))
    bLastAddr := unsafe.Add(bFirstAddr, uintptr(cap(outer)-1)*unsafe.Sizeof(outer[0]))

    switch {
    case uintptr(aFirstAddr) >= uintptr(bFirstAddr) && uintptr(aFirstAddr) <= uintptr(bLastAddr),
        uintptr(bFirstAddr) >= uintptr(aFirstAddr) && uintptr(bFirstAddr) <= uintptr(aLastAddr):
        return true
    default:
        return false
    }
}

func main() {
    a := []int{1, 2, 3}
    b := a[1:2]
    c := a[2:3]
    d := []int{1, 2, 3}
    e := append(a, 4)
    println(slicesShareMemory(b, a)) // true
    println(slicesShareMemory(a, b)) // true
    println(slicesShareMemory(b, c)) // true
    println(slicesShareMemory(c, b)) // true
    println(slicesShareMemory(a, c)) // true
    println(slicesShareMemory(c, d)) // false
    println(slicesShareMemory(c, e)) // true или false в зависимости от реализации
}
```

Функция `slicesShareMemory` проверяет, пересекаются ли диапазоны памяти двух слайсов. Она работает, сравнивая адреса первого и последнего элементов каждого слайса. Если адрес первого элемента одного слайса находится в диапазоне адресов другого слайса (или наоборот), то слайсы используют общий базовый массив.

## Важные замечания о слайсах

1. Слайс состоит из трех компонентов: указатель на базовый массив, длина и емкость.
2. При передаче слайса в функцию, указатель на массив передается по значению, но сам массив не копируется.
3. Функции, изменяющие элементы слайса, влияют на исходный слайс.
4. Функции, добавляющие элементы через `append()`, создают новый слайс, который не влияет на исходный.
5. Подслайсы могут влиять на исходный слайс, если они используют общий базовый массив.
6. Емкость слайса увеличивается по определенным правилам, когда текущей емкости недостаточно.
7. Третий параметр при создании подслайса позволяет ограничить его емкость.

Понимание этих особенностей слайсов в Go позволяет эффективно использовать их и избегать неожиданного поведения в программах.


>[!quote] Старая версия
```
	## Передача слайса как аргумента функции
	
	Длина и вместимость передаются по значению, но массив значений передается по указателю. Вследствие этого получается неявное поведение: добавленные элементы не сохранятся в исходный слайс, но изменение существующих останется.
	
	```go
	package main
	
	import "fmt"
	
	func main() {
		cap := 4 // если 3, то ответы разные; если 4 - одинаковые
		var a = make([]int, 0, cap)
		a = append(a, 111, 222, 333)
	
		fmt.Printf("%#v\n", getArray(a))
		fmt.Printf("%#v\n", a)
	}
	
	func remove(slice []int, s int) []int {
		return append(slice[:s], slice[s+1:]...)
	}
	
	func getArray(a []int) []int {
		a = append(a, 444)
		a = remove(a, 0)
		return a
	}
	```
	
	Общий формат среза: a[начало:конец:шаг]. Если начало не указано, то по умолчанию начало считается 0. Если конец не указан, то по умолчанию конец считается длиной массива. Если шаг не указан, то по умолчанию шаг считается равным 1.
	
	Эмпирически установлено. Если отрезать слайс сначала, то capacity уменьшается до новой длины, а если с конца, то остаётся равен исходному размеру слайса. Третий параметр позволяет указать capacity явно (но не больше исходного), и он тоже уменьшается от указанного, если отрезать слайс сначала.
	
	Если append() добавляет новый элемент в слайс, у которого превышена capacity, то capacity увеличивается в два раза от исходного. Но если добавлять за раз несколько элементов (больше чем в два раза от исходного), то дальше capacity увеличивается с шагом два.
	
	Видео: [Что нужно знать о слайсах в Go](https://www.youtube.com/watch?v=1vAIvqDo5LE)
	
	### Больше практики по слайсам
	
	```go
	package main
	
	import (
		"fmt"
	)
	
	// что будет выведено?
	// если где-то будет паника, то в какой сторке и почему?
	
	func example1Slice() {
		var slice []int
		fmt.Printf("slice is nil %t\n", slice == nil) // true (!)
		slice2 := []int{}
		fmt.Printf("slice2 is nil %t\n", slice2 == nil) // false
		// append() увеличивает емкость среза в два раза:
		slice = append(slice, 1)
		fmt.Printf("slise = %+v len = %d; cap = %d;\n", slice, len(slice), cap(slice))
		// slise = [1] len = 1 cap = 1
		slice = append(slice, 2)
		fmt.Printf("slice = %+v len = %d; cap = %d;\n", slice, len(slice), cap(slice))
		// slise = [1, 2] len = 2 cap = 2
		slice = append(slice, 3)
		fmt.Printf("slice = %+v len = %d; cap = %d;\n", slice, len(slice), cap(slice))
		// slise = [1, 2, 3] len = 3 cap = 4 (!)
	}
	
	func example2Slice() {
		sl := []int{1, 2, 3, 4, 5, 6}
		sl1 := sl[:3]
		sl2 := sl[1:3:4]
	
		fmt.Printf("sl1 = %+v len = %d; cap = %d;\n", sl1, len(sl1), cap(sl1))
		// sl1 = [1, 2, 3] len = 3 cap = 6
		fmt.Printf("sl2 = %+v len = %d; cap = %d;\n", sl2, len(sl2), cap(sl2))
		// sl2 = [2, 3] len = 2 cap = 3
	
		sl2 = append(sl2, 9)
		sl1 = sl1[:4]
	
		fmt.Printf("sl = %+v len = %d; cap = %d;\n", sl, len(sl), cap(sl))
		// sl = [1, 2, 3, 9, 5, 6] len = 6 cap = 6
		fmt.Printf("sl1 = %+v len = %d; cap = %d;\n", sl1, len(sl1), cap(sl1))
		// sl1 = [1, 2, 3, 9] len = 4 cap = 6
		fmt.Printf("sl2 = %+v len = %d; cap = %d;\n", sl2, len(sl2), cap(sl2))
		// sl2 = [2, 3, 9] len = 3 cap = 3
	
		add(sl1, 8)
		fmt.Printf("sl = %+v len = %d; cap = %d;\n", sl, len(sl), cap(sl))
		// sl = [1, 2, 3, 9, 8, 6] len = 6 cap = 6
		fmt.Printf("sl1 = %+v len = %d; cap = %d;\n", sl1, len(sl1), cap(sl1))
		// sl1 = [1, 2, 3, 9] len = 4 cap = 6
		fmt.Printf("sl2 = %+v len = %d; cap = %d;\n", sl2, len(sl2), cap(sl2))
		// sl2 = [2, 3, 9] len = 3 cap = 3
	
		changeSlice(sl, 5, 20)
		fmt.Printf("sl = %+v len = %d; cap = %d;\n", sl, len(sl), cap(sl))
		// sl = [1, 2, 3, 9, 8, 20] len = 6 cap = 6
	
		sl = append(sl, 7)
		fmt.Printf("sl = %+v len = %d; cap = %d;\n", sl, len(sl), cap(sl))
		// sl = [1, 2, 3, 9, 8, 20, 7] len = 7 cap = 12
	
		// sl1 = sl1[:7] - panic, cap = 6
		// fmt.Printf("sl1 = %+v len = %d; cap = %d;\n", sl1, len(sl1), cap(sl1))
		// sl1 = [1, 2, 3, 9] len = 4 cap = 6
	}
	
	func example3Map() {
		var myMap map[int]int
		fmt.Printf("myMap is nil %t len = %d;\n", myMap == nil, len(myMap)) // true, len = 0
		// myMap[5] = 55 // panic
		// fmt.Printf("myMap is nil %t len = %d;\n", myMap == nil, len(myMap))
	
		myMap = map[int]int{}
		fmt.Printf("myMap is nil %t len = %d;\n", myMap == nil, len(myMap)) // false, len = 0
		changeMap(myMap, 6, 66)
		fmt.Printf("myMap is nil %t len = %d;\n", myMap == nil, len(myMap)) // false, len = 1
	}
	
	func changeSlice(sl []int, idx int, val int) {
		if 0 <= idx && idx < len(sl) {
			sl[idx] = val
		}
	}
	
	func changeMap(myMap map[int]int, key int, val int) {
		myMap[key] = val
	}
	
	func add(sl []int, val int) {
		sl = append(sl, val)
	}
	
	func main() {
		// example1Slice()
		// example2Slice()
		example3Map()
	}
	
	```
	
	
	
	
	### Как узнать, что два слайса используют один базовый массив?
	
	```go
	package main
	
	import "unsafe"
	
	func slicesShareMemory[T any](inner, outer []T) bool {
		if len(inner) == 0 || len(outer) == 0 {
			return false
		}
	
		aFirstAddr := unsafe.Pointer(&inner[0])
		bFirstAddr := unsafe.Pointer(&outer[0])
		aLastAddr := unsafe.Add(aFirstAddr, uintptr(cap(inner)-1)*unsafe.Sizeof(inner[0]))
		bLastAddr := unsafe.Add(bFirstAddr, uintptr(cap(outer)-1)*unsafe.Sizeof(outer[0]))
	
		switch {
		case uintptr(aFirstAddr) >= uintptr(bFirstAddr) && uintptr(aFirstAddr) <= uintptr(bLastAddr),
			uintptr(bFirstAddr) >= uintptr(aFirstAddr) && uintptr(bFirstAddr) <= uintptr(aLastAddr):
			return true
		default:
			return false
		}
	}
	
	func main() {
		a := []int{1, 2, 3}
		b := a[1:2]
		c := a[2:3]
		d := []int{1, 2, 3}
		e := append(a, 4)
		println(slicesShareMemory(b, a))
		println(slicesShareMemory(a, b))
		println(slicesShareMemory(b, c))
		println(slicesShareMemory(c, b))
		println(slicesShareMemory(a, c))
		println(slicesShareMemory(c, d))
		println(slicesShareMemory(c, e))
	}
	```
```


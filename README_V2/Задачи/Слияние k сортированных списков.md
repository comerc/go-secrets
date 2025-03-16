#algorithms #mergedsortedlists #heaps #golang #datastructures

# Слияние k сортированных списков

```table-of-contents
```

## Условие задачи

Даны k отсортированных в порядке неубывания массивов неотрицательных целых чисел, каждое из которых не превосходит 100. Требуется построить результат их слияния: отсортированный в порядке неубывания массив, содержащий все элементы исходных k массивов. Длина каждого массива не превосходит 10 \* k.

## Алгоритм решения

Для решения этой задачи можно использовать структуру данных, которая позволяет эффективно извлекать минимальный элемент - кучу (heap) или приоритетную очередь. Алгоритм работает следующим образом:

1. Создаем указатели на текущие элементы каждого массива (изначально все указатели указывают на первые элементы)
2. Помещаем в кучу элементы, на которые указывают указатели, вместе с информацией о том, из какого массива взят элемент
3. Извлекаем минимальный элемент из кучи, добавляем его в результат
4. Сдвигаем указатель в массиве, из которого был взят минимальный элемент
5. Если в этом массиве остались элементы, добавляем новый элемент в кучу
6. Повторяем шаги 3-5, пока куча не опустеет

Временная сложность алгоритма: O(n log k), где n - общее количество элементов во всех массивах, а k - количество массивов. Пространственная сложность: O(k) для кучи и O(n) для результирующего массива.

[[f|Слияние k сортированных списков]]
## Реализация на Go с использованием встроенной кучи

```go
package main

import (
	"container/heap"
	"fmt"
)

// Item представляет элемент в куче
type Item struct {
	Value     int // Значение элемента
	ArrayIdx  int // Индекс массива
	ElementIdx int // Позиция в массиве
}

// MinHeap реализует интерфейс heap.Interface для кучи минимумов
type MinHeap []*Item

func (h MinHeap) Len() int { return len(h) }
func (h MinHeap) Less(i, j int) bool { return h[i].Value < h[j].Value }
func (h MinHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *MinHeap) Push(x interface{}) {
	*h = append(*h, x.(*Item))
}

func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

func mergeKSortedArrays(arrays [][]int) []int {
	k := len(arrays)
	result := []int{}
	
	// Создаем и инициализируем кучу
	h := &MinHeap{}
	heap.Init(h)
	
	// Добавляем первый элемент каждого массива в кучу
	for i := 0; i < k; i++ {
		if len(arrays[i]) > 0 {
			heap.Push(h, &Item{
				Value:     arrays[i][0],
				ArrayIdx:  i,
				ElementIdx: 0,
			})
		}
	}
	
	// Извлекаем минимальные элементы и добавляем следующие
	for h.Len() > 0 {
		// Извлекаем минимальный элемент
		item := heap.Pop(h).(*Item)
		result = append(result, item.Value)
		
		// Если в массиве есть следующий элемент, добавляем его в кучу
		if item.ElementIdx+1 < len(arrays[item.ArrayIdx]) {
			heap.Push(h, &Item{
				Value:     arrays[item.ArrayIdx][item.ElementIdx+1],
				ArrayIdx:  item.ArrayIdx,
				ElementIdx: item.ElementIdx + 1,
			})
		}
	}
	
	return result
}

func main() {
	// Пример использования
	arrays := [][]int{
		{0, 6, 28},
		{0, 3, 7},
		{5, 6},
	}
	
	result := mergeKSortedArrays(arrays)
	fmt.Println(result) // Ожидаемый вывод: [0 0 3 5 6 6 7 28]
}
```

## Реализация с чтением входных данных

Для полноценного решения задачи необходима обработка входных данных. Вот пример реализации, которая считывает входные данные в формате, описанном в условии задачи:

```go
package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Item представляет элемент в куче
type Item struct {
	Value     int
	ArrayIdx  int
	ElementIdx int
}

// MinHeap реализует интерфейс heap.Interface
type MinHeap []*Item

func (h MinHeap) Len() int { return len(h) }
func (h MinHeap) Less(i, j int) bool { return h[i].Value < h[j].Value }
func (h MinHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *MinHeap) Push(x interface{}) {
	*h = append(*h, x.(*Item))
}

func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

func mergeKSortedArrays(arrays [][]int) []int {
	k := len(arrays)
	result := []int{}
	
	h := &MinHeap{}
	heap.Init(h)
	
	for i := 0; i < k; i++ {
		if len(arrays[i]) > 0 {
			heap.Push(h, &Item{
				Value:     arrays[i][0],
				ArrayIdx:  i,
				ElementIdx: 0,
			})
		}
	}
	
	for h.Len() > 0 {
		item := heap.Pop(h).(*Item)
		result = append(result, item.Value)
		
		if item.ElementIdx+1 < len(arrays[item.ArrayIdx]) {
			heap.Push(h, &Item{
				Value:     arrays[item.ArrayIdx][item.ElementIdx+1],
				ArrayIdx:  item.ArrayIdx,
				ElementIdx: item.ElementIdx + 1,
			})
		}
	}
	
	return result
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	
	// Считываем количество массивов k
	scanner.Scan()
	k, _ := strconv.Atoi(scanner.Text())
	
	arrays := make([][]int, k)
	
	// Считываем массивы
	for i := 0; i < k; i++ {
		scanner.Scan()
		line := scanner.Text()
		numbers := strings.Split(line, " ")
		
		// Первое число - длина массива
		n, _ := strconv.Atoi(numbers[0])
		arrays[i] = make([]int, n)
		
		// Остальные числа - элементы массива
		for j := 0; j < n; j++ {
			arrays[i][j], _ = strconv.Atoi(numbers[j+1])
		}
	}
	
	// Сливаем массивы
	result := mergeKSortedArrays(arrays)
	
	// Выводим результат
	fmt.Println(len(result))
	for _, v := range result {
		fmt.Printf("%d ", v)
	}
	fmt.Println()
}
```

## Альтернативная реализация без использования встроенной кучи

Для лучшего понимания принципов работы кучи, можно реализовать собственную кучу. Вот пример решения без использования встроенного пакета `container/heap`:

```go
package main

import (
	"fmt"
)

// Структура для хранения элемента, массива и индекса
type Item struct {
	Value     int
	ArrayIdx  int
	ElementIdx int
}

// Мин-куча
type MinHeap struct {
	items []*Item
}

// Создание новой кучи
func NewMinHeap() *MinHeap {
	return &MinHeap{items: []*Item{}}
}

// Получение индексов родителя и детей
func (h *MinHeap) parent(i int) int { return (i - 1) / 2 }
func (h *MinHeap) leftChild(i int) int { return 2*i + 1 }
func (h *MinHeap) rightChild(i int) int { return 2*i + 2 }

// Просеивание вверх
func (h *MinHeap) siftUp(i int) {
	for i > 0 && h.items[h.parent(i)].Value > h.items[i].Value {
		h.items[h.parent(i)], h.items[i] = h.items[i], h.items[h.parent(i)]
		i = h.parent(i)
	}
}

// Просеивание вниз
func (h *MinHeap) siftDown(i int) {
	minIndex := i
	l := h.leftChild(i)
	
	if l < len(h.items) && h.items[l].Value < h.items[minIndex].Value {
		minIndex = l
	}
	
	r := h.rightChild(i)
	if r < len(h.items) && h.items[r].Value < h.items[minIndex].Value {
		minIndex = r
	}
	
	if i != minIndex {
		h.items[i], h.items[minIndex] = h.items[minIndex], h.items[i]
		h.siftDown(minIndex)
	}
}

// Добавление элемента в кучу
func (h *MinHeap) Push(item *Item) {
	h.items = append(h.items, item)
	h.siftUp(len(h.items) - 1)
}

// Извлечение минимального элемента
func (h *MinHeap) Pop() *Item {
	if len(h.items) == 0 {
		return nil
	}
	
	root := h.items[0]
	lastIdx := len(h.items) - 1
	h.items[0] = h.items[lastIdx]
	h.items = h.items[:lastIdx]
	
	if len(h.items) > 0 {
		h.siftDown(0)
	}
	
	return root
}

// Проверка, пуста ли куча
func (h *MinHeap) IsEmpty() bool {
	return len(h.items) == 0
}

func mergeKSortedArrays(arrays [][]int) []int {
	result := []int{}
	heap := NewMinHeap()
	
	// Добавляем первые элементы каждого массива в кучу
	for i := 0; i < len(arrays); i++ {
		if len(arrays[i]) > 0 {
			heap.Push(&Item{
				Value:     arrays[i][0],
				ArrayIdx:  i,
				ElementIdx: 0,
			})
		}
	}
	
	// Извлекаем минимальные элементы
	for !heap.IsEmpty() {
		item := heap.Pop()
		result = append(result, item.Value)
		
		// Если в массиве есть следующий элемент, добавляем его
		if item.ElementIdx+1 < len(arrays[item.ArrayIdx]) {
			heap.Push(&Item{
				Value:     arrays[item.ArrayIdx][item.ElementIdx+1],
				ArrayIdx:  item.ArrayIdx,
				ElementIdx: item.ElementIdx + 1,
			})
		}
	}
	
	return result
}

func main() {
	arrays := [][]int{
		{0, 6, 28},
		{0, 3, 7},
		{5, 6},
	}
	
	result := mergeKSortedArrays(arrays)
	fmt.Println(result) // [0 0 3 5 6 6 7 28]
}
```

## Заключение

Задача слияния k сортированных списков является классической задачей, которая демонстрирует эффективность использования структуры данных "куча" или "приоритетная очередь". Куча позволяет нам эффективно извлекать минимальный элемент из множества элементов.

Важно отметить, что в условии задачи первый элемент каждой строки указывает длину соответствующего массива, а не является его элементом. Это необходимо учитывать при разборе входных данных.

Алгоритм слияния k сортированных списков имеет временную сложность O(n log k), где n - общее количество элементов, а k - количество списков, что является оптимальным для этой задачи.


>[!quote] Старая версия
```
	### Задача F. Слияние k сортированных списков
	
> 	Даны k отсортированных в порядке неубывания массивов неотрицательных целых чисел, каждое из которых не превосходит 100. Требуется построить результат их слияния: отсортированный в порядке неубывания массив, содержащий все элементы исходных k массивов. Длина каждого массива не превосходит 10 \* k.
	
	Для каждого массива создадим по указателю; изначально каждый указатель расположен в начале соответствующего массива. Элементы, соответствующие позициям указателей, поместим в любую структуру данных, которая поддерживает извлечение минимума — это может быть мультимножество или, например, куча. Далее будем извлекать из этой структуры минимальный элемент, помещать его в ответ, сдвигать позицию указателя в соответствующем массиве и помещать в структуру данных очередной элемент из этого массива.
	
	В этой задаче многие испытывают сложности с форматом ввода. Обратите внимание на то, что первые элементы строк не описывают элементы массивов, они описывают длины массивов!
	
	#### Answer #1:
	
> 	В данном коде мы создаем структуру Item, которая содержит значение элемента, номер массива, из которого был взят элемент, и индекс элемента в массиве. Затем мы создаем приоритетную очередь PriorityQueue на основе этой структуры, которая будет использоваться для извлечения минимального элемента.
>	
> 	Далее мы создаем указатели на начало каждого массива и добавляем первые элементы каждого массива в приоритетную очередь. Затем мы начинаем извлекать минимальные элементы из очереди, добавляя их в результирующий массив и сдвигая указатель на соответствующий массив. Если указатель не достиг конца массива, мы добавляем следующий элемент из этого массива в приоритетную очередь.
>	
> 	В итоге мы получаем отсортированный в порядке неубывания массив, содержащий все элементы исходных k массивов. В примере выше мы передаем в функцию mergeArrays три отсортированных массива и получаем отсортированный массив [0 0 3 5 6 6 7 28].
	
	#### Answer #2:
	
> 	В этом коде мы используем те же самые структуры данных и алгоритм, что и в предыдущем примере, но более лаконично записываем код. Мы создаем массив pointers для хранения указателей на текущий элемент в каждом массиве, инициализируем его нулями и используем его для проверки достижения конца каждого массива.
>	
> 	Также мы используем оператор append для добавления элементов в результирующий массив, вместо явного указания индекса. Это делает код более читаемым и лаконичным.
>	
> 	В итоге мы получаем тот же отсортированный в порядке неубывания массив, содержащий все элементы исходных k массивов.
	
	#### Answer #3:
	
> 	Этот код считывает количество массивов k, затем считывает каждый массив и его элементы, сливает массивы в один отсортированный массив и выводит его элементы. Для слияния массивов используется куча, которая поддерживает извлечение минимума и добавление элементов. Код также содержит функции для построения кучи и поддержания ее свойств.
```


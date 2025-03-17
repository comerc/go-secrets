#go #arrays #merging #algorithms #heap #priority_queue #data_structures #sorting #software_engineering #go

# Слияние K отсортированных массивов

```table-of-contents
```

## Задача

Требуется разработать алгоритм и реализовать функцию на языке Go, которая эффективно сливает `k` отсортированных массивов целых чисел в один отсортированный массив.

## Решение

Задача слияния `k` отсортированных массивов является классической задачей, которая часто встречается в различных областях, таких как базы данных, поисковые системы и обработка больших данных. Эффективное решение этой задачи имеет важное значение для оптимизации производительности.

Существует несколько подходов к решению этой задачи. Рассмотрим три из них, реализованные в предоставленном коде. Все они используют [[приоритетную очередь (кучу)]] для достижения оптимальной производительности.

### Подход 1: MergeArrays1 (с использованием указателей на int)

1.  **Инициализация указателей.**
    Создается массив `pointers` размером `k`, где `k` – количество входных массивов. Каждый элемент этого массива является указателем на целое число (*int). Эти указатели будут использоваться для отслеживания текущей позиции в каждом из входных массивов. Изначально все указатели инициализируются нулями, указывая на начало каждого массива.

2.  **Инициализация приоритетной очереди (кучи).**
    Создается приоритетная очередь `pq` типа `PriorityQueue`. `PriorityQueue` – это пользовательский тип, представляющий собой мин-кучу. В кучу добавляются элементы типа `Item`. Структура `Item` содержит:
    *   `value`: Значение элемента.
    *   `arrayNum`: Номер массива, из которого взят элемент.
    *   `index`: Индекс элемента в массиве.
    Далее вызывается `heap.Init(&pq)` для инициализации кучи.

3.  **Добавление первых элементов в кучу.**
    В цикле по всем входным массивам, первый элемент каждого массива добавляется в кучу.

4.  **Основной цикл слияния.**
    Пока куча не пуста:
    *   Извлекается элемент с минимальным значением из кучи (`heap.Pop(&pq)`). Этот элемент добавляется в результирующий массив `result`.
    *   Проверяется, есть ли еще элементы в том массиве, из которого был взят текущий минимальный элемент.
    *   Если есть, указатель для этого массива сдвигается на следующую позицию, и новый элемент добавляется в кучу.

### Подход 2: MergeArrays2 (с использованием срезов int)

Этот подход очень похож на первый, но вместо использования указателей на `int` (`*int`) для отслеживания позиций в массивах, он использует обычные целые числа (`int`), хранящиеся в срезе `pointers`. Остальная логика идентична `MergeArrays1`.

### Подход 3: MergeArrays3 (с использованием собственной реализации кучи)

Этот вариант отличается тем, что не использует стандартный пакет `container/heap`, а реализует функции построения и поддержки кучи вручную: `buildHeap`, `heapify`, `heapifyUp`.

1.  **Инициализация указателей:** Создается срез `pointers`, где каждый элемент инициализируется значением `-1`. Это указывает, что изначально ни один из массивов не обработан.

2.  **Создание кучи:** Создается срез `heap` для хранения элементов кучи. В него добавляются первые элементы из каждого непустого входного массива, и соответствующие указатели в `pointers` устанавливаются в 0.

3.  **Построение кучи:** Вызывается функция `buildHeap`, которая преобразует срез `heap` в мин-кучу.

4.  **Основной цикл слияния:**
    *   Пока куча не пуста:
        *   Минимальный элемент извлекается из корня кучи (первый элемент среза `heap`) и добавляется в результирующий массив `result`.
        *   Корень кучи заменяется последним элементом, размер кучи уменьшается, и вызывается `heapify` для восстановления свойств кучи.
        *   Затем ищется массив, из которого был взят извлеченный минимальный элемент. Если в этом массиве есть еще элементы, следующий элемент добавляется в кучу, и вызывается `heapifyUp` для восстановления свойств кучи.

### Сравнение подходов

Все три подхода используют кучу (приоритетную очередь) для эффективного поиска минимального элемента среди текущих элементов всех массивов. Различия заключаются в деталях реализации:

*   **MergeArrays1 и MergeArrays2:** Используют стандартную библиотеку `container/heap`. `MergeArrays1` использует указатели на `int`, что может быть чуть менее эффективно из-за косвенного доступа к значениям. `MergeArrays2` использует обычные `int` в срезе `pointers`, что немного проще и потенциально эффективнее.
*   **MergeArrays3:** Реализует кучу вручную. Это позволяет избежать использования стандартной библиотеки, но требует больше кода и потенциально более подвержено ошибкам. Преимущество может заключаться в большей гибкости и возможности оптимизации под конкретную задачу, но в данном случае это вряд ли даст значительный выигрыш по сравнению со стандартной реализацией.

**Выбор оптимального подхода:**

В большинстве случаев `MergeArrays2` является предпочтительным выбором. Он сочетает простоту и эффективность, используя стандартную библиотеку `container/heap` и избегая излишней сложности с указателями. `MergeArrays3` может быть полезен, если требуется очень тонкая настройка работы кучи, но в стандартных сценариях это излишне. `MergeArrays1` менее предпочтителен из-за использования указателей на `int`.

### Пример использования

```go
package main

import (
	"article"
	"fmt"
)

func main() {
	arrays := [][]int{
		{1, 4, 7, 10},
		{2, 5, 8, 11},
		{3, 6, 9, 12},
	}

	merged1 := article.MergeArrays1(arrays)
	fmt.Println("Merged (Method 1):", merged1) // Вывод: Merged (Method 1): [1 2 3 4 5 6 7 8 9 10 11 12]

	merged2 := article.MergeArrays2(arrays)
	fmt.Println("Merged (Method 2):", merged2) // Вывод: Merged (Method 2): [1 2 3 4 5 6 7 8 9 10 11 12]

	merged3 := article.MergeArrays3(arrays)
	fmt.Println("Merged (Method 3):", merged3) // Вывод: Merged (Method 3): [1 2 3 4 5 6 7 8 9 10 11 12]

	arrays = [][]int{
		{1, 5, 9},
		{2, 6, 10},
		{3, 7, 11},
		{4, 8, 12},
	}

	merged1 = article.MergeArrays1(arrays)
	fmt.Println("Merged (Method 1):", merged1) // Вывод: Merged (Method 1): [1 2 3 4 5 6 7 8 9 10 11 12]

	merged2 = article.MergeArrays2(arrays)
	fmt.Println("Merged (Method 2):", merged2) // Вывод: Merged (Method 2): [1 2 3 4 5 6 7 8 9 10 11 12]

	merged3 = article.MergeArrays3(arrays)
	fmt.Println("Merged (Method 3):", merged3) // Вывод: Merged (Method 3): [1 2 3 4 5 6 7 8 9 10 11 12]
}
```

Этот пример демонстрирует использование всех трех функций `MergeArrays` с двумя различными наборами входных данных. Как видно из вывода, все три метода дают одинаковый, правильно отсортированный результат.

### Сложность

Все три представленных алгоритма имеют одинаковую временную сложность: $O(N \log k)$, где $N$ – общее количество элементов во всех массивах, а $k$ – количество массивов. Это связано с тем, что каждая операция вставки и удаления в куче занимает $O(\log k)$ времени, а всего выполняется $N$ таких операций (по одной для каждого элемента). Пространственная сложность составляет $O(k)$ для хранения кучи и указателей.

```old
\`\`\`go
package article

import (
	"container/heap"
)

type Item struct {
	value    int // значение элемента
	arrayNum int // номер массива, из которого был взят элемент
	index    int // индекс элемента в массиве
}

type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].value < pq[j].value
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*Item)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func MergeArrays1(arrays [][]int) []int {
	k := len(arrays)
	pointers := make([]*int, k)
	for i := 0; i < k; i++ {
		pointers[i] = new(int)
	}
	pq := make(PriorityQueue, 0)
	heap.Init(&pq)
	for i := 0; i < k; i++ {
		heap.Push(&pq, &Item{value: arrays[i][0], arrayNum: i, index: 0})
	}
	result := make([]int, 0)
	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*Item)
		result = append(result, item.value)
		if *pointers[item.arrayNum] < len(arrays[item.arrayNum])-1 {
			*pointers[item.arrayNum]++
			heap.Push(&pq, &Item{value: arrays[item.arrayNum][*pointers[item.arrayNum]], arrayNum: item.arrayNum, index: *pointers[item.arrayNum]})
		}
	}
	return result
}

func MergeArrays2(arrays [][]int) []int {
	k := len(arrays)
	pointers := make([]int, k)
	pq := make(PriorityQueue, 0)
	heap.Init(&pq)
	for i := 0; i < k; i++ {
		heap.Push(&pq, &Item{value: arrays[i][0], arrayNum: i, index: 0})
	}
	result := make([]int, 0)
	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*Item)
		result = append(result, item.value)
		if pointers[item.arrayNum] < len(arrays[item.arrayNum])-1 {
			pointers[item.arrayNum]++
			heap.Push(&pq, &Item{value: arrays[item.arrayNum][pointers[item.arrayNum]], arrayNum: item.arrayNum, index: pointers[item.arrayNum]})
		}
	}
	return result
}

func MergeArrays3(arrays [][]int) []int {
	// Создание указателей на начало каждого массива
	pointers := make([]int, len(arrays))
	for i := range pointers {
		pointers[i] = -1
	}

	// Создание кучи и добавление первых элементов из каждого массива
	heap := make([]int, 0)
	for i, array := range arrays {
		if len(array) > 0 {
			pointers[i] = 0
			heap = append(heap, array[0])
		}
	}
	buildHeap(heap)

	// Слияние массивов
	result := make([]int, 0)
	for len(heap) > 0 {
		// Извлечение минимального элемента из кучи
		min := heap[0]
		result = append(result, min)
		heap[0] = heap[len(heap)-1]
		heap = heap[:len(heap)-1]
		heapify(heap, 0)

		// Добавление следующего элемента из соответствующего массива
		for i, array := range arrays {
			if pointers[i] >= 0 && pointers[i] < len(array) && array[pointers[i]] == min {
				pointers[i]++
				if pointers[i] < len(array) {
					heap = append(heap, array[pointers[i]])
					heapifyUp(heap, len(heap)-1)
				}
			}
		}
	}

	return result
}

func buildHeap(heap []int) {
	for i := len(heap) / 2; i >= 0; i-- {
		heapify(heap, i)
	}
}

func heapify(heap []int, i int) {
	left := 2*i + 1
	right := 2*i + 2
	smallest := i
	if left < len(heap) && heap[left] < heap[smallest] {
		smallest = left
	}
	if right < len(heap) && heap[right] < heap[smallest] {
		smallest = right
	}
	if smallest != i {
		heap[i], heap[smallest] = heap[smallest], heap[i]
		heapify(heap, smallest)
	}
}

func heapifyUp(heap []int, i int) {
	parent := (i - 1) / 2
	if parent >= 0 && heap[i] < heap[parent] {
		heap[i], heap[parent] = heap[parent], heap[i]
		heapifyUp(heap, parent)
	}
}
\`\`\`
```
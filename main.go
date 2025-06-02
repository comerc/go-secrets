package main

import (
	"fmt"
	"go-secrets/ordered_set"
)

func main() {
	fmt.Println("🚀 Демонстрация упорядоченных множеств на Go")
	fmt.Println("============================================")

	// Запускаем примеры использования
	ordered_set.RunExamples()

	fmt.Println("\n📊 Дополнительная демонстрация:")
	performanceComparison()

	fmt.Println("\n✅ Демонстрация завершена!")
}

func performanceComparison() {
	fmt.Println("\n6. Сравнение времени выполнения операций:")

	// Создаем множества
	sliceSet := ordered_set.NewSliceBasedSet[ordered_set.Integer]()
	treeSet := ordered_set.NewTreeSet[ordered_set.Integer]()

	// Тестируем добавление большого количества элементов
	n := 1000
	fmt.Printf("Добавляем %d элементов в каждое множество...\n", n)

	// SliceBasedSet
	for i := 0; i < n; i++ {
		sliceSet.Add(ordered_set.Integer(i))
	}
	fmt.Printf("SliceBasedSet: добавлено %d элементов\n", sliceSet.Size())

	// TreeSet
	for i := 0; i < n; i++ {
		treeSet.Add(ordered_set.Integer(i))
	}
	fmt.Printf("TreeSet: добавлено %d элементов\n", treeSet.Size())

	// Тестируем поиск
	fmt.Println("\nТестируем поиск элементов:")
	testElement := ordered_set.Integer(n / 2)

	fmt.Printf("Поиск элемента %d в SliceBasedSet: %t\n", testElement, sliceSet.Contains(testElement))
	fmt.Printf("Поиск элемента %d в TreeSet: %t\n", testElement, treeSet.Contains(testElement))

	fmt.Println("\nТеоретическая сложность операций:")
	fmt.Println("SliceBasedSet:")
	fmt.Println("  - Добавление: O(n)")
	fmt.Println("  - Поиск: O(log n)")
	fmt.Println("  - Удаление: O(n)")
	fmt.Println("  - Память: O(n)")

	fmt.Println("TreeSet (красно-чёрное дерево):")
	fmt.Println("  - Добавление: O(log n)")
	fmt.Println("  - Поиск: O(log n)")
	fmt.Println("  - Удаление: O(log n)")
	fmt.Println("  - Память: O(n)")
}

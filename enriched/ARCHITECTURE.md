#SOLID #OOP #Go #DesignPatterns #SoftwareArchitecture #ProgrammingPrinciples #SingleResponsibilityPrinciple #OpenClosedPrinciple #LiskovSubstitutionPrinciple #InterfaceSegregationPrinciple #DependencyInversionPrinciple

# Принципы SOLID в Go

```table-of-contents
```

## Введение в SOLID

SOLID - это аббревиатура, представляющая собой пять основных принципов объектно-ориентированного программирования и проектирования, сформулированных Робертом Мартином (Uncle Bob). Эти принципы призваны помочь разработчикам создавать гибкие, легко поддерживаемые и масштабируемые системы. Применение SOLID принципов позволяет избежать многих проблем, связанных с "жестким" кодом, который трудно изменять и повторно использовать.

Рассмотрим каждый из принципов SOLID с примерами на языке Go, а так же проанализируем, как эти принципы соотносятся с другими концепциями, такими как [[Domain-Driven Design (DDD)]], [[луковичная архитектура]] и [[гексагональная архитектура]], [[Event Modeling]].

## 1. Single Responsibility Principle (SRP) - Принцип единственной ответственности

**Формулировка:**  Класс должен иметь только одну причину для изменения. Другими словами, класс должен выполнять только одну задачу или иметь только одну зону ответственности.

**Обоснование:** Если класс отвечает за несколько несвязанных задач, то изменение одной из них может привести к необходимости изменения других, что усложняет поддержку и повышает вероятность ошибок.  Разделение ответственности упрощает понимание, тестирование и модификацию кода.

**Пример (Неправильно):**

```go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
)

type User struct {
	Name  string
	Email string
}

type UserManager struct{}

func (um *UserManager) CreateUser(name, email string) *User {
	user := &User{Name: name, Email: email}
	// ... логика создания пользователя ...
	fmt.Println("User created:", user)
	return user
}

func (um *UserManager) SaveUserToFile(user *User, filename string) {
	data := fmt.Sprintf("Name: %s\nEmail: %s\n", user.Name, user.Email)
	err := ioutil.WriteFile(filename, []byte(data), 0644)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("User saved to file:", filename)
}

func (um *UserManager) SendEmail(user *User, message string) {
    // ... логика отправки email ...
    fmt.Printf("Email sent to %s: %s\n", user.Email, message)
}

func main() {
	um := &UserManager{}
	user := um.CreateUser("John Doe", "john.doe@example.com")
	um.SaveUserToFile(user, "user.txt")
    um.SendEmail(user, "Welcome!")
}
```

В этом примере `UserManager` отвечает за создание пользователя, сохранение его в файл и отправку email.  Это нарушает SRP, так как есть несколько причин для изменения класса: изменение логики создания, изменение формата хранения данных или изменение способа отправки email.

**Пример (Правильно):**

```go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
)

type User struct {
	Name  string
	Email string
}

// UserService отвечает за создание пользователей.
type UserService struct{}

func (us *UserService) CreateUser(name, email string) *User {
	user := &User{Name: name, Email: email}
	// ... логика создания пользователя ...
	fmt.Println("User created:", user)
	return user
}

// UserRepository отвечает за сохранение пользователей.
type UserRepository struct{}

func (ur *UserRepository) SaveToFile(user *User, filename string) {
	data := fmt.Sprintf("Name: %s\nEmail: %s\n", user.Name, user.Email)
	err := ioutil.WriteFile(filename, []byte(data), 0644)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("User saved to file:", filename)
}

// EmailService отвечает за отправку email.
type EmailService struct {}

func (es *EmailService) Send(to, message string) {
    // ... логика отправки email ...
	fmt.Printf("Email sent to %s: %s\n", to, message)
}
func main() {
	us := &UserService{}
    ur := &UserRepository{}
    es := &EmailService{}

	user := us.CreateUser("John Doe", "john.doe@example.com")
	ur.SaveToFile(user, "user.txt")
    es.Send(user.Email, "Welcome!")
}
```
В этом примере мы разделили ответственность на три отдельных класса: `UserService`, `UserRepository` и `EmailService`. Каждый класс имеет единственную ответственность, что делает код более чистым, понятным и легким в поддержке.

**Связь с другими концепциями:** SRP тесно связан с [[высокой связностью (high cohesion)]] и [[низкой связанностью (low coupling)]]. Высокая связность означает, что элементы внутри модуля (например, класса) тесно связаны друг с другом и выполняют общую задачу. Низкая связанность означает, что модули слабо зависят друг от друга. SRP способствует и тому, и другому.

## 2. Open/Closed Principle (OCP) - Принцип открытости/закрытости

**Формулировка:** Программные сущности (классы, модули, функции и т. д.) должны быть открыты для расширения, но закрыты для модификации.

**Обоснование:**  Этот принцип говорит о том, что мы должны иметь возможность добавлять новую функциональность, не изменяя существующий код. Это снижает риск внесения ошибок в уже работающий код и упрощает добавление новых возможностей.

**Пример (Неправильно):**

```go
package main

import "fmt"

type Rectangle struct {
	Width  float64
	Height float64
}

type Circle struct {
	Radius float64
}

type AreaCalculator struct{}

func (ac *AreaCalculator) CalculateArea(shape interface{}) float64 {
	switch s := shape.(type) {
	case Rectangle:
		return s.Width * s.Height
	case Circle:
		return 3.14159 * s.Radius * s.Radius
	default:
		return 0
	}
}

func main() {
	ac := &AreaCalculator{}
	rect := Rectangle{Width: 5, Height: 10}
	circle := Circle{Radius: 3}

	fmt.Println("Rectangle area:", ac.CalculateArea(rect))
	fmt.Println("Circle area:", ac.CalculateArea(circle))

    // Если добавить новый тип фигуры (например, Triangle),
    // придется изменить CalculateArea, добавив новый case.
}
```

В этом примере, если мы захотим добавить новый тип фигуры (например, треугольник), нам придется изменить метод `CalculateArea`, добавив новый `case` в оператор `switch`. Это нарушает OCP.

**Пример (Правильно):**

```go
package main

import "fmt"

type Shape interface {
	Area() float64
}

type Rectangle struct {
	Width  float64
	Height float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return 3.14159 * c.Radius * c.Radius
}

type Triangle struct { // Новая фигура
    Base float64
    Height float64
}

func (t Triangle) Area() float64 { // Реализация интерфейса Shape
    return 0.5 * t.Base * t.Height
}
type AreaCalculator struct{}

func (ac *AreaCalculator) CalculateArea(shapes []Shape) float64 { // Используем интерфейс Shape
    totalArea := 0.0
    for _, shape := range shapes {
        totalArea += shape.Area()
    }
    return totalArea
}

func main() {
    ac := &AreaCalculator{}
	rect := Rectangle{Width: 5, Height: 10}
	circle := Circle{Radius: 3}
    triangle := Triangle{Base: 4, Height: 6} // Новая фигура

	fmt.Println("Rectangle area:", rect.Area())
	fmt.Println("Circle area:", circle.Area())
    fmt.Println("Triangle area:", triangle.Area()) // Вычисляем площадь новой фигуры

    shapes := []Shape{rect, circle, triangle} // Массив различных фигур
    fmt.Println("Total area:", ac.CalculateArea(shapes)) // Общая площадь
}
```

В этом примере мы ввели интерфейс `Shape` с методом `Area()`.  Каждая фигура реализует этот интерфейс.  Теперь, чтобы добавить новую фигуру, нам достаточно создать новый тип, реализующий интерфейс `Shape`, и нам не нужно изменять `AreaCalculator`.  `AreaCalculator` открыт для расширения (мы можем добавлять новые фигуры), но закрыт для модификации (нам не нужно изменять его код).

**Связь с другими концепциями:** OCP часто достигается с помощью [[абстракций]] (интерфейсов в Go) и [[полиморфизма]].  Интерфейсы позволяют нам работать с разными типами объектов единообразно, не зная их конкретной реализации.

## 3. Liskov Substitution Principle (LSP) - Принцип подстановки Барбары Лисков

**Формулировка:**  Подтипы должны быть подставимы вместо своих базовых типов без изменения корректности программы.  Другими словами, если у вас есть функция, которая принимает базовый тип, вы должны иметь возможность передать ей любой подтип, и функция должна продолжать работать правильно.

**Обоснование:** Нарушение LSP приводит к неожиданному поведению программы и затрудняет понимание и поддержку кода. Соблюдение этого принципа обеспечивает надежность и предсказуемость системы.

**Пример (Неправильно):**

```go
package main

import "fmt"

type Rectangle struct {
	Width  float64
	Height float64
}

func (r *Rectangle) SetWidth(width float64) {
	r.Width = width
}

func (r *Rectangle) SetHeight(height float64) {
	r.Height = height
}

func (r *Rectangle) Area() float64 {
    return r.Width * r.Height
}

type Square struct { // Квадрат - это прямоугольник?
	Rectangle
}

func (s *Square) SetWidth(width float64) {
	s.Width = width
	s.Height = width // Нарушение LSP!
}

func (s *Square) SetHeight(height float64) {
	s.Width = height  // Нарушение LSP!
	s.Height = height
}

func ResizeAndPrintArea(r *Rectangle) {
	r.SetWidth(5)
	r.SetHeight(10)
	fmt.Println("Expected area:", 5*10) // Ожидаем 50
	fmt.Println("Actual area:", r.Area())      // Но получаем другое значение, если передадим Square
}

func main() {
	rect := &Rectangle{Width: 2, Height: 3}
	ResizeAndPrintArea(rect) // Работает правильно

	square := &Square{Rectangle: Rectangle{Width: 2, Height: 3}}
	ResizeAndPrintArea(&square.Rectangle) // Нарушение LSP!
}
```

В этом примере `Square` наследуется от `Rectangle`. Однако методы `SetWidth` и `SetHeight` у `Square` переопределены таким образом, что они изменяют и ширину, и высоту, чтобы поддерживать инвариант квадрата (ширина равна высоте). Это нарушает LSP, потому что функция `ResizeAndPrintArea`, ожидающая `Rectangle`, не будет работать корректно с `Square`. После вызова `SetWidth(5)` и `SetHeight(10)` площадь квадрата будет равна 100, а не 50, как ожидалось.

**Пример (Правильно):**

```go
package main

import "fmt"

type Shape interface { // Общий интерфейс
    Area() float64
}

type Rectangle struct {
    Width  float64
    Height float64
}

func (r *Rectangle) Area() float64 {
    return r.Width * r.Height
}

type Square struct {
    Side float64
}

func (s *Square) Area() float64 {
    return s.Side * s.Side
}

func PrintArea(s Shape) { // Работает с любым Shape
    fmt.Println("Area:", s.Area())
}

func main() {
    rect := &Rectangle{Width: 5, Height: 10}
    PrintArea(rect)

    square := &Square{Side: 5}
    PrintArea(square)
}
```
В этом примере и `Rectangle`, и `Square` реализуют интерфейс `Shape`. `Square` больше не наследуется от `Rectangle`, а имеет собственное поле `Side`. Функция `PrintArea` принимает любой объект, реализующий интерфейс `Shape`, и корректно работает с обоими типами. LSP соблюден.

**Связь с другими концепциями:** LSP тесно связан с [[контрактным программированием]] (Design by Contract). Контракты (предусловия, постусловия, инварианты) базового класса должны соблюдаться и подтипами.

## 4. Interface Segregation Principle (ISP) - Принцип разделения интерфейсов

**Формулировка:**  Клиенты не должны зависеть от методов, которые они не используют.  Создавайте узкоспециализированные интерфейсы вместо "толстых" интерфейсов, содержащих множество методов.

**Обоснование:**  "Толстые" интерфейсы приводят к тому, что классы вынуждены реализовывать методы, которые им не нужны. Это нарушает SRP и делает код менее гибким и более сложным в поддержке.

**Пример (Неправильно):**

```go
package main

import "fmt"

type Worker interface {
	Work()
	Eat()
	Sleep()
}

type Robot struct{}

func (r *Robot) Work() {
	fmt.Println("Robot is working...")
}

func (r *Robot) Eat() {
	// Роботу не нужно есть! Но он вынужден реализовывать этот метод.
	fmt.Println("Robot is pretending to eat...") // Или паниковать, или ничего не делать
}

func (r *Robot) Sleep() {
	// Роботу не нужно спать!
	fmt.Println("Robot is pretending to sleep...")
}

type Human struct{}

func (h *Human) Work() {
	fmt.Println("Human is working...")
}

func (h *Human) Eat() {
	fmt.Println("Human is eating...")
}

func (h *Human) Sleep() {
	fmt.Println("Human is sleeping...")
}

func main() {
	robot := &Robot{}
	human := &Human{}

	robot.Work()
	robot.Eat() // Вынужденный вызов
	robot.Sleep() // Вынужденный вызов

	human.Work()
	human.Eat()
	human.Sleep()
}
```

В этом примере интерфейс `Worker` содержит методы `Work`, `Eat` и `Sleep`.  Класс `Robot` вынужден реализовывать методы `Eat` и `Sleep`, хотя они ему не нужны. Это нарушает ISP.

**Пример (Правильно):**

```go
package main

import "fmt"

type Worker interface {
	Work()
}

type Eater interface {
	Eat()
}

type Sleeper interface {
	Sleep()
}
type Human struct{}

func (h *Human) Work() {
    fmt.Println("Human is working...")
}

func (h *Human) Eat() {
    fmt.Println("Human is eating...")
}

func (h *Human) Sleep() {
    fmt.Println("Human is sleeping...")
}

type Robot struct{}

func (r *Robot) Work() {
	fmt.Println("Robot is working...")
}

//  HumanWorker реализует все три интерфейса
type HumanWorker struct {
    Human
}

// RobotWorker реализует только Worker
type RobotWorker struct {
    Robot
}
func DoWork(w Worker) {
    w.Work()
}

func HaveLunch(e Eater) {
    e.Eat()
}

func GetSomeSleep(s Sleeper){
    s.Sleep()
}

func main() {
    hw := HumanWorker{}
    rw := RobotWorker{}

    DoWork(hw) // Human может работать
    DoWork(rw) // Robot может работать
    HaveLunch(hw) // Human может есть
    //HaveLunch(rw) -> Ошибка компиляции, робот не Eater
    GetSomeSleep(hw)
}
```

В этом примере мы разделили интерфейс `Worker` на три более мелких интерфейса: `Worker`, `Eater` и `Sleeper`.  Теперь `Robot` реализует только интерфейс `Worker`, а `Human` может реализовать все три.  Классы зависят только от тех методов, которые им действительно нужны.

**Связь с другими концепциями:** ISP тесно связан с SRP.  Если класс имеет несколько обязанностей, то, скорее всего, он будет зависеть от "толстого" интерфейса. Разделение обязанностей (SRP) часто приводит к разделению интерфейсов (ISP).

## 5. Dependency Inversion Principle (DIP) - Принцип инверсии зависимостей

**Формулировка:**

*   Модули верхних уровней не должны зависеть от модулей нижних уровней.  Оба типа модулей должны зависеть от абстракций.
*   Абстракции не должны зависеть от деталей.  Детали должны зависеть от абстракций.

**Обоснование:**  Этот принцип позволяет создавать слабосвязанные системы, в которых модули верхнего уровня (например, бизнес-логика) не зависят от конкретных реализаций модулей нижнего уровня (например, доступа к базе данных, отправки email).  Это делает систему более гибкой, тестируемой и легкой в поддержке.

**Пример (Неправильно):**

```go
package main

import "fmt"

type MySQLDatabase struct{} // Конкретная реализация базы данных

func (db *MySQLDatabase) GetData() string {
	return "Data from MySQL"
}

type DataService struct { // Высокоуровневый модуль
	DB MySQLDatabase // Зависимость от конкретной реализации
}

func (ds *DataService) GetDataFromDB() string {
	return ds.DB.GetData()
}

func main() {
	db := MySQLDatabase{}
	ds := DataService{DB: db}
	fmt.Println(ds.GetDataFromDB())
     // Если захотим использовать другую базу данных (например, PostgreSQL),
    // придется изменить DataService.
}
```

В этом примере `DataService` (высокоуровневый модуль) напрямую зависит от `MySQLDatabase` (низкоуровневый модуль). Это нарушает DIP. Если мы захотим использовать другую базу данных, нам придется изменить `DataService`.

**Пример (Правильно):**

```go
package main

import "fmt"

type Database interface { // Абстракция (интерфейс)
	GetData() string
}

type MySQLDatabase struct{} // Конкретная реализация

func (db *MySQLDatabase) GetData() string {
	return "Data from MySQL"
}

type PostgreSQLDatabase struct{} // Другая конкретная реализация

func (db *PostgreSQLDatabase) GetData() string {
	return "Data from PostgreSQL"
}

type DataService struct { // Высокоуровневый модуль
	DB Database // Зависимость от абстракции
}

func (ds *DataService) GetDataFromDB() string {
	return ds.DB.GetData()
}

func main() {
	mysqlDB := &MySQLDatabase{}
	ds1 := DataService{DB: mysqlDB} // Используем MySQL
	fmt.Println(ds1.GetDataFromDB())

	postgresDB := &PostgreSQLDatabase{}
	ds2 := DataService{DB: postgresDB} // Используем PostgreSQL
	fmt.Println(ds2.GetDataFromDB())
    // DataService не изменился!
}
```

В этом примере мы ввели интерфейс `Database`.  `DataService` теперь зависит от этого интерфейса, а не от конкретной реализации базы данных.  Мы можем легко переключаться между разными базами данных (`MySQLDatabase`, `PostgreSQLDatabase`), не изменяя `DataService`.  DIP соблюден.

**Связь с другими концепциями:** DIP является основой для многих других паттернов и архитектурных подходов, таких как [[Dependency Injection (Внедрение зависимостей)]], [[Inversion of Control (Инверсия управления)]], [[луковичная архитектура]] и [[гексагональная архитектура]].

## Заключение

Принципы SOLID - это мощный инструмент для создания качественного программного обеспечения. Они помогают создавать гибкие, легко поддерживаемые и масштабируемые системы. Применение SOLID принципов в сочетании с другими концепциями, такими как DDD, луковичная и гексагональная архитектуры, позволяет строить сложные системы, которые легко адаптируются к изменяющимся требованиям. Важно помнить, что SOLID - это не жесткие правила, а скорее рекомендации, которые следует применять с умом, учитывая контекст конкретной задачи.

```old
https://habr.com/ru/companies/productivity_inside/articles/505430/ - SOLID в картинках

https://habr.com/ru/articles/811305/ - SOLID тезисно

https://habr.com/ru/articles/809831/ - DDD

https://habr.com/ru/articles/672328/ - луковичная архитектура

https://habr.com/ru/companies/timeweb/articles/771338/ - гексагональная архитектура

https://habr.com/ru/articles/682424/ - Event Modeling
```
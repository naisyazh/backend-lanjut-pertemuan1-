package main

import "fmt"

type Student struct {
	ID       int
	Name     string
	Grade    float64
	IsActive bool
}

func (s Student) GetInfo() string {
	return fmt.Sprintf("Student[%d]: %s, Grade=%.2f, Active=%v", s.ID, s.Name, s.Grade, s.IsActive)
}

func (s *Student) UpdateGrade(grade float64) {
	s.Grade = grade
}

func (s *Student) Activate() {
	s.IsActive = true
}

func (s *Student) Deactivate() {
	s.IsActive = false
}

func main() {
	fmt.Println("===== DEMO STRUCT STUDENT =====\n")
	
	student1 := Student{
		ID:       1,
		Name:     "Naisya",
		Grade:    85.5,
		IsActive: false,
	}
	
	fmt.Println("1. Info Awal:")
	fmt.Println(student1.GetInfo())
	
	fmt.Println("\n2. Aktivasi Student:")
	student1.Activate()
	fmt.Println(student1.GetInfo())

	fmt.Println("\n3. Update Grade:")
	student1.UpdateGrade(92.0)
	fmt.Println(student1.GetInfo())

	fmt.Println("\n4. Deaktivasi Student:")
	student1.Deactivate()
	fmt.Println(student1.GetInfo())

	fmt.Println("\n5. Student Kedua:")
	student2 := Student{2, "Rizky", 78.0, true}
	fmt.Println(student2.GetInfo())
}

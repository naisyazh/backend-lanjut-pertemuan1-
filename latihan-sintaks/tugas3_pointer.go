package main

import "fmt"

func swap(a *int, b *int) {
	temp := *a
	*a = *b
	*b = temp
}

func updateSlice(s *[]string, newItem string) {
	*s = append(*s, newItem)
}

func ubahNilai(x int) {
	x = 999
	fmt.Println("Di dalam function:", x)
}

func ubahNilaiPointer(x *int) {
	*x = 999
	fmt.Println("Di dalam function:", *x)
}

func main() {
	fmt.Println("===== DEMO POINTER =====\n")
	
	fmt.Println("1. Test Swap:")
	x := 10
	y := 20
	fmt.Printf("Sebelum swap: x=%d, y=%d\n", x, y)
	swap(&x, &y)
	fmt.Printf("Setelah swap: x=%d, y=%d\n", x, y)
	
	fmt.Println("\n2. Test UpdateSlice:")
	hobi := []string{"scroll", "tiktok"}
	fmt.Println("Sebelum update:", hobi)
	updateSlice(&hobi, "gaming")
	fmt.Println("Setelah update:", hobi)
	
	fmt.Println("\n3. Perbandingan Pass by Value vs Pointer:")
	
	angka := 100
	fmt.Println("\nPass by Value:")
	fmt.Println("Sebelum:", angka)
	ubahNilai(angka)
	fmt.Println("Setelah:", angka, "(tidak berubah karena hanya salinan)")
	
	fmt.Println("\nPass by Pointer:")
	fmt.Println("Sebelum:", angka)
	ubahNilaiPointer(&angka)
	fmt.Println("Setelah:", angka, "(berubah karena mengubah alamat asli)")
}

package main

import "fmt"

func main() {

	var nama string = "Naisya"
	umur := 20
	
ipk := 3.75                          
aktif := true                        
hobi := []string{"scroll", "tiktok"} 
	
	fmt.Println("Nama:", nama)
	fmt.Println("Umur:", umur)
	fmt.Printf("IPK: %.2f\n", ipk)
	fmt.Println("Aktif:", aktif)
	fmt.Println("Hobi:", hobi)

	fmt.Println("\n===== BAGIAN MAP =====")
	
	nilaiMahasiswa := make(map[string]int)
	
	nilaiMahasiswa["Isna"] = 85
	nilaiMahasiswa["Yoel"] = 90
	nilaiMahasiswa["Arja"] = 78
	
	nilai, ada := nilaiMahasiswa["Isna"]
	if ada {
		fmt.Printf("Isna dapat nilai: %d\n", nilai)
	} else {
		fmt.Println("Isna tidak ditemukan")
	}
	
	delete(nilaiMahasiswa, "Isna")
	fmt.Println("Isna sudah dihapus")
	
	fmt.Println("Daftar mahasiswa setelah Isna dihapus:")
	for nama, nilai := range nilaiMahasiswa {
		fmt.Printf("- %s: %d\n", nama, nilai)
	}
}




// Alat uji beban portofolio — membuktikan batas 20 proyek tahan serentak.
//
// Cara pakai (server harus sedang berjalan):
//
//	go run ./tools/portfolioload -student 3
//
// Alat ini akan:
// 1. Reset portofolio student_id yang diberikan
// 2. Isi 18 proyek secara berurutan
// 3. Kirim 10 request ADD proyek secara serentak
// 4. Cetak hasil: harusnya tepat 2 berhasil, 8 ditolak 409
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	baseURL   = "http://localhost:3002/api/v1"
	jwtSecret = "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456"
	jwtIssuer = "praktikum-backend"
)

// makeToken membuat JWT token untuk admin (userID=63).
func makeToken() string {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      "63",
		"username": "admin",
		"role":     "admin",
		"iss":      jwtIssuer,
		"iat":      now.Unix(),
		"exp":      now.Add(2 * time.Hour).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(jwtSecret))
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal membuat token: %v\n", err)
		os.Exit(1)
	}
	return signed
}

func doPut(url, token string, body any) string {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "ERR_NETWORK"
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return "200 OK"
	}
	raw, _ := io.ReadAll(resp.Body)
	var out struct{ Code string `json:"code"` }
	json.Unmarshal(raw, &out)
	return fmt.Sprintf("%d %s", resp.StatusCode, out.Code)
}

func doPost(url, token string, body any) string {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "ERR_NETWORK"
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 201 || resp.StatusCode == 200 {
		return fmt.Sprintf("%d OK", resp.StatusCode)
	}
	var out struct{ Code string `json:"code"` }
	json.Unmarshal(raw, &out)
	return fmt.Sprintf("%d %s", resp.StatusCode, out.Code)
}

func doGet(url, token string) string {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "ERR_NETWORK"
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	var out struct {
		Data struct {
			Projects []struct {
				ID string `json:"id"`
			} `json:"projects"`
		} `json:"data"`
	}
	json.Unmarshal(raw, &out)
	return fmt.Sprintf("%d proyek", len(out.Data.Projects))
}

// fire menjalankan n pekerjaan secara serentak — pola yang sama dengan loadtest Modul 9.
func fire(n int, job func(i int) string) map[string]int {
	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make(chan string, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			results <- job(i)
		}(i)
	}

	close(start)
	wg.Wait()
	close(results)

	tally := map[string]int{}
	for r := range results {
		tally[r]++
	}
	return tally
}

func main() {
	studentID := flag.Int("student", 3, "ID mahasiswa (students.id di PostgreSQL)")
	flag.Parse()

	token := makeToken()
	portfolioURL := fmt.Sprintf("%s/students/%d/portfolio", baseURL, *studentID)
	projectsURL := fmt.Sprintf("%s/students/%d/portfolio/projects", baseURL, *studentID)

	fmt.Printf("=== Uji Batas 20 Proyek — student_id=%d ===\n\n", *studentID)

	// Langkah 1: Reset portofolio (PUT dengan projects tidak berubah karena upsert hanya set headline/bio)
	fmt.Println("Langkah 1: Reset portofolio...")
	result := doPut(portfolioURL, token, map[string]string{
		"headline": "Backend Developer | Go & PostgreSQL",
		"bio":      "Mahasiswa Teknik Informatika yang passionate di backend development",
	})
	fmt.Printf("  PUT portfolio: %s\n", result)

	// Langkah 2: Hapus semua proyek yang ada dengan mengambil ID-nya dulu
	// (tidak ada endpoint clear semua, jadi kita tidak bisa reset proyek)
	// Langkah 2: Isi 18 proyek secara berurutan
	fmt.Println("\nLangkah 2: Mengisi 18 proyek secara berurutan...")
	added := 0
	for i := 1; i <= 20; i++ {
		r := doPost(projectsURL, token, map[string]any{
			"title":        fmt.Sprintf("Proyek %02d", i),
			"technologies": []string{"Go", "PostgreSQL"},
			"link":         "https://github.com/naisyazh/backend-lanjut-pertemuan1-",
			"year":         2026,
		})
		if r == "201 OK" || r == "200 OK" {
			added++
		}
		if added == 18 {
			fmt.Printf("  Berhasil mengisi %d proyek\n", added)
			break
		}
		if r != "201 OK" && r != "200 OK" {
			fmt.Printf("  Proyek %d: %s\n", i, r)
			break
		}
	}

	// Cek jumlah proyek sekarang
	current := doGet(portfolioURL, token)
	fmt.Printf("  Jumlah proyek saat ini: %s\n", current)

	// Langkah 3: Kirim 10 request serentak
	fmt.Println("\nLangkah 3: Mengirim 10 request proyek secara serentak...")
	t0 := time.Now()
	tally := fire(10, func(i int) string {
		return doPost(projectsURL, token, map[string]any{
			"title":        fmt.Sprintf("Proyek Serentak %02d", i+1),
			"technologies": []string{"Go", "MongoDB", "Fiber"},
			"link":         "https://github.com/naisyazh/backend-lanjut-pertemuan1-",
			"year":         2026,
		})
	})
	elapsed := time.Since(t0)

	fmt.Println("\nhasil request:")
	ok := 0
	for k, v := range tally {
		fmt.Printf("  %-30s %d\n", k, v)
		if k == "201 OK" || k == "200 OK" {
			ok += v
		}
	}
	fmt.Printf("waktu total : %v\n", elapsed.Round(time.Millisecond))

	// Langkah 4: Cek jumlah akhir
	final := doGet(portfolioURL, token)
	fmt.Printf("\nJumlah proyek akhir: %s\n", final)

	// Kesimpulan
	fmt.Println("\nKesimpulan:")
	if ok == 2 {
		fmt.Println("  ✅ BENAR — tepat 2 berhasil, batas 20 proyek tahan saat diserang serentak")
	} else {
		fmt.Printf("  ❌ TIDAK BENAR — %d berhasil (seharusnya 2)\n", ok)
	}

	// Petunjuk cek di database
	fmt.Println("\nVerifikasi di mongosh/Compass:")
	fmt.Printf("  db.student_portfolios.findOne({student_id: %d}).projects.length\n", *studentID)
	fmt.Println("  Harus menunjukkan 20")

	// Penggunaan token
	_ = strconv.Itoa(*studentID)
}

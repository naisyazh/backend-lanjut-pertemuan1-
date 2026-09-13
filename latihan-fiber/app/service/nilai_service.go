package service

import (
	"fmt"
	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
)

type NilaiService struct {
	nilaiRepo *repository.NilaiRepository
}

func NewNilaiService(nilaiRepo *repository.NilaiRepository) *NilaiService {
	return &NilaiService{nilaiRepo: nilaiRepo}
}

// GetNilaiByNIM - Business logic untuk ambil nilai berdasarkan NIM
func (s *NilaiService) GetNilaiByNIM(nim string) ([]model.NilaiWithStudent, error) {
	if nim == "" {
		return nil, fmt.Errorf("NIM tidak boleh kosong")
	}

	nilaiList, err := s.nilaiRepo.GetNilaiByNIM(nim)
	if err != nil {
		return nil, err
	}

	if len(nilaiList) == 0 {
		return nil, fmt.Errorf("tidak ada nilai ditemukan untuk NIM: %s", nim)
	}

	return nilaiList, nil
}

// CreateNilai - Business logic untuk tambah nilai
func (s *NilaiService) CreateNilai(req model.CreateNilaiRequest) (*model.Nilai, error) {
	// Validasi input
	if req.NamaMataKuliah == "" {
		return nil, fmt.Errorf("nama mata kuliah tidak boleh kosong")
	}
	if req.StudentNIM == "" {
		return nil, fmt.Errorf("NIM mahasiswa tidak boleh kosong")
	}
	if req.Nilai < 0 || req.Nilai > 4 {
		return nil, fmt.Errorf("nilai harus antara 0 dan 4")
	}

	return s.nilaiRepo.CreateNilai(req)
}

// GetAllNilai - Business logic untuk ambil semua nilai
func (s *NilaiService) GetAllNilai() ([]model.NilaiWithStudent, error) {
	return s.nilaiRepo.GetAllNilai()
}

// GetStatistikNilai - Statistik nilai mahasiswa berdasarkan NIM
func (s *NilaiService) GetStatistikNilai(nim string) (map[string]interface{}, error) {
	nilaiList, err := s.nilaiRepo.GetNilaiByNIM(nim)
	if err != nil {
		return nil, err
	}

	if len(nilaiList) == 0 {
		return nil, fmt.Errorf("tidak ada nilai ditemukan untuk NIM: %s", nim)
	}

	// Hitung statistik
	var total float64
	var tertinggi, terendah float64
	jumlahMatkul := len(nilaiList)

	// Set nilai awal
	tertinggi = nilaiList[0].Nilai
	terendah = nilaiList[0].Nilai

	for _, nilai := range nilaiList {
		total += nilai.Nilai
		if nilai.Nilai > tertinggi {
			tertinggi = nilai.Nilai
		}
		if nilai.Nilai < terendah {
			terendah = nilai.Nilai
		}
	}

	rataRata := total / float64(jumlahMatkul)

	return map[string]interface{}{
		"nim":           nim,
		"nama":          nilaiList[0].StudentName,
		"jumlah_matkul": jumlahMatkul,
		"rata_rata":     fmt.Sprintf("%.2f", rataRata),
		"nilai_tertinggi": tertinggi,
		"nilai_terendah":  terendah,
		"total_nilai":   total,
		"detail_nilai":  nilaiList,
	}, nil
}
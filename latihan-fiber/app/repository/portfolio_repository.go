package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"latihan-fiber/app/model"
)

var (
	ErrProjectLimit   = errors.New("jumlah proyek sudah mencapai batas maksimal")
	ErrProjectNotFound = errors.New("proyek tidak ditemukan")
)

// MaxProjectsPerPortfolio membatasi pertumbuhan array projects.
// Pola yang sama dengan MaxRepliesPerReview di Modul 9/10.
const MaxProjectsPerPortfolio = 20

type PortfolioRepository interface {
	EnsureIndexes(ctx context.Context) error

	// FindByStudent mengambil portofolio milik satu mahasiswa.
	// Mengembalikan ErrNotFound bila belum ada.
	FindByStudent(ctx context.Context, studentID int) (model.Portfolio, error)

	// Upsert membuat portofolio baru atau mengganti seluruhnya.
	// Bila belum ada, dibuat dengan projects dan certificates kosong.
	Upsert(ctx context.Context, studentID int, req model.UpsertPortfolioRequest) (model.Portfolio, error)

	// AddProject menambah satu proyek secara ATOMIK.
	// Dibatasi MaxProjectsPerPortfolio — pola "replies.49" dari Modul 10.
	AddProject(ctx context.Context, studentID int, project model.Project) (model.Portfolio, error)

	// DeleteProject menghapus satu proyek berdasarkan project ID.
	DeleteProject(ctx context.Context, studentID int, projectID string) (model.Portfolio, error)

	// TopTechnologies menghasilkan N teknologi paling banyak dipakai
	// di seluruh portofolio menggunakan aggregation pipeline.
	TopTechnologies(ctx context.Context, limit int) ([]model.TechCount, error)
}

// portfolioDocument adalah bentuk dokumen SEBAGAIMANA TERSIMPAN di MongoDB.
// Dipisah dari model.Portfolio agar package model bebas dari tipe driver.
type portfolioDocument struct {
	ID           bson.ObjectID        `bson:"_id,omitempty"`
	StudentID    int                  `bson:"student_id"`
	Headline     string               `bson:"headline"`
	Bio          string               `bson:"bio"`
	Projects     []projectDocument    `bson:"projects"`
	Certificates []certificateDocument `bson:"certificates"`
	CreatedAt    time.Time            `bson:"created_at"`
	UpdatedAt    time.Time            `bson:"updated_at"`
}

type projectDocument struct {
	ID           bson.ObjectID `bson:"_id"`
	Title        string        `bson:"title"`
	Technologies []string      `bson:"technologies"`
	Link         string        `bson:"link"`
	Year         int           `bson:"year"`
	CreatedAt    time.Time     `bson:"created_at"`
}

type certificateDocument struct {
	ID        bson.ObjectID `bson:"_id"`
	Name      string        `bson:"name"`
	Issuer    string        `bson:"issuer"`
	IssuedAt  time.Time     `bson:"issued_at"`
	CreatedAt time.Time     `bson:"created_at"`
}

func (d portfolioDocument) toModel() model.Portfolio {
	projects := make([]model.Project, 0, len(d.Projects))
	for _, p := range d.Projects {
		techs := p.Technologies
		if techs == nil {
			techs = []string{}
		}
		projects = append(projects, model.Project{
			ID:           p.ID.Hex(),
			Title:        p.Title,
			Technologies: techs,
			Link:         p.Link,
			Year:         p.Year,
			CreatedAt:    p.CreatedAt,
		})
	}

	certs := make([]model.Certificate, 0, len(d.Certificates))
	for _, c := range d.Certificates {
		certs = append(certs, model.Certificate{
			ID:        c.ID.Hex(),
			Name:      c.Name,
			Issuer:    c.Issuer,
			IssuedAt:  c.IssuedAt,
			CreatedAt: c.CreatedAt,
		})
	}

	return model.Portfolio{
		ID:           d.ID.Hex(),
		StudentID:    d.StudentID,
		Headline:     d.Headline,
		Bio:          d.Bio,
		Projects:     projects,
		Certificates: certs,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

type portfolioMongoRepository struct{ coll *mongo.Collection }

func NewPortfolioRepository(db *mongo.Database) PortfolioRepository {
	return &portfolioMongoRepository{coll: db.Collection("student_portfolios")}
}

// EnsureIndexes membuat index yang dibutuhkan collection.
// Dipanggil sekali saat aplikasi menyala — aman dipanggil berulang kali.
func (r *portfolioMongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			// Satu mahasiswa hanya boleh memiliki SATU portofolio.
			// Dijaga di level database, bukan hanya di kode.
			Keys:    bson.D{{Key: "student_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("student_unique"),
		},
		{
			// Mendukung pencarian proyek berdasarkan teknologi.
			// Multikey index: setiap elemen array technologies diindeks.
			Keys:    bson.D{{Key: "projects.technologies", Value: 1}},
			Options: options.Index().SetName("project_technologies"),
		},
	})
	if err != nil {
		return fmt.Errorf("membuat index student_portfolios: %w", err)
	}
	return nil
}

func (r *portfolioMongoRepository) FindByStudent(ctx context.Context, studentID int) (model.Portfolio, error) {
	var doc portfolioDocument
	err := r.coll.FindOne(ctx, bson.M{"student_id": studentID}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Portfolio{}, ErrNotFound
		}
		return model.Portfolio{}, fmt.Errorf("mengambil portofolio: %w", err)
	}
	return doc.toModel(), nil
}

// Upsert membuat atau mengganti portofolio mahasiswa.
// Menggunakan FindOneAndUpdate dengan upsert=true agar atomik.
func (r *portfolioMongoRepository) Upsert(ctx context.Context, studentID int, req model.UpsertPortfolioRequest) (model.Portfolio, error) {
	t := now()
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	update := bson.M{
		"$set": bson.M{
			"headline":   req.Headline,
			"bio":        req.Bio,
			"updated_at": t,
		},
		"$setOnInsert": bson.M{
			"student_id":   studentID,
			"projects":     []projectDocument{},
			"certificates": []certificateDocument{},
			"created_at":   t,
		},
	}

	var doc portfolioDocument
	err := r.coll.FindOneAndUpdate(ctx,
		bson.M{"student_id": studentID},
		update,
		opts,
	).Decode(&doc)
	if err != nil {
		return model.Portfolio{}, fmt.Errorf("menyimpan portofolio: %w", err)
	}
	return doc.toModel(), nil
}

// AddProject menambahkan satu proyek secara ATOMIK.
//
// Syarat "projects.19 tidak ada" berarti array belum berisi 20 elemen
// (indeks 0 sampai 19). Pemeriksaan batas dan penambahan terjadi dalam
// satu operasi — pola yang sama dengan AddReply di review_repository.go.
func (r *portfolioMongoRepository) AddProject(ctx context.Context, studentID int, project model.Project) (model.Portfolio, error) {
	techs := project.Technologies
	if techs == nil {
		techs = []string{}
	}

	doc := projectDocument{
		ID:           bson.NewObjectID(),
		Title:        project.Title,
		Technologies: techs,
		Link:         project.Link,
		Year:         project.Year,
		CreatedAt:    now(),
	}

	// Syarat "projects.19 tidak ada" = array belum mencapai 20 elemen
	limitField := "projects." + strconv.Itoa(MaxProjectsPerPortfolio-1)

	var updated portfolioDocument
	err := r.coll.FindOneAndUpdate(ctx,
		bson.M{
			"student_id": studentID,
			limitField:   bson.M{"$exists": false},
		},
		bson.M{
			"$push":      bson.M{"projects": doc},
			"$set":       bson.M{"updated_at": now()},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)
	if err == nil {
		return updated.toModel(), nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return model.Portfolio{}, fmt.Errorf("menambah proyek: %w", err)
	}

	// Tidak cocok — cari tahu penyebabnya
	count, err := r.coll.CountDocuments(ctx, bson.M{"student_id": studentID})
	if err != nil {
		return model.Portfolio{}, fmt.Errorf("memeriksa portofolio: %w", err)
	}
	if count == 0 {
		return model.Portfolio{}, ErrNotFound
	}
	return model.Portfolio{}, ErrProjectLimit
}

// DeleteProject menghapus satu proyek menggunakan $pull.
func (r *portfolioMongoRepository) DeleteProject(ctx context.Context, studentID int, projectID string) (model.Portfolio, error) {
	oid, err := parseID(projectID)
	if err != nil {
		return model.Portfolio{}, err
	}

	var updated portfolioDocument
	err = r.coll.FindOneAndUpdate(ctx,
		bson.M{"student_id": studentID, "projects._id": oid},
		bson.M{
			"$pull": bson.M{"projects": bson.M{"_id": oid}},
			"$set":  bson.M{"updated_at": now()},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Portfolio{}, ErrProjectNotFound
		}
		return model.Portfolio{}, fmt.Errorf("menghapus proyek: %w", err)
	}
	return updated.toModel(), nil
}

// TopTechnologies menghasilkan N teknologi paling banyak dipakai
// menggunakan aggregation pipeline dengan dua kali $unwind:
// 1. $unwind projects — satu portofolio dengan 3 proyek menjadi 3 dokumen
// 2. $unwind technologies — satu proyek dengan 3 teknologi menjadi 3 dokumen
func (r *portfolioMongoRepository) TopTechnologies(ctx context.Context, limit int) ([]model.TechCount, error) {
	pipeline := mongo.Pipeline{
		// Pecah array projects
		{{Key: "$unwind", Value: "$projects"}},
		// Pecah array technologies di dalam setiap proyek
		{{Key: "$unwind", Value: "$projects.technologies"}},
		// Hitung kemunculan tiap teknologi
		{{Key: "$group", Value: bson.M{
			"_id":   "$projects.technologies",
			"total": bson.M{"$sum": 1},
		}}},
		// Urutkan dari terbanyak
		{{Key: "$sort", Value: bson.D{
			{Key: "total", Value: -1},
			{Key: "_id", Value: 1},
		}}},
		// Ambil N teratas
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("menghitung teknologi: %w", err)
	}
	defer cursor.Close(ctx)

	var out []struct {
		Tech  string `bson:"_id"`
		Total int    `bson:"total"`
	}
	if err := cursor.All(ctx, &out); err != nil {
		return nil, fmt.Errorf("membaca statistik teknologi: %w", err)
	}

	result := make([]model.TechCount, 0, len(out))
	for _, o := range out {
		result = append(result, model.TechCount{Technology: o.Tech, Total: o.Total})
	}
	return result, nil
}

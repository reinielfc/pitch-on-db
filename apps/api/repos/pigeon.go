package repos

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"pitch-on-db/db"
	"pitch-on-db/domain"
)

type PigeonRepository interface {
	Create(ctx context.Context, pigeon domain.Pigeon) (domain.Pigeon, error)
	ListAll(ctx context.Context) ([]domain.Pigeon, error)
	Get(ctx context.Context, id int64) (*domain.Pigeon, error)
	Update(ctx context.Context, id int64, pigeonPatch domain.PigeonPatch) (domain.Pigeon, error)
	Delete(ctx context.Context, id int64) error

	Exists(ctx context.Context, id int64) (bool, error)
	GetParents(ctx context.Context, childID int64) (*domain.PigeonParents, error)
	GetChildren(ctx context.Context, parentID int64) ([]domain.Pigeon, error)
	HasChildren(ctx context.Context, parentID int64) (bool, error)
	AssignChild(ctx context.Context, parentID int64, childID int64) error
}

type ParentHandle interface {
	GetChildren(ctx context.Context) ([]domain.Pigeon, error)
	HasChildren(ctx context.Context) (bool, error)
	AssignChild(ctx context.Context, childID int64) error
}

type pigeonRepository struct {
	db      *sql.DB
	queries *db.Queries
}

func NewPigeonRepository(sqlDB *sql.DB) PigeonRepository {
	return &pigeonRepository{db: sqlDB, queries: db.New(sqlDB)}
}

func (r *pigeonRepository) Create(ctx context.Context, pigeon domain.Pigeon) (domain.Pigeon, error) {
	row, err := r.queries.CreatePigeon(ctx, db.CreatePigeonParams{
		Name:       pigeon.Name,
		BandNumber: pigeon.BandNumber,
		BirthDate:  pigeon.BirthDate,
		Sex:        (*string)(pigeon.Sex),
	})
	if err != nil {
		return domain.Pigeon{}, err
	}
	return toDomainPigeon(row), nil
}

func (r *pigeonRepository) ListAll(ctx context.Context) ([]domain.Pigeon, error) {
	rows, err := r.queries.ListPigeons(ctx)
	if err != nil {
		return nil, err
	}

	pigeons := make([]domain.Pigeon, len(rows))
	for i, row := range rows {
		pigeons[i] = toDomainPigeon(row)
	}
	return pigeons, nil
}

func (r *pigeonRepository) Get(ctx context.Context, id int64) (*domain.Pigeon, error) {
	row, err := r.queries.GetPigeon(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	pigeon := toDomainPigeon(row)
	return &pigeon, nil
}

func (r *pigeonRepository) Update(ctx context.Context, id int64, patch domain.PigeonPatch) (domain.Pigeon, error) {
	row, err := r.queries.UpdatePigeon(ctx, db.UpdatePigeonParams{
		ID:   id,
		Name: patch.Name,

		SetBandNumber: patch.BandNumber != nil,
		BandNumber:    patch.BandNumber,

		SetBirthDate: patch.BirthDate != nil,
		BirthDate:    patch.BirthDate,

		SetSex: patch.Sex != nil,
		Sex:    (*string)(patch.Sex),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Pigeon{}, domain.ErrNotFound("pigeon by id %d", id)
	}
	if err != nil {
		return domain.Pigeon{}, fmt.Errorf("update pigeon %d: %w", id, err)
	}

	return toDomainPigeon(row), nil
}

func (r *pigeonRepository) Delete(ctx context.Context, id int64) error {
	return r.queries.DeletePigeon(ctx, id)
}

func (r *pigeonRepository) Exists(ctx context.Context, id int64) (bool, error) {
	return r.queries.CheckPigeonExists(ctx, id)
}

func (r *pigeonRepository) GetParents(ctx context.Context, childID int64) (*domain.PigeonParents, error) {
	parents := &domain.PigeonParents{}

	if fatherRow, err := r.queries.GetPigeonFather(ctx, childID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("get father for pigeon %d: %w", childID, err)
	} else if err == nil {
		father := toDomainPigeon(fatherRow)
		parents.Father = &father
	}

	if motherRow, err := r.queries.GetPigeonMother(ctx, childID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("get mother for pigeon %d: %w", childID, err)
	} else if err == nil {
		mother := toDomainPigeon(motherRow)
		parents.Mother = &mother
	}

	return parents, nil
}

func mapToDomainPigeons(rows []db.Pigeon) []domain.Pigeon {
	pigeons := make([]domain.Pigeon, len(rows))
	for i, row := range rows {
		pigeons[i] = toDomainPigeon(row)
	}
	return pigeons
}

func toDomainPigeon(row db.Pigeon) domain.Pigeon {
	var sex *domain.Sex
	if row.Sex != nil {
		if domainSex, err := domain.ParseSex(*row.Sex); err == nil {
			sex = &domainSex
		}
	}

	return domain.Pigeon{
		ID:         row.ID,
		Name:       row.Name,
		CreatedAt:  row.CreatedAt,
		BandNumber: row.BandNumber,
		BirthDate:  row.BirthDate,
		Sex:        sex,
	}
}

// region Parents

func (r *pigeonRepository) GetChildren(ctx context.Context, parentID int64) ([]domain.Pigeon, error) {
	parentHandle, err := r.ResolveParentHandle(ctx, parentID)
	if err != nil {
		return nil, err
	}
	return parentHandle.GetChildren(ctx)
}

func (r *pigeonRepository) HasChildren(ctx context.Context, parentID int64) (bool, error) {
	parentHandle, err := r.ResolveParentHandle(ctx, parentID)
	if err != nil {
		return false, err
	}
	return parentHandle.HasChildren(ctx)
}

func (r *pigeonRepository) AssignChild(ctx context.Context, parentID int64, childID int64) error {
	parentHandle, err := r.ResolveParentHandle(ctx, parentID)
	if err != nil {
		return err
	}
	return parentHandle.AssignChild(ctx, childID)
}

func (r *pigeonRepository) ResolveParentHandle(ctx context.Context, id int64) (ParentHandle, error) {
	sex, err := r.getPigeonSex(ctx, id)
	if err != nil {
		return nil, err
	}
	if sex == nil {
		return nil, domain.ErrInvalid("parent pigeon %d has no sex set", id)
	}

	switch *sex {
	case domain.SexMale:
		return fatherHandle{repo: r, id: id}, nil
	case domain.SexFemale:
		return motherHandle{repo: r, id: id}, nil
	default:
		return nil, domain.ErrInvalid("parent pigeon %d has invalid sex", id)
	}
}

func (r *pigeonRepository) getPigeonSex(ctx context.Context, id int64) (*domain.Sex, error) {
	sexStr, err := r.queries.GetPigeonSex(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound("pigeon by id %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get sex for pigeon %d: %w", id, err)
	}

	if sexStr == nil {
		return nil, nil
	}

	domainSex, err := domain.ParseSex(*sexStr)
	if err != nil {
		return nil, fmt.Errorf("parse sex for pigeon %d: %w", id, err)
	}
	return &domainSex, nil
}

type fatherHandle struct {
	repo *pigeonRepository
	id   int64
}

func (f fatherHandle) GetChildren(ctx context.Context) ([]domain.Pigeon, error) {
	rows, err := f.repo.queries.GetPigeonChildrenAsFather(ctx, f.id)
	if err != nil {
		return nil, err
	}

	return mapToDomainPigeons(rows), nil
}

func (f fatherHandle) HasChildren(ctx context.Context) (bool, error) {
	return f.repo.queries.CheckPigeonHasChildrenAsFather(ctx, toNullableInt64(&f.id))
}

func (f fatherHandle) AssignChild(ctx context.Context, childID int64) error {
	return f.repo.queries.SetPigeonFather(ctx, db.SetPigeonFatherParams{
		ID:       childID,
		FatherID: toNullableInt64(&f.id),
	})
}

type motherHandle struct {
	repo *pigeonRepository
	id   int64
}

func (m motherHandle) GetChildren(ctx context.Context) ([]domain.Pigeon, error) {
	rows, err := m.repo.queries.GetPigeonChildrenAsMother(ctx, m.id)
	if err != nil {
		return nil, err
	}

	return mapToDomainPigeons(rows), nil
}

func (m motherHandle) HasChildren(ctx context.Context) (bool, error) {
	return m.repo.queries.CheckPigeonHasChildrenAsMother(ctx, toNullableInt64(&m.id))
}

func (m motherHandle) AssignChild(ctx context.Context, childID int64) error {
	return m.repo.queries.SetPigeonMother(ctx, db.SetPigeonMotherParams{
		ID:       childID,
		MotherID: toNullableInt64(&m.id),
	})
}

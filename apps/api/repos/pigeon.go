package repos

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/reinielfc/pitch-on-db/apps/api/db"
	"github.com/reinielfc/pitch-on-db/apps/api/domain"
)

// PigeonRepository defines the data access contract for pigeon records.
type PigeonRepository interface {
	// Create persists a new pigeon record and returns the created pigeon.
	Create(ctx context.Context, pigeon domain.Pigeon) (domain.Pigeon, error)

	// List returns all pigeon records.
	List(ctx context.Context) ([]domain.Pigeon, error)

	// Get returns the pigeon with the given id.
	// Returns [ErrNotFound] if the pigeon does not exist.
	Get(ctx context.Context, id int64) (*domain.Pigeon, error)

	// GetParents returns currently assigned parents for the given child pigeon.
	// Missing parent records are returned as nil fields.
	GetParents(ctx context.Context, childID int64) (*domain.PigeonParents, error)

	// GetChildren returns all children of the given parent pigeon.
	// Errors:
	//   - [ErrNotFound] if the parent pigeon does not exist.
	//   - [ErrPigeonSexUnset] if the parent pigeon has no sex set.
	GetChildren(ctx context.Context, parentID int64) ([]domain.Pigeon, error)

	// Exists reports whether a pigeon with the given id exists.
	Exists(ctx context.Context, id int64) (bool, error)

	// HasChildren reports whether the given parent pigeon has any assigned children.
	// Errors:
	//   - [ErrNotFound] if the parent pigeon does not exist.
	//   - [ErrPigeonSexUnset] if the parent pigeon has no sex set.
	HasChildren(ctx context.Context, parentID int64) (bool, error)

	// Update applies the given patch to the pigeon with the given id and returns the updated pigeon.
	// Returns [ErrNotFound] if the pigeon does not exist.
	Update(ctx context.Context, id int64, pigeonPatch domain.PigeonPatch) (domain.Pigeon, error)

	// AssignChild assigns the child pigeon to the given parent pigeon.
	// Errors:
	//   - [ErrPigeonSexUnset] if the parent pigeon has no sex set.
	//   - [ErrNotFound] if the parent pigeon does not exist.
	AssignChild(ctx context.Context, parentID int64, childID int64) error

	// Delete removes the pigeon with the given id.
	Delete(ctx context.Context, id int64) error
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
		BirthDate:  pigeon.BirthDate,
		Sex:        (*string)(pigeon.Sex),
		Properties: pigeon.Properties,
	})
	if err != nil {
		return domain.Pigeon{}, fmt.Errorf("create pigeon: %w", err)
	}
	return toDomainPigeon(row), nil
}

func (r *pigeonRepository) List(ctx context.Context) ([]domain.Pigeon, error) {
	rows, err := r.queries.ListPigeons(ctx)
	if err != nil {
		return nil, fmt.Errorf("list pigeons: %w", err)
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
		return nil, fmt.Errorf("get pigeon %d: %w", id, ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("get pigeon %d: %w", id, err)
	}

	pigeon := toDomainPigeon(row)
	return &pigeon, nil
}

func (r *pigeonRepository) Update(ctx context.Context, id int64, patch domain.PigeonPatch) (domain.Pigeon, error) {
	row, err := r.queries.UpdatePigeon(ctx, db.UpdatePigeonParams{
		ID:   id,
		Name: patch.Name,

		SetBirthDate: patch.BirthDate != nil,
		BirthDate:    patch.BirthDate,

		SetSex: patch.Sex != nil,
		Sex:    (*string)(patch.Sex),

		SetProperties: patch.Properties != nil,
		Properties:    patch.Properties,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Pigeon{}, fmt.Errorf("update pigeon %d: %w", id, ErrNotFound)
	}
	if err != nil {
		return domain.Pigeon{}, fmt.Errorf("update pigeon %d: %w", id, err)
	}

	return toDomainPigeon(row), nil
}

func (r *pigeonRepository) Delete(ctx context.Context, id int64) error {
	if err := r.queries.DeletePigeon(ctx, id); err != nil {
		return fmt.Errorf("delete pigeon %d: %w", id, err)
	}
	return nil
}

func (r *pigeonRepository) Exists(ctx context.Context, id int64) (bool, error) {
	exists, err := r.queries.CheckPigeonExists(ctx, id)
	if err != nil {
		return false, fmt.Errorf("check if pigeon %d exists: %w", id, err)
	}
	return exists, nil
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
		BirthDate:  row.BirthDate,
		Sex:        sex,
		Properties: row.Properties,
	}
}

// region Parents

func (r *pigeonRepository) GetChildren(ctx context.Context, parentID int64) ([]domain.Pigeon, error) {
	parentHandle, err := r.resolveParentHandle(ctx, parentID)
	if err != nil {
		return nil, err
	}
	return parentHandle.GetChildren(ctx)
}

func (r *pigeonRepository) HasChildren(ctx context.Context, parentID int64) (bool, error) {
	parentHandle, err := r.resolveParentHandle(ctx, parentID)
	if err != nil {
		return false, err
	}
	return parentHandle.HasChildren(ctx)
}

func (r *pigeonRepository) AssignChild(ctx context.Context, parentID int64, childID int64) error {
	parentHandle, err := r.resolveParentHandle(ctx, parentID)
	if err != nil {
		return err
	}
	return parentHandle.AssignChild(ctx, childID)
}

var ErrPigeonSexUnset = errors.New("parent pigeon sex is unset")

func (r *pigeonRepository) resolveParentHandle(ctx context.Context, id int64) (ParentHandle, error) {
	sex, err := r.getPigeonSex(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("resolve parent pigeon %d: %w", id, err)
	}
	if sex == nil {
		return nil, fmt.Errorf("resolve parent pigeon %d: %w", id, ErrPigeonSexUnset)
	}

	switch *sex {
	case domain.SexMale:
		return fatherHandle{repo: r, id: id}, nil
	case domain.SexFemale:
		return motherHandle{repo: r, id: id}, nil
	default:
		return nil, fmt.Errorf("resolve parent pigeon %d: unknown sex '%s'", id, *sex)
	}
}

func (r *pigeonRepository) getPigeonSex(ctx context.Context, id int64) (*domain.Sex, error) {
	sexStr, err := r.queries.GetPigeonSex(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("get sex for pigeon %d: %w", id, ErrNotFound)
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
		return nil, fmt.Errorf("get children for father pigeon %d: %w", f.id, err)
	}

	return mapToDomainPigeons(rows), nil
}

func (f fatherHandle) HasChildren(ctx context.Context) (bool, error) {
	hasChildren, err := f.repo.queries.CheckPigeonHasChildrenAsFather(ctx, toNullableInt64(&f.id))
	if err != nil {
		return false, fmt.Errorf("check children for father pigeon %d: %w", f.id, err)
	}
	return hasChildren, nil
}

func (f fatherHandle) AssignChild(ctx context.Context, childID int64) error {
	err := f.repo.queries.UpdatePigeonFather(ctx, db.UpdatePigeonFatherParams{
		ID:       childID,
		FatherID: toNullableInt64(&f.id),
	})
	if err != nil {
		return fmt.Errorf("assign child %d to father pigeon %d: %w", childID, f.id, err)
	}
	return nil
}

type motherHandle struct {
	repo *pigeonRepository
	id   int64
}

func (m motherHandle) GetChildren(ctx context.Context) ([]domain.Pigeon, error) {
	rows, err := m.repo.queries.GetPigeonChildrenAsMother(ctx, m.id)
	if err != nil {
		return nil, fmt.Errorf("get children for mother pigeon %d: %w", m.id, err)
	}

	return mapToDomainPigeons(rows), nil
}

func (m motherHandle) HasChildren(ctx context.Context) (bool, error) {
	hasChildren, err := m.repo.queries.CheckPigeonHasChildrenAsMother(ctx, toNullableInt64(&m.id))
	if err != nil {
		return false, fmt.Errorf("check children for mother pigeon %d: %w", m.id, err)
	}
	return hasChildren, nil
}

func (m motherHandle) AssignChild(ctx context.Context, childID int64) error {
	err := m.repo.queries.UpdatePigeonMother(ctx, db.UpdatePigeonMotherParams{
		ID:       childID,
		MotherID: toNullableInt64(&m.id),
	})
	if err != nil {
		return fmt.Errorf("assign child %d to mother pigeon %d: %w", childID, m.id, err)
	}
	return nil
}

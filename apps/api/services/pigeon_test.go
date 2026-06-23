package services_test

import (
	"context"
	"testing"

	"github.com/reinielfc/pitch-on-db/apps/api/domain"
	"github.com/reinielfc/pitch-on-db/apps/api/services"
	"github.com/reinielfc/pitch-on-db/apps/api/services/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPigeonService_Get(t *testing.T) {
	t.Run("returns not found when repo returns nil", func(t *testing.T) {
		// Given
		deps := newTestDeps(t)
		deps.Repo.On("Get", mock.Anything, int64(42)).
			Return((*domain.Pigeon)(nil), nil).Once()
		// When
		p, err := deps.Svc.Get(context.Background(), 42)
		// Then
		assert.Nil(t, p)
		assert.ErrorIs(t, err, domain.ErrNotFound)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("returns pigeon when found", func(t *testing.T) {
		// Given
		deps := newTestDeps(t)
		expected := &domain.Pigeon{ID: 1, Name: "Archie"}
		deps.Repo.On("Get", mock.Anything, int64(1)).
			Return(expected, nil).Once()
		// When
		p, err := deps.Svc.Get(context.Background(), 1)
		// Then
		assert.NoError(t, err)
		assert.Equal(t, expected, p)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		// Given
		deps := newTestDeps(t)
		deps.Repo.On("Get", mock.Anything, int64(42)).
			Return((*domain.Pigeon)(nil), assert.AnError).Once()
		// When
		_, err := deps.Svc.Get(context.Background(), 42)
		// Then
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorIs(t, err, domain.ErrInternal)
		deps.Repo.AssertExpectations(t)
	})
}

func TestPigeonService_Update(t *testing.T) {
	t.Run("blocks sex change when pigeon has children", func(t *testing.T) {
		// Given
		deps := newTestDeps(t)
		deps.Repo.On("HasChildren", mock.Anything, int64(11)).
			Return(true, nil).Once()
		sex := domain.SexFemale
		// When
		_, err := deps.Svc.Update(context.Background(), 11, domain.PigeonPatch{Sex: &sex})
		// Then
		assert.ErrorIs(t, err, domain.ErrInvalid)
		deps.Repo.AssertNotCalled(t, "Update", mock.Anything, int64(11), mock.Anything)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("skips child check when sex not patched", func(t *testing.T) {
		// Given
		deps := newTestDeps(t)
		name := "Archie"
		patch := domain.PigeonPatch{Name: &name}
		expected := domain.Pigeon{ID: 3, Name: name}
		deps.Repo.On("Update", mock.Anything, int64(3), patch).
			Return(expected, nil).Once()
		// When
		result, err := deps.Svc.Update(context.Background(), 3, patch)
		// Then
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		deps.Repo.AssertNotCalled(t, "HasChildren", mock.Anything, mock.Anything)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("succeeds when sex patched and no children", func(t *testing.T) {
		// Given
		deps := newTestDeps(t)
		sex := domain.SexMale
		patch := domain.PigeonPatch{Sex: &sex}
		expected := domain.Pigeon{ID: 4, Sex: &sex}
		deps.Repo.On("HasChildren", mock.Anything, int64(4)).
			Return(false, nil).Once()
		deps.Repo.On("Update", mock.Anything, int64(4), patch).
			Return(expected, nil).Once()
		// When
		result, err := deps.Svc.Update(context.Background(), 4, patch)
		// Then
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("propagates error when HasChildren fails", func(t *testing.T) {
		// Given
		deps := newTestDeps(t)
		sex := domain.SexMale
		deps.Repo.On("HasChildren", mock.Anything, int64(5)).
			Return(false, assert.AnError).Once()
		// When
		_, err := deps.Svc.Update(context.Background(), 5, domain.PigeonPatch{Sex: &sex})
		// Then
		assert.ErrorIs(t, err, domain.ErrInternal)
		deps.Repo.AssertNotCalled(t, "Update", mock.Anything, int64(5), mock.Anything)
		deps.Repo.AssertExpectations(t)
	})
}

func TestPigeonService_AssignChild(t *testing.T) {
	t.Run("rejects when parent and child are the same", func(t *testing.T) {
		// Given
		deps := newTestDeps(t)
		// When
		err := deps.Svc.AssignChild(context.Background(), 5, 5)
		// Then
		assert.ErrorIs(t, err, domain.ErrInvalid)
		deps.Repo.AssertNotCalled(t, "AssignChild", mock.Anything, int64(5), int64(5))
	})

	t.Run("returns not found when child does not exist", func(t *testing.T) {
		assertNotFoundWhenPigeonMissing(t, 9, func(svc services.PigeonService) error {
			return svc.AssignChild(context.Background(), 8, 9)
		})
	})

	t.Run("returns not found when parent does not exist", func(t *testing.T) {
		// Given
		deps := newTestDeps(t)
		deps.Repo.On("Exists", mock.Anything, int64(9)).
			Return(true, nil).Once()
		deps.Repo.On("Exists", mock.Anything, int64(8)).
			Return(false, nil).Once()
		// When
		err := deps.Svc.AssignChild(context.Background(), 8, 9)
		// Then
		assert.ErrorIs(t, err, domain.ErrNotFound)
		deps.Repo.AssertNotCalled(t, "AssignChild", mock.Anything, mock.Anything, mock.Anything)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("calls repo when child exists", func(t *testing.T) {
		// Given
		deps := newTestDeps(t)
		deps.Repo.On("Exists", mock.Anything, int64(9)).Return(true, nil).Once()
		deps.Repo.On("Exists", mock.Anything, int64(8)).Return(true, nil).Once()
		deps.Repo.On("AssignChild", mock.Anything, int64(8), int64(9)).Return(nil).Once()
		// When
		err := deps.Svc.AssignChild(context.Background(), 8, 9)
		// Then
		assert.NoError(t, err)
		deps.Repo.AssertExpectations(t)
	})
}

func TestPigeonService_GetParents(t *testing.T) {
	t.Run("returns not found when pigeon does not exist", func(t *testing.T) {
		assertNotFoundWhenPigeonMissing(t, 10, func(svc services.PigeonService) error {
			_, err := svc.GetParents(context.Background(), 10)
			return err
		})
	})

	t.Run("returns parents when pigeon exists", func(t *testing.T) {
		// Given
		deps := newTestDeps(t)
		father := &domain.Pigeon{ID: 1, Name: "Father"}
		expected := &domain.PigeonParents{Father: father}
		deps.Repo.On("Exists", mock.Anything, int64(10)).
			Return(true, nil).Once()
		deps.Repo.On("GetParents", mock.Anything, int64(10)).
			Return(expected, nil).Once()
		// When
		parents, err := deps.Svc.GetParents(context.Background(), 10)
		// Then
		assert.NoError(t, err)
		assert.Equal(t, expected, parents)
		deps.Repo.AssertExpectations(t)
	})
}

func TestPigeonService_GetChildren(t *testing.T) {
	t.Run("returns not found when pigeon does not exist", func(t *testing.T) {
		assertNotFoundWhenPigeonMissing(t, 20, func(svc services.PigeonService) error {
			_, err := svc.GetChildren(context.Background(), 20)
			return err
		})
	})

	t.Run("returns children when pigeon exists", func(t *testing.T) {
		// Given
		deps := newTestDeps(t)
		expected := []domain.Pigeon{{ID: 21, Name: "Chick"}}
		deps.Repo.On("Exists", mock.Anything, int64(20)).
			Return(true, nil).Once()
		deps.Repo.On("GetChildren", mock.Anything, int64(20)).
			Return(expected, nil).Once()
		// When
		children, err := deps.Svc.GetChildren(context.Background(), 20)
		// Then
		assert.NoError(t, err)
		assert.Equal(t, expected, children)
		deps.Repo.AssertExpectations(t)
	})
}

func TestPigeonService_SetTags(t *testing.T) {
	t.Run("returns not found when pigeon does not exist", func(t *testing.T) {
		assertNotFoundWhenPigeonMissing(t, 7, func(svc services.PigeonService) error {
			return svc.SetTags(context.Background(), 7, []string{"racer", "blue"})
		})
	})

	t.Run("calls tag repo when pigeon exists", func(t *testing.T) {
		// Given
		deps := newTestDeps(t)
		deps.Repo.On("Exists", mock.Anything, int64(7)).
			Return(true, nil).Once()
		deps.TagRepo.On("SetPigeonTags", mock.Anything, int64(7), []string{"racer", "blue"}).
			Return(nil).Once()
		// When
		err := deps.Svc.SetTags(context.Background(), 7, []string{"racer", "blue"})
		// Then
		assert.NoError(t, err)
		deps.Repo.AssertExpectations(t)
		deps.TagRepo.AssertExpectations(t)
	})
}

func TestPigeonService_GetTags(t *testing.T) {
	t.Run("returns not found when pigeon does not exist", func(t *testing.T) {
		assertNotFoundWhenPigeonMissing(t, 7, func(svc services.PigeonService) error {
			_, err := svc.GetTags(context.Background(), 7)
			return err
		})
	})

	t.Run("returns tags when pigeon exists", func(t *testing.T) {
		// Given
		deps := newTestDeps(t)
		expected := []string{"racer", "blue"}
		deps.Repo.On("Exists", mock.Anything, int64(7)).
			Return(true, nil).Once()
		deps.TagRepo.On("GetPigeonTags", mock.Anything, int64(7)).
			Return(expected, nil).Once()
		// When
		tags, err := deps.Svc.GetTags(context.Background(), 7)
		// Then
		assert.NoError(t, err)
		assert.Equal(t, expected, tags)
		deps.Repo.AssertExpectations(t)
		deps.TagRepo.AssertExpectations(t)
	})
}

type testDeps struct {
	Repo    *mocks.PigeonRepository
	TagRepo *mocks.TagRepository
	Svc     services.PigeonService
}

func newTestDeps(t *testing.T) testDeps {
	t.Helper()
	repo := mocks.NewPigeonRepository(t)
	tagRepo := mocks.NewTagRepository(t)
	svc := services.NewPigeonService(repo, tagRepo)
	return testDeps{Repo: repo, TagRepo: tagRepo, Svc: svc}
}

func assertNotFoundWhenPigeonMissing(t *testing.T, id int64, call func(svc services.PigeonService) error) {
	t.Helper()
	deps := newTestDeps(t)
	deps.Repo.On("Exists", mock.Anything, id).Return(false, nil).Once()

	err := call(deps.Svc)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	deps.Repo.AssertExpectations(t)
	deps.TagRepo.AssertExpectations(t)
}

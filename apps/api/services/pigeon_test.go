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

func assertNotFoundWhenPigeonMissing(t *testing.T, id int64, call func(svc services.PigeonService) error) {
	t.Helper()
	repo := mocks.NewPigeonRepository(t)
	tagRepo := mocks.NewTagRepository(t)
	repo.On("Exists", mock.Anything, id).Return(false, nil).Once()

	svc := services.NewPigeonService(repo, tagRepo)
	err := call(svc)
	assert.True(t, domain.IsNotFound(err))
	repo.AssertExpectations(t)
}

func TestPigeonService_Get(t *testing.T) {
	t.Run("returns not found when repo returns nil", func(t *testing.T) {
		repo := mocks.NewPigeonRepository(t)
		tagRepo := mocks.NewTagRepository(t)
		repo.On("Get", mock.Anything, int64(42)).Return((*domain.Pigeon)(nil), nil).Once()

		svc := services.NewPigeonService(repo, tagRepo)

		p, err := svc.Get(context.Background(), 42)
		assert.Nil(t, p)
		assert.True(t, domain.IsNotFound(err))
		repo.AssertExpectations(t)
	})

	t.Run("returns pigeon when found", func(t *testing.T) {
		repo := mocks.NewPigeonRepository(t)
		tagRepo := mocks.NewTagRepository(t)
		expected := &domain.Pigeon{ID: 1, Name: "Archie"}
		repo.On("Get", mock.Anything, int64(1)).Return(expected, nil).Once()

		svc := services.NewPigeonService(repo, tagRepo)

		p, err := svc.Get(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, expected, p)
		repo.AssertExpectations(t)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		repo := mocks.NewPigeonRepository(t)
		tagRepo := mocks.NewTagRepository(t)
		repo.On("Get", mock.Anything, int64(1)).Return((*domain.Pigeon)(nil), assert.AnError).Once()

		svc := services.NewPigeonService(repo, tagRepo)

		p, err := svc.Get(context.Background(), 1)
		assert.Nil(t, p)
		assert.ErrorIs(t, err, assert.AnError)
		repo.AssertExpectations(t)
	})
}

func TestPigeonService_Update(t *testing.T) {
	t.Run("blocks sex change when pigeon has children", func(t *testing.T) {
		repo := mocks.NewPigeonRepository(t)
		tagRepo := mocks.NewTagRepository(t)
		repo.On("HasChildren", mock.Anything, int64(11)).Return(true, nil).Once()
		sex := domain.SexFemale

		svc := services.NewPigeonService(repo, tagRepo)

		_, err := svc.Update(context.Background(), 11, domain.PigeonPatch{Sex: &sex})
		assert.True(t, domain.IsInvalid(err))
		repo.AssertNotCalled(t, "Update", mock.Anything, int64(11), mock.Anything)
		repo.AssertExpectations(t)
	})

	t.Run("skips child check when sex not patched", func(t *testing.T) {
		repo := mocks.NewPigeonRepository(t)
		tagRepo := mocks.NewTagRepository(t)
		name := "Archie"
		patch := domain.PigeonPatch{Name: &name}
		expected := domain.Pigeon{ID: 3, Name: name}
		repo.On("Update", mock.Anything, int64(3), patch).Return(expected, nil).Once()

		svc := services.NewPigeonService(repo, tagRepo)

		result, err := svc.Update(context.Background(), 3, patch)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		repo.AssertNotCalled(t, "HasChildren", mock.Anything, mock.Anything)
		repo.AssertExpectations(t)
	})

	t.Run("succeeds when sex patched and no children", func(t *testing.T) {
		repo := mocks.NewPigeonRepository(t)
		tagRepo := mocks.NewTagRepository(t)
		sex := domain.SexMale
		patch := domain.PigeonPatch{Sex: &sex}
		expected := domain.Pigeon{ID: 4, Sex: &sex}
		repo.On("HasChildren", mock.Anything, int64(4)).Return(false, nil).Once()
		repo.On("Update", mock.Anything, int64(4), patch).Return(expected, nil).Once()

		svc := services.NewPigeonService(repo, tagRepo)

		result, err := svc.Update(context.Background(), 4, patch)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		repo.AssertExpectations(t)
	})

	t.Run("propagates error when HasChildren fails", func(t *testing.T) {
		repo := mocks.NewPigeonRepository(t)
		tagRepo := mocks.NewTagRepository(t)
		sex := domain.SexMale
		repo.On("HasChildren", mock.Anything, int64(5)).Return(false, assert.AnError).Once()

		svc := services.NewPigeonService(repo, tagRepo)

		_, err := svc.Update(context.Background(), 5, domain.PigeonPatch{Sex: &sex})
		assert.ErrorIs(t, err, assert.AnError)
		repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
		repo.AssertExpectations(t)
	})
}

func TestPigeonService_AssignChild(t *testing.T) {
	t.Run("rejects when parent and child are the same", func(t *testing.T) {
		repo := mocks.NewPigeonRepository(t)
		tagRepo := mocks.NewTagRepository(t)

		svc := services.NewPigeonService(repo, tagRepo)

		err := svc.AssignChild(context.Background(), 5, 5)
		assert.True(t, domain.IsInvalid(err))
		repo.AssertNotCalled(t, "AssignChild", mock.Anything, int64(5), int64(5))
	})

	t.Run("returns not found when child does not exist", func(t *testing.T) {
		assertNotFoundWhenPigeonMissing(t, 9, func(svc services.PigeonService) error {
			return svc.AssignChild(context.Background(), 8, 9)
		})
	})

	t.Run("returns not found when parent does not exist", func(t *testing.T) {
		repo := mocks.NewPigeonRepository(t)
		tagRepo := mocks.NewTagRepository(t)
		repo.On("Exists", mock.Anything, int64(9)).Return(true, nil).Once()
		repo.On("Exists", mock.Anything, int64(8)).Return(false, nil).Once()

		svc := services.NewPigeonService(repo, tagRepo)

		err := svc.AssignChild(context.Background(), 8, 9)
		assert.True(t, domain.IsNotFound(err))
		repo.AssertNotCalled(t, "AssignChild", mock.Anything, mock.Anything, mock.Anything)
		repo.AssertExpectations(t)
	})

	t.Run("calls repo when child exists", func(t *testing.T) {
		repo := mocks.NewPigeonRepository(t)
		tagRepo := mocks.NewTagRepository(t)
		repo.On("Exists", mock.Anything, int64(9)).Return(true, nil).Once()
		repo.On("Exists", mock.Anything, int64(8)).Return(true, nil).Once()
		repo.On("AssignChild", mock.Anything, int64(8), int64(9)).Return(nil).Once()

		svc := services.NewPigeonService(repo, tagRepo)

		err := svc.AssignChild(context.Background(), 8, 9)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
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
		repo := mocks.NewPigeonRepository(t)
		tagRepo := mocks.NewTagRepository(t)
		father := &domain.Pigeon{ID: 1, Name: "Father"}
		expected := &domain.PigeonParents{Father: father}
		repo.On("Exists", mock.Anything, int64(10)).Return(true, nil).Once()
		repo.On("GetParents", mock.Anything, int64(10)).Return(expected, nil).Once()

		svc := services.NewPigeonService(repo, tagRepo)

		parents, err := svc.GetParents(context.Background(), 10)
		assert.NoError(t, err)
		assert.Equal(t, expected, parents)
		repo.AssertExpectations(t)
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
		repo := mocks.NewPigeonRepository(t)
		tagRepo := mocks.NewTagRepository(t)
		expected := []domain.Pigeon{{ID: 21, Name: "Chick"}}
		repo.On("Exists", mock.Anything, int64(20)).Return(true, nil).Once()
		repo.On("GetChildren", mock.Anything, int64(20)).Return(expected, nil).Once()

		svc := services.NewPigeonService(repo, tagRepo)

		children, err := svc.GetChildren(context.Background(), 20)
		assert.NoError(t, err)
		assert.Equal(t, expected, children)
		repo.AssertExpectations(t)
	})
}

func TestPigeonService_SetTags(t *testing.T) {
	t.Run("returns not found when pigeon does not exist", func(t *testing.T) {
		assertNotFoundWhenPigeonMissing(t, 7, func(svc services.PigeonService) error {
			return svc.SetTags(context.Background(), 7, []string{"racer", "blue"})
		})
	})

	t.Run("calls tag repo when pigeon exists", func(t *testing.T) {
		repo := mocks.NewPigeonRepository(t)
		tagRepo := mocks.NewTagRepository(t)
		repo.On("Exists", mock.Anything, int64(7)).Return(true, nil).Once()
		tagRepo.On("SetPigeonTags", mock.Anything, int64(7), []string{"racer", "blue"}).Return(nil).Once()

		svc := services.NewPigeonService(repo, tagRepo)

		err := svc.SetTags(context.Background(), 7, []string{"racer", "blue"})
		assert.NoError(t, err)
		repo.AssertExpectations(t)
		tagRepo.AssertExpectations(t)
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
		repo := mocks.NewPigeonRepository(t)
		tagRepo := mocks.NewTagRepository(t)
		expected := []string{"racer", "blue"}
		repo.On("Exists", mock.Anything, int64(7)).Return(true, nil).Once()
		tagRepo.On("GetPigeonTags", mock.Anything, int64(7)).Return(expected, nil).Once()

		svc := services.NewPigeonService(repo, tagRepo)

		tags, err := svc.GetTags(context.Background(), 7)
		assert.NoError(t, err)
		assert.Equal(t, expected, tags)
		repo.AssertExpectations(t)
		tagRepo.AssertExpectations(t)
	})
}

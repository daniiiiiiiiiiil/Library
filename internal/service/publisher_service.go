package service

import (
	"context"
	"fmt"
	"library/internal/domain"
	"library/internal/repository"
	"library/pkg/errors"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type PublisherService struct {
	publisherRepo repository.PublisherRepository
	bookRepo      repository.BookRepository
	logger        *zap.Logger
}

func NewPublisherService(
	publisherRepo repository.PublisherRepository,
	bookRepo repository.BookRepository,
	logger *zap.Logger,
) *PublisherService {
	return &PublisherService{
		publisherRepo: publisherRepo,
		bookRepo:      bookRepo,
		logger:        logger,
	}
}

func (s *PublisherService) CreatePublisher(ctx context.Context, conn *pgx.Conn, publisher *domain.Publisher) (*domain.Publisher, error) {
	s.logger.Info("create publisher started", zap.String("name", publisher.Name))

	if err := publisher.Validate(); err != nil {
		s.logger.Warn("publisher validation failed", zap.String("name", publisher.Name), zap.Error(err))
		return nil, errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации" + err.Error(),
		}
	}

	exists, err := s.publisherRepo.ExistsByName(ctx, conn, publisher.Name)
	if err != nil {
		s.logger.Error("failed to check publisher existence by name", zap.String("name", publisher.Name), zap.Error(err))
		return nil, err
	}
	if exists {
		s.logger.Warn("publisher already exists", zap.String("name", publisher.Name))
		return nil, errors.BusinessError{
			Code:    "ErrPublisherAlreadyExists",
			Message: fmt.Sprintf("Издательство с названием '%s' уже существует", publisher.Name),
		}
	}

	createdPublisher, err := s.publisherRepo.Create(ctx, conn, *publisher)
	if err != nil {
		s.logger.Error("failed to create publisher", zap.String("name", publisher.Name), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrCreatePublisher",
			Message: "Не удалось создать издательство: " + err.Error(),
		}
	}

	s.logger.Info("publisher created successfully", zap.Int("publisher_id", createdPublisher.ID), zap.String("name", createdPublisher.Name))
	return createdPublisher, nil
}

func (s *PublisherService) GetPublisher(ctx context.Context, conn *pgx.Conn, id int) (*domain.Publisher, error) {
	s.logger.Debug("get publisher started", zap.Int("publisher_id", id))

	if id <= 0 {
		s.logger.Warn("invalid publisher id", zap.Int("publisher_id", id))
		return nil, errors.BusinessError{
			Code:    "ErrInvalidID",
			Message: "ID должен быть положительным числом",
		}
	}

	publisher, err := s.publisherRepo.GetByID(ctx, conn, id)
	if err != nil {
		s.logger.Warn("publisher not found", zap.Int("publisher_id", id), zap.Error(err))
		return nil, errors.NotFoundError{
			Entity: "Publisher",
			ID:     id,
		}
	}

	s.logger.Debug("get publisher finished", zap.Int("publisher_id", publisher.ID))
	return &publisher, nil
}

func (s *PublisherService) UpdatePublisher(ctx context.Context, conn *pgx.Conn, id int, updates map[string]interface{}) (*domain.Publisher, error) {
	s.logger.Info("update publisher started", zap.Int("publisher_id", id))

	existingPublisher, err := s.publisherRepo.GetByID(ctx, conn, id)
	if err != nil {
		s.logger.Warn("publisher not found for update", zap.Int("publisher_id", id), zap.Error(err))
		return nil, errors.NotFoundError{
			Entity: "Publisher",
			ID:     id,
		}
	}

	if name, ok := updates["name"].(string); ok {
		existingPublisher.Name = name
	}
	if address, ok := updates["address"].(string); ok {
		existingPublisher.Address = address
	}
	if phone, ok := updates["phone"].(string); ok {
		existingPublisher.Phone = phone
	}

	if err := existingPublisher.Validate(); err != nil {
		s.logger.Warn("publisher validation failed on update", zap.Int("publisher_id", id), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "validation_error",
			Message: "Ошибка валидации: " + err.Error(),
		}
	}

	exists, err := s.publisherRepo.ExistsByNameExcludeID(ctx, conn, existingPublisher.Name, id)
	if err != nil {
		s.logger.Error("failed to check publisher existence by name exclude id", zap.Int("publisher_id", id), zap.String("name", existingPublisher.Name), zap.Error(err))
		return nil, err
	}
	if exists {
		s.logger.Warn("publisher already exists with same name", zap.Int("publisher_id", id), zap.String("name", existingPublisher.Name))
		return nil, errors.BusinessError{
			Code:    "ErrPublisherAlreadyExists",
			Message: fmt.Sprintf("Издательство с названием '%s' уже существует", existingPublisher.Name),
		}
	}

	if err := s.publisherRepo.Update(ctx, conn, id, &existingPublisher); err != nil {
		s.logger.Error("failed to update publisher", zap.Int("publisher_id", id), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrPublisherUpdate",
			Message: "Не удалось обновить издательство: " + err.Error(),
		}
	}

	s.logger.Info("publisher updated successfully", zap.Int("publisher_id", id))
	return &existingPublisher, nil
}

func (s *PublisherService) DeletePublisher(ctx context.Context, conn *pgx.Conn, id int) error {
	s.logger.Info("delete publisher started", zap.Int("publisher_id", id))

	_, err := s.publisherRepo.GetByID(ctx, conn, id)
	if err != nil {
		s.logger.Warn("publisher not found for delete", zap.Int("publisher_id", id), zap.Error(err))
		return errors.NotFoundError{
			Entity: "Publisher",
			ID:     id,
		}
	}

	books, err := s.bookRepo.GetByPublisherID(ctx, conn, id, 1, 1)
	if err != nil {
		s.logger.Error("failed to get publisher books", zap.Int("publisher_id", id), zap.Error(err))
		return err
	}
	if len(books) > 0 {
		s.logger.Warn("publisher has books, cannot delete", zap.Int("publisher_id", id), zap.Int("book_count", len(books)))
		return errors.BusinessError{
			Code:    "ErrPublisherHasBooks",
			Message: fmt.Sprintf("Нельзя удалить издательство, у него есть %d книг", len(books)),
		}
	}

	if err := s.publisherRepo.Delete(ctx, conn, id); err != nil {
		s.logger.Error("failed to delete publisher", zap.Int("publisher_id", id), zap.Error(err))
		return errors.BusinessError{
			Code:    "ErrPublisherDelete",
			Message: "Не удалось удалить издательство: " + err.Error(),
		}
	}

	s.logger.Info("publisher deleted successfully", zap.Int("publisher_id", id))
	return nil
}

func (s *PublisherService) ListPublishers(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Publisher, int, error) {
	limit, offset = limitOffset(limit, offset)
	s.logger.Debug("list publishers started", zap.Int("limit", limit), zap.Int("offset", offset))

	publishers, err := s.publisherRepo.List(ctx, conn, limit, offset)
	if err != nil {
		s.logger.Error("failed to list publishers", zap.Int("limit", limit), zap.Int("offset", offset), zap.Error(err))
		return nil, 0, errors.BusinessError{
			Code:    "ErrPublisherList",
			Message: "Не удалось получить список издательств: " + err.Error(),
		}
	}

	total, err := s.publisherRepo.Count(ctx, conn)
	if err != nil {
		s.logger.Error("failed to count publishers", zap.Error(err))
		return nil, 0, errors.BusinessError{
			Code:    "ErrPublisherCount",
			Message: "Не удалось подсчитать издательства: " + err.Error(),
		}
	}

	s.logger.Debug("list publishers finished", zap.Int("returned", len(publishers)), zap.Int("total", total))
	return publishers, total, nil
}

func (s *PublisherService) SearchPublishers(ctx context.Context, conn *pgx.Conn, column, search string, limit, offset int) ([]domain.Publisher, int, error) {
	limit, offset = limitOffset(limit, offset)
	s.logger.Debug("search publishers started", zap.String("column", column), zap.String("search", search), zap.Int("limit", limit), zap.Int("offset", offset))

	if search == "" {
		s.logger.Warn("empty search query")
		return nil, 0, errors.BusinessError{
			Code:    "ErrPublisherSearch",
			Message: "Поисковый запрос не может быть пустым",
		}
	}

	publishers, count, err := s.publisherRepo.Search(ctx, conn, column, search, limit, offset)
	if err != nil {
		s.logger.Error("failed to search publishers", zap.String("column", column), zap.String("search", search), zap.Error(err))
		return nil, 0, errors.BusinessError{
			Code:    "ErrPublisherSearch",
			Message: "Ошибка при поиске издательств: " + err.Error(),
		}
	}

	s.logger.Debug("search publishers finished", zap.Int("found", count))
	return publishers, count, nil
}

func (s *PublisherService) GetPublisherBooks(ctx context.Context, conn *pgx.Conn, publisherID, limit, offset int) ([]domain.Book, int, error) {
	limit, offset = limitOffset(limit, offset)
	s.logger.Debug("get publisher books started", zap.Int("publisher_id", publisherID), zap.Int("limit", limit), zap.Int("offset", offset))

	exists, err := s.publisherRepo.Exists(ctx, conn, publisherID)
	if err != nil {
		s.logger.Error("failed to check publisher existence", zap.Int("publisher_id", publisherID), zap.Error(err))
		return nil, 0, err
	}
	if !exists {
		s.logger.Warn("publisher not found for get books", zap.Int("publisher_id", publisherID))
		return nil, 0, errors.NotFoundError{
			Entity: "Publisher",
			ID:     publisherID,
		}
	}

	books, err := s.bookRepo.GetByPublisherID(ctx, conn, publisherID, limit, offset)
	if err != nil {
		s.logger.Error("failed to get publisher books", zap.Int("publisher_id", publisherID), zap.Error(err))
		return nil, 0, errors.BusinessError{
			Code:    "ErrPublisherGetBooks",
			Message: "Не удалось получить книги издательства: " + err.Error(),
		}
	}

	total, err := s.bookRepo.CountByPublisherID(ctx, conn, publisherID)
	if err != nil {
		s.logger.Error("failed to count publisher books", zap.Int("publisher_id", publisherID), zap.Error(err))
		return nil, 0, errors.BusinessError{
			Code:    "ErrPublisherCountBooks",
			Message: "Не удалось подсчитать книги издательства: " + err.Error(),
		}
	}

	s.logger.Debug("get publisher books finished", zap.Int("publisher_id", publisherID), zap.Int("books_count", len(books)), zap.Int("total", total))
	return books, total, nil
}

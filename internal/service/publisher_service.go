package service

import (
	"context"
	"fmt"
	"library/internal/domain"
	"library/internal/repository"
	"library/pkg/errors"

	"github.com/jackc/pgx/v5"
)

type PublisherService struct {
	publisherRepo repository.PublisherRepository
	bookRepo      repository.BookRepository
}

func NewPublisherService(
	publisherRepo repository.PublisherRepository,
	bookRepo repository.BookRepository,
) *PublisherService {
	return &PublisherService{
		publisherRepo: publisherRepo,
		bookRepo:      bookRepo,
	}
}

func (s *PublisherService) CreatePublisher(ctx context.Context, conn *pgx.Conn, publisher *domain.Publisher) (*domain.Publisher, error) {
	if err := publisher.Validate(); err != nil {
		return nil, errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации" + err.Error(),
		}
	}

	exists, err := s.publisherRepo.ExistsByName(ctx, conn, publisher.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.BusinessError{
			Code:    "ErrPublisherAlreadyExists",
			Message: fmt.Sprintf("Издательство с названием '%s' уже существует", publisher.Name),
		}
	}

	createdPublisher, err := s.publisherRepo.Create(ctx, conn, *publisher)
	if err != nil {
		return nil, errors.BusinessError{
			Code:    "ErrCreatePublisher",
			Message: "Не удалось создать издательство: " + err.Error(),
		}
	}

	return createdPublisher, nil
}

func (s *PublisherService) GetPublisher(ctx context.Context, conn *pgx.Conn, id int) (*domain.Publisher, error) {
	if id <= 0 {
		return nil, errors.BusinessError{
			Code:    "ErrInvalidID",
			Message: "ID должен быть положительным числом",
		}
	}

	publisher, err := s.publisherRepo.GetByID(ctx, conn, id)
	if err != nil {
		return nil, errors.NotFoundError{
			Entity: "Publisher",
			ID:     id,
		}
	}
	return &publisher, nil
}

func (s *PublisherService) UpdatePublisher(ctx context.Context, conn *pgx.Conn, id int, updates map[string]interface{}) (*domain.Publisher, error) {
	existingPublisher, err := s.publisherRepo.GetByID(ctx, conn, id)
	if err != nil {
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
		return nil, errors.BusinessError{
			Code:    "validation_error",
			Message: "Ошибка валидации: " + err.Error(),
		}
	}

	exists, err := s.publisherRepo.ExistsByNameExcludeID(ctx, conn, existingPublisher.Name, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.BusinessError{
			Code:    "ErrPublisherAlreadyExists",
			Message: fmt.Sprintf("Издательство с названием '%s' уже существует", existingPublisher.Name),
		}
	}

	if err := s.publisherRepo.Update(ctx, conn, id, existingPublisher); err != nil {
		return nil, errors.BusinessError{
			Code:    "ErrPublisherUpdate",
			Message: "Не удалось обновить издательство: " + err.Error(),
		}
	}

	return &existingPublisher, nil
}

func (s *PublisherService) DeletePublisher(ctx context.Context, conn *pgx.Conn, id int) error {
	_, err := s.publisherRepo.GetByID(ctx, conn, id)
	if err != nil {
		return errors.NotFoundError{
			Entity: "Publisher",
			ID:     id,
		}
	}

	books, err := s.bookRepo.GetByPublisherID(ctx, conn, id, 1, 1)
	if err != nil {
		return err
	}
	if len(books) > 0 {
		return errors.BusinessError{
			Code:    "ErrPublisherHasBooks",
			Message: fmt.Sprintf("Нельзя удалить издательство, у него есть %d книг", len(books)),
		}
	}

	if err := s.publisherRepo.Delete(ctx, conn, id); err != nil {
		return errors.BusinessError{
			Code:    "ErrPublisherDelete",
			Message: "Не удалось удалить издательство: " + err.Error(),
		}
	}

	return nil
}

func (s *PublisherService) ListPublishers(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Publisher, int, error) {
	limitOffset(limit, offset)

	publishers, err := s.publisherRepo.List(ctx, conn, limit, offset)
	if err != nil {
		return nil, 0, errors.BusinessError{
			Code:    "ErrPublisherList",
			Message: "Не удалось получить список издательств: " + err.Error(),
		}
	}

	total, err := s.publisherRepo.Count(ctx, conn)
	if err != nil {
		return nil, 0, errors.BusinessError{
			Code:    "ErrPublisherCount",
			Message: "Не удалось подсчитать издательства: " + err.Error(),
		}
	}

	return publishers, total, nil
}

func (s *PublisherService) SearchPublishers(ctx context.Context, conn *pgx.Conn, column, search string, limit, offset int) ([]domain.Publisher, int, error) {
	limitOffset(limit, offset)

	if search == "" {
		return nil, 0, errors.BusinessError{
			Code:    "ErrPublisherSearch",
			Message: "Поисковый запрос не может быть пустым",
		}
	}

	publishers, count, err := s.publisherRepo.Search(ctx, conn, column, search, limit, offset)
	if err != nil {
		return nil, 0, errors.BusinessError{
			Code:    "ErrPublisherSearch",
			Message: "Ошибка при поиске издательств: " + err.Error(),
		}
	}

	return publishers, count, nil
}

func (s *PublisherService) GetPublisherBooks(ctx context.Context, conn *pgx.Conn, publisherID, limit, offset int) ([]domain.Book, int, error) {
	limitOffset(limit, offset)

	exists, err := s.publisherRepo.Exists(ctx, conn, publisherID)
	if err != nil {
		return nil, 0, err
	}
	if !exists {
		return nil, 0, errors.NotFoundError{
			Entity: "Publisher",
			ID:     publisherID,
		}
	}

	books, err := s.bookRepo.GetByPublisherID(ctx, conn, publisherID, limit, offset)
	if err != nil {
		return nil, 0, errors.BusinessError{
			Code:    "ErrPublisherGetBooks",
			Message: "Не удалось получить книги издательства: " + err.Error(),
		}
	}

	total, err := s.bookRepo.CountByPublisherID(ctx, conn, publisherID)
	if err != nil {
		return nil, 0, errors.BusinessError{
			Code:    "ErrPublisherCountBooks",
			Message: "Не удалось подсчитать книги издательства: " + err.Error(),
		}
	}

	return books, total, nil
}

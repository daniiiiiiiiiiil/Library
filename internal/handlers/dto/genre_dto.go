package dto

import (
	"library/internal/domain"
	"library/pkg/errors"
	"strings"
)

type CreateGenreRequest struct {
	Name     string `json:"name"`
	ParentID *int   `json:"parent_id,omitempty"`
}

func (r *CreateGenreRequest) Validate() error {
	var errs errors.ValidationErrors
	if strings.TrimSpace(r.Name) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "name",
			Message: "Имя не может быть пустым",
		})
	}
	if len(r.Name) > 100 {
		errs = append(errs, errors.ValidationError{
			Field:   "name",
			Message: "Имя не может превышать 100 символов",
		})
	}
	if r.ParentID != nil && *r.ParentID < 0 {
		errs = append(errs, errors.ValidationError{
			Field:   "parent_id",
			Message: "parentID не может быть меньше нуля",
		})
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}
func (r *CreateGenreRequest) ToDomain() domain.Genre {
	return domain.NewGenre(
		0,
		r.Name,
		r.ParentID,
	)
}

type UpdateGenreRequest struct {
	Name     *string `json:"name"`
	ParentID *int    `json:"parent_id"`
}

func (r *UpdateGenreRequest) ToDomain(genreID int) (domain.Genre, error) {
	if genreID < 0 {
		return domain.Genre{}, errors.ValidationError{
			Field:   "genreID",
			Message: "ID жанра не может быть меньше 0",
		}
	}
	return domain.NewGenre(
		genreID,
		"",
		r.ParentID,
	), nil
}

type GenreResponse struct {
	ID             int    `json:"ID"`
	Name           string `json:"name"`
	ParentID       *int   `json:"parent_id"`
	ParentName     string `json:"parent_name"`
	SubgenresCount int    `json:"subgenres_count"`
	BooksCount     int    `json:"books_count"`
}

type GenreHierarchyResponse struct {
	ID       int                      `json:"ID"`
	Name     string                   `json:"name"`
	Children []GenreHierarchyResponse `json:"children"`
}

package service

import (
	"context"
	"fmt"
	"library/internal/domain"
	"library/internal/infrastructure/audit"
	"library/internal/repository"
	"library/pkg/errors"
	"strconv"

	"github.com/jackc/pgx/v5"
)

type SettingService struct {
	settingRepo repository.SettingRepository
	auditRepo   repository.AuditLogRepository
}

func NewSettingService(
	settingRepo repository.SettingRepository,
	auditRepo repository.AuditLogRepository,
) *SettingService {
	return &SettingService{
		settingRepo: settingRepo,
		auditRepo:   auditRepo,
	}
}

func (s *SettingService) CreateSetting(ctx context.Context, conn *pgx.Conn, setting *domain.Setting) (*domain.Setting, error) {
	if err := setting.Validate(); err != nil {
		return nil, errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации: " + err.Error(),
		}
	}

	exists, err := s.settingRepo.Exists(ctx, conn, setting.Key)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.BusinessError{
			Code:    "setting_already_exists",
			Message: fmt.Sprintf("Настройка с ключом '%s' уже существует", setting.Key),
		}
	}

	if err := s.settingRepo.CreateSetting(ctx, conn, *setting); err != nil {
		return nil, errors.BusinessError{
			Code:    "create_setting_error",
			Message: "Не удалось создать настройку: " + err.Error(),
		}
	}

	if err := s.auditRepo.CreateAuditLog(ctx, conn, audit.AuditLog{
		UserID:     nil,
		Action:     "CREATE",
		EntityType: "setting",
		EntityID:   setting.ID,
	}); err != nil {
		return nil, err
	}

	return setting, nil
}

func (s *SettingService) GetSetting(ctx context.Context, conn *pgx.Conn, id int) (*domain.Setting, error) {
	if id <= 0 {
		return nil, errors.BusinessError{
			Code:    "invalid_id",
			Message: "ID должен быть положительным числом",
		}
	}

	setting, err := s.settingRepo.GetByID(ctx, conn, id)
	if err != nil {
		return nil, errors.NotFoundError{
			Entity: "Setting",
			ID:     id,
		}
	}
	return &setting, nil
}

func (s *SettingService) GetSettingByKey(ctx context.Context, conn *pgx.Conn, key string) (*domain.Setting, error) {
	if key == "" {
		return nil, errors.BusinessError{
			Code:    "empty_key",
			Message: "Ключ не может быть пустым",
		}
	}

	setting, err := s.settingRepo.GetByKey(ctx, conn, key)
	if err != nil {
		return nil, errors.NotFoundError{
			Entity: "Setting",
			ID:     0,
		}
	}
	return &setting, nil
}

func (s *SettingService) GetSettingValue(ctx context.Context, conn *pgx.Conn, key string) (string, error) {
	setting, err := s.GetSettingByKey(ctx, conn, key)
	if err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (s *SettingService) GetSettingInt(ctx context.Context, conn *pgx.Conn, key string) (int, error) {
	value, err := s.GetSettingValue(ctx, conn, key)
	if err != nil {
		return 0, err
	}

	intVal, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.BusinessError{
			Code:    "invalid_type",
			Message: fmt.Sprintf("Значение '%s' не является числом", value),
		}
	}
	return intVal, nil
}

func (s *SettingService) GetSettingFloat(ctx context.Context, conn *pgx.Conn, key string) (float64, error) {
	value, err := s.GetSettingValue(ctx, conn, key)
	if err != nil {
		return 0, err
	}

	floatVal, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, errors.BusinessError{
			Code:    "invalid_type",
			Message: fmt.Sprintf("Значение '%s' не является числом", value),
		}
	}
	return floatVal, nil
}

func (s *SettingService) GetSettingBool(ctx context.Context, conn *pgx.Conn, key string) (bool, error) {
	value, err := s.GetSettingValue(ctx, conn, key)
	if err != nil {
		return false, err
	}

	switch value {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	default:
		return false, errors.BusinessError{
			Code:    "invalid_type",
			Message: fmt.Sprintf("Значение '%s' не является логическим", value),
		}
	}
}

func (s *SettingService) UpdateSetting(ctx context.Context, conn *pgx.Conn, id int, updates map[string]interface{}) (*domain.Setting, error) {
	existingSetting, err := s.settingRepo.GetByID(ctx, conn, id)
	if err != nil {
		return nil, errors.NotFoundError{
			Entity: "Setting",
			ID:     id,
		}
	}

	if value, ok := updates["value"].(string); ok {
		existingSetting.Value = value
	}
	if description, ok := updates["description"].(string); ok {
		existingSetting.Description = description
	}

	if err := existingSetting.Validate(); err != nil {
		return nil, errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации: " + err.Error(),
		}
	}

	if err := s.settingRepo.Update(ctx, conn, existingSetting); err != nil {
		return nil, errors.BusinessError{
			Code:    "update_setting_error",
			Message: "Не удалось обновить настройку: " + err.Error(),
		}
	}

	return &existingSetting, nil
}

func (s *SettingService) UpdateSettingByKey(ctx context.Context, conn *pgx.Conn, key, value string) error {
	if key == "" {
		return errors.BusinessError{
			Code:    "empty_key",
			Message: "Ключ не может быть пустым",
		}
	}

	exists, err := s.settingRepo.Exists(ctx, conn, key)
	if err != nil {
		return err
	}
	if !exists {
		return errors.NotFoundError{
			Entity: "Setting",
			ID:     0,
		}
	}

	if err := s.settingRepo.UpdateByKey(ctx, conn, key, value); err != nil {
		return errors.BusinessError{
			Code:    "update_setting_error",
			Message: "Не удалось обновить настройку: " + err.Error(),
		}
	}

	return nil
}

func (s *SettingService) DeleteSetting(ctx context.Context, conn *pgx.Conn, id int) error {
	_, err := s.settingRepo.GetByID(ctx, conn, id)
	if err != nil {
		return errors.NotFoundError{
			Entity: "Setting",
			ID:     id,
		}
	}

	if err := s.settingRepo.Delete(ctx, conn, id); err != nil {
		return errors.BusinessError{
			Code:    "delete_setting_error",
			Message: "Не удалось удалить настройку: " + err.Error(),
		}
	}

	return nil
}

func (s *SettingService) DeleteSettingByKey(ctx context.Context, conn *pgx.Conn, key string) error {
	if key == "" {
		return errors.BusinessError{
			Code:    "empty_key",
			Message: "Ключ не может быть пустым",
		}
	}

	exists, err := s.settingRepo.Exists(ctx, conn, key)
	if err != nil {
		return err
	}
	if !exists {
		return errors.NotFoundError{
			Entity: "Setting",
			ID:     0,
		}
	}

	if err := s.settingRepo.DeleteByKey(ctx, conn, key); err != nil {
		return errors.BusinessError{
			Code:    "delete_setting_error",
			Message: "Не удалось удалить настройку: " + err.Error(),
		}
	}

	return nil
}

func (s *SettingService) ListSettings(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Setting, int, error) {
	limitOffset(limit, offset)

	settingsList, err := s.settingRepo.List(ctx, conn, limit, offset)
	if err != nil {
		return nil, 0, errors.BusinessError{
			Code:    "list_settings_error",
			Message: "Не удалось получить список настроек: " + err.Error(),
		}
	}

	return settingsList, len(settingsList), nil
}

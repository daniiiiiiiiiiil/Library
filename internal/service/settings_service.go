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
	"go.uber.org/zap"
)

type SettingService struct {
	settingRepo repository.SettingRepository
	auditRepo   repository.AuditLogRepository
	logger      *zap.Logger
}

func NewSettingService(
	settingRepo repository.SettingRepository,
	auditRepo repository.AuditLogRepository,
	logger *zap.Logger,
) *SettingService {
	return &SettingService{
		settingRepo: settingRepo,
		auditRepo:   auditRepo,
		logger:      logger,
	}
}

func (s *SettingService) CreateSetting(ctx context.Context, conn *pgx.Conn, setting *domain.Setting) (*domain.Setting, error) {
	s.logger.Info("create setting started", zap.String("key", setting.Key), zap.String("value", setting.Value))

	if err := setting.Validate(); err != nil {
		s.logger.Warn("setting validation failed", zap.String("key", setting.Key), zap.Error(err))
		return nil, errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации: " + err.Error(),
		}
	}

	exists, err := s.settingRepo.Exists(ctx, conn, setting.Key)
	if err != nil {
		s.logger.Error("failed to check setting existence", zap.String("key", setting.Key), zap.Error(err))
		return nil, err
	}
	if exists {
		s.logger.Warn("setting already exists", zap.String("key", setting.Key))
		return nil, errors.BusinessError{
			Code:    "setting_already_exists",
			Message: fmt.Sprintf("Настройка с ключом '%s' уже существует", setting.Key),
		}
	}

	if err := s.settingRepo.CreateSetting(ctx, conn, *setting); err != nil {
		s.logger.Error("failed to create setting", zap.String("key", setting.Key), zap.Error(err))
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
		s.logger.Error("failed to create audit log for setting create", zap.String("key", setting.Key), zap.Error(err))
		return nil, err
	}

	s.logger.Info("setting created successfully", zap.Int("setting_id", setting.ID), zap.String("key", setting.Key))
	return setting, nil
}

func (s *SettingService) GetSetting(ctx context.Context, conn *pgx.Conn, id int) (*domain.Setting, error) {
	s.logger.Debug("get setting started", zap.Int("setting_id", id))

	if id <= 0 {
		s.logger.Warn("invalid setting id", zap.Int("setting_id", id))
		return nil, errors.BusinessError{
			Code:    "invalid_id",
			Message: "ID должен быть положительным числом",
		}
	}

	setting, err := s.settingRepo.GetByID(ctx, conn, id)
	if err != nil {
		s.logger.Warn("setting not found", zap.Int("setting_id", id), zap.Error(err))
		return nil, errors.NotFoundError{
			Entity: "Setting",
			ID:     id,
		}
	}

	s.logger.Debug("get setting finished", zap.Int("setting_id", setting.ID), zap.String("key", setting.Key))
	return &setting, nil
}

func (s *SettingService) GetSettingByKey(ctx context.Context, conn *pgx.Conn, key string) (*domain.Setting, error) {
	s.logger.Debug("get setting by key started", zap.String("key", key))

	if key == "" {
		s.logger.Warn("empty key")
		return nil, errors.BusinessError{
			Code:    "empty_key",
			Message: "Ключ не может быть пустым",
		}
	}

	setting, err := s.settingRepo.GetByKey(ctx, conn, key)
	if err != nil {
		s.logger.Warn("setting not found by key", zap.String("key", key), zap.Error(err))
		return nil, errors.NotFoundError{
			Entity: "Setting",
			ID:     0,
		}
	}

	s.logger.Debug("get setting by key finished", zap.String("key", key), zap.String("value", setting.Value))
	return &setting, nil
}

func (s *SettingService) GetSettingValue(ctx context.Context, conn *pgx.Conn, key string) (string, error) {
	s.logger.Debug("get setting value started", zap.String("key", key))

	setting, err := s.GetSettingByKey(ctx, conn, key)
	if err != nil {
		s.logger.Warn("failed to get setting value", zap.String("key", key), zap.Error(err))
		return "", err
	}

	s.logger.Debug("get setting value finished", zap.String("key", key), zap.String("value", setting.Value))
	return setting.Value, nil
}

func (s *SettingService) GetSettingInt(ctx context.Context, conn *pgx.Conn, key string) (int, error) {
	s.logger.Debug("get setting int started", zap.String("key", key))

	value, err := s.GetSettingValue(ctx, conn, key)
	if err != nil {
		s.logger.Warn("failed to get setting value as int", zap.String("key", key), zap.Error(err))
		return 0, err
	}

	intVal, err := strconv.Atoi(value)
	if err != nil {
		s.logger.Warn("failed to convert setting to int", zap.String("key", key), zap.String("value", value), zap.Error(err))
		return 0, errors.BusinessError{
			Code:    "invalid_type",
			Message: fmt.Sprintf("Значение '%s' не является числом", value),
		}
	}

	s.logger.Debug("get setting int finished", zap.String("key", key), zap.Int("value", intVal))
	return intVal, nil
}

func (s *SettingService) GetSettingFloat(ctx context.Context, conn *pgx.Conn, key string) (float64, error) {
	s.logger.Debug("get setting float started", zap.String("key", key))

	value, err := s.GetSettingValue(ctx, conn, key)
	if err != nil {
		s.logger.Warn("failed to get setting value as float", zap.String("key", key), zap.Error(err))
		return 0, err
	}

	floatVal, err := strconv.ParseFloat(value, 64)
	if err != nil {
		s.logger.Warn("failed to convert setting to float", zap.String("key", key), zap.String("value", value), zap.Error(err))
		return 0, errors.BusinessError{
			Code:    "invalid_type",
			Message: fmt.Sprintf("Значение '%s' не является числом", value),
		}
	}

	s.logger.Debug("get setting float finished", zap.String("key", key), zap.Float64("value", floatVal))
	return floatVal, nil
}

func (s *SettingService) GetSettingBool(ctx context.Context, conn *pgx.Conn, key string) (bool, error) {
	s.logger.Debug("get setting bool started", zap.String("key", key))

	value, err := s.GetSettingValue(ctx, conn, key)
	if err != nil {
		s.logger.Warn("failed to get setting value as bool", zap.String("key", key), zap.Error(err))
		return false, err
	}

	switch value {
	case "true", "1", "yes", "on":
		s.logger.Debug("get setting bool finished", zap.String("key", key), zap.Bool("value", true))
		return true, nil
	case "false", "0", "no", "off":
		s.logger.Debug("get setting bool finished", zap.String("key", key), zap.Bool("value", false))
		return false, nil
	default:
		s.logger.Warn("invalid boolean value", zap.String("key", key), zap.String("value", value))
		return false, errors.BusinessError{
			Code:    "invalid_type",
			Message: fmt.Sprintf("Значение '%s' не является логическим", value),
		}
	}
}

func (s *SettingService) UpdateSetting(ctx context.Context, conn *pgx.Conn, id int, updates map[string]interface{}) (*domain.Setting, error) {
	s.logger.Info("update setting started", zap.Int("setting_id", id))

	existingSetting, err := s.settingRepo.GetByID(ctx, conn, id)
	if err != nil {
		s.logger.Warn("setting not found for update", zap.Int("setting_id", id), zap.Error(err))
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
		s.logger.Warn("setting validation failed on update", zap.Int("setting_id", id), zap.String("key", existingSetting.Key), zap.Error(err))
		return nil, errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации: " + err.Error(),
		}
	}

	if err := s.settingRepo.Update(ctx, conn, existingSetting); err != nil {
		s.logger.Error("failed to update setting", zap.Int("setting_id", id), zap.String("key", existingSetting.Key), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "update_setting_error",
			Message: "Не удалось обновить настройку: " + err.Error(),
		}
	}

	s.logger.Info("setting updated successfully", zap.Int("setting_id", id), zap.String("key", existingSetting.Key))
	return &existingSetting, nil
}

func (s *SettingService) UpdateSettingByKey(ctx context.Context, conn *pgx.Conn, key, value string) error {
	s.logger.Info("update setting by key started", zap.String("key", key), zap.String("value", value))

	if key == "" {
		s.logger.Warn("empty key")
		return errors.BusinessError{
			Code:    "empty_key",
			Message: "Ключ не может быть пустым",
		}
	}

	exists, err := s.settingRepo.Exists(ctx, conn, key)
	if err != nil {
		s.logger.Error("failed to check setting existence", zap.String("key", key), zap.Error(err))
		return err
	}
	if !exists {
		s.logger.Warn("setting not found for update by key", zap.String("key", key))
		return errors.NotFoundError{
			Entity: "Setting",
			ID:     0,
		}
	}

	if err := s.settingRepo.UpdateByKey(ctx, conn, key, value); err != nil {
		s.logger.Error("failed to update setting by key", zap.String("key", key), zap.Error(err))
		return errors.BusinessError{
			Code:    "update_setting_error",
			Message: "Не удалось обновить настройку: " + err.Error(),
		}
	}

	s.logger.Info("setting updated by key successfully", zap.String("key", key))
	return nil
}

func (s *SettingService) DeleteSetting(ctx context.Context, conn *pgx.Conn, id int) error {
	s.logger.Info("delete setting started", zap.Int("setting_id", id))

	_, err := s.settingRepo.GetByID(ctx, conn, id)
	if err != nil {
		s.logger.Warn("setting not found for delete", zap.Int("setting_id", id), zap.Error(err))
		return errors.NotFoundError{
			Entity: "Setting",
			ID:     id,
		}
	}

	if err := s.settingRepo.Delete(ctx, conn, id); err != nil {
		s.logger.Error("failed to delete setting", zap.Int("setting_id", id), zap.Error(err))
		return errors.BusinessError{
			Code:    "delete_setting_error",
			Message: "Не удалось удалить настройку: " + err.Error(),
		}
	}

	s.logger.Info("setting deleted successfully", zap.Int("setting_id", id))
	return nil
}

func (s *SettingService) DeleteSettingByKey(ctx context.Context, conn *pgx.Conn, key string) error {
	s.logger.Info("delete setting by key started", zap.String("key", key))

	if key == "" {
		s.logger.Warn("empty key")
		return errors.BusinessError{
			Code:    "empty_key",
			Message: "Ключ не может быть пустым",
		}
	}

	exists, err := s.settingRepo.Exists(ctx, conn, key)
	if err != nil {
		s.logger.Error("failed to check setting existence", zap.String("key", key), zap.Error(err))
		return err
	}
	if !exists {
		s.logger.Warn("setting not found for delete by key", zap.String("key", key))
		return errors.NotFoundError{
			Entity: "Setting",
			ID:     0,
		}
	}

	if err := s.settingRepo.DeleteByKey(ctx, conn, key); err != nil {
		s.logger.Error("failed to delete setting by key", zap.String("key", key), zap.Error(err))
		return errors.BusinessError{
			Code:    "delete_setting_error",
			Message: "Не удалось удалить настройку: " + err.Error(),
		}
	}

	s.logger.Info("setting deleted by key successfully", zap.String("key", key))
	return nil
}

func (s *SettingService) ListSettings(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Setting, int, error) {
	limit, offset = limitOffset(limit, offset)
	s.logger.Debug("list settings started", zap.Int("limit", limit), zap.Int("offset", offset))

	settingsList, err := s.settingRepo.List(ctx, conn, limit, offset)
	if err != nil {
		s.logger.Error("failed to list settings", zap.Int("limit", limit), zap.Int("offset", offset), zap.Error(err))
		return nil, 0, errors.BusinessError{
			Code:    "list_settings_error",
			Message: "Не удалось получить список настроек: " + err.Error(),
		}
	}

	s.logger.Debug("list settings finished", zap.Int("returned", len(settingsList)))
	return settingsList, len(settingsList), nil
}

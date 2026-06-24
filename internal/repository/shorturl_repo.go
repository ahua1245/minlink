package repository

import (
	"minlink/internal/model"
	"time"

	"github.com/jinzhu/gorm"
)

type ShortURLRepository interface {
	Create(shortURL *model.ShortURL) error
	FindByShortCode(shortCode string) (*model.ShortURL, error)
	FindByID(id uint) (*model.ShortURL, error)
	FindByLongURL(longURL string) (*model.ShortURL, error)
	Update(shortURL *model.ShortURL) error
	Delete(shortCode string) error
	UpdateVisitCount(shortCode string, count int) error
	List(userID uint, page, limit int) ([]model.ShortURL, error)
	Count(userID uint) (int, error)
	CreateVisitLog(log *model.VisitLog) error
	GetTodayVisits(shortCode string) (int, error)
}

type shortURLRepository struct {
	db *gorm.DB
}

func NewShortURLRepository(db *gorm.DB) ShortURLRepository {
	return &shortURLRepository{db: db}
}

func (r *shortURLRepository) Create(shortURL *model.ShortURL) error {
	return r.db.Create(shortURL).Error
}

func (r *shortURLRepository) FindByShortCode(shortCode string) (*model.ShortURL, error) {
	var shortURL model.ShortURL
	err := r.db.Where("short_code = ?", shortCode).First(&shortURL).Error
	if err != nil {
		return nil, err
	}
	return &shortURL, nil
}

func (r *shortURLRepository) FindByID(id uint) (*model.ShortURL, error) {
	var shortURL model.ShortURL
	err := r.db.Where("id = ?", id).First(&shortURL).Error
	if err != nil {
		return nil, err
	}
	return &shortURL, nil
}

func (r *shortURLRepository) FindByLongURL(longURL string) (*model.ShortURL, error) {
	var shortURL model.ShortURL
	err := r.db.Where("long_url = ?", longURL).First(&shortURL).Error
	if err != nil {
		return nil, err
	}
	return &shortURL, nil
}

func (r *shortURLRepository) Update(shortURL *model.ShortURL) error {
	return r.db.Save(shortURL).Error
}

func (r *shortURLRepository) Delete(shortCode string) error {
	return r.db.Where("short_code = ?", shortCode).Delete(&model.ShortURL{}).Error
}

func (r *shortURLRepository) UpdateVisitCount(shortCode string, count int) error {
	return r.db.Exec("UPDATE short_urls SET visit_count = visit_count + ? WHERE short_code = ?", count, shortCode).Error
}

func (r *shortURLRepository) List(userID uint, page, limit int) ([]model.ShortURL, error) {
	var shortURLs []model.ShortURL
	offset := (page - 1) * limit
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(limit).Find(&shortURLs).Error
	if err != nil {
		return nil, err
	}
	return shortURLs, nil
}

func (r *shortURLRepository) Count(userID uint) (int, error) {
	var count int
	err := r.db.Model(&model.ShortURL{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *shortURLRepository) CreateVisitLog(log *model.VisitLog) error {
	return r.db.Create(log).Error
}

func (r *shortURLRepository) GetTodayVisits(shortCode string) (int, error) {
	var count int
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := r.db.Model(&model.VisitLog{}).Where("short_code = ? AND created_at >= ? AND created_at < ?", shortCode, startOfDay, endOfDay).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

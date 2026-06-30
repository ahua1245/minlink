package service

import (
	"errors"
	"log"
	"minlink/internal/model"
	"minlink/internal/repository"
	"minlink/internal/util"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
)

var (
	ErrShortCodeNotFound = errors.New("short code not found")
	ErrShortCodeExpired  = errors.New("short code expired")
	ErrInvalidURL        = errors.New("invalid URL")
)

type ShortURLService interface {
	CreateShortURL(longURL string, expireDays int, userID uint, username string, name, remark string) (*model.ShortURL, error)
	GetLongURL(shortCode string) (string, error)
	RecordVisit(shortCode, ip, userAgent, referer string) error
	GetStats(shortCode string) (map[string]interface{}, error)
	UpdateStatus(shortCode string, status int) error
	DeleteShortURL(shortCode string) error
	ListShortURLs(userID uint, page, limit int) ([]model.ShortURL, int, error)
}

type shortURLService struct {
	repo    repository.ShortURLRepository
	counter *CounterAggregator
}

type CounterAggregator struct {
	mu   sync.Mutex
	data map[string]int64
}

func NewCounterAggregator() *CounterAggregator {
	c := &CounterAggregator{
		data: make(map[string]int64),
	}
	go c.startFlushLoop()
	return c
}

func (c *CounterAggregator) Add(shortCode string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[shortCode]++
}

func (c *CounterAggregator) Flush(repo repository.ShortURLRepository) {
	c.mu.Lock()
	batch := make(map[string]int64)
	for k, v := range c.data {
		batch[k] = v
	}
	c.data = make(map[string]int64)
	c.mu.Unlock()

	for code, count := range batch {
		err := repo.UpdateVisitCount(code, int(count))
		if err != nil {
			log.Printf("Failed to update visit count for %s: %v", code, err)
		}
	}
}

func (c *CounterAggregator) startFlushLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
	}
}

func NewShortURLService(db *gorm.DB) ShortURLService {
	repo := repository.NewShortURLRepository(db)
	service := &shortURLService{
		repo:    repo,
		counter: NewCounterAggregator(),
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			service.counter.Flush(repo)
		}
	}()

	return service
}

func (s *shortURLService) CreateShortURL(longURL string, expireDays int, userID uint, username string, name, remark string) (*model.ShortURL, error) {
	if !util.IsValidURL(longURL) {
		return nil, ErrInvalidURL
	}

	// 检查是否已存在相同的长链接
	existing, err := s.repo.FindByLongURL(longURL)
	if err == nil && existing != nil && existing.Status == model.StatusActive {
		return existing, nil
	}

	shortCode := util.GenerateShortCode()

	now := time.Now()
	var expireAt *time.Time
	if expireDays > 0 {
		expire := now.Add(time.Duration(expireDays) * 24 * time.Hour)
		expireAt = &expire
	}

	// 创建人用户名：登录用户用 username，游客用 "guest"
	createdBy := username
	if createdBy == "" {
		createdBy = "guest"
	}

	shortURL := &model.ShortURL{
		ShortCode: shortCode,
		Name:      name,
		Remark:    remark,
		LongURL:   longURL,
		UserID:    userID,
		CreatedBy: createdBy,
		ExpireAt:  expireAt,
		Status:    model.StatusActive,
	}

	if err := s.repo.Create(shortURL); err != nil {
		return nil, err
	}

	return shortURL, nil
}

func (s *shortURLService) GetLongURL(shortCode string) (string, error) {
	if !util.IsValidShortCode(shortCode) {
		return "", ErrShortCodeNotFound
	}

	shortURL, err := s.repo.FindByShortCode(shortCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrShortCodeNotFound
		}
		return "", err
	}

	if shortURL.Status != model.StatusActive {
		return "", ErrShortCodeNotFound
	}

	if shortURL.ExpireAt != nil && shortURL.ExpireAt.Before(time.Now()) {
		shortURL.Status = model.StatusExpired
		s.repo.Update(shortURL)
		return "", ErrShortCodeExpired
	}

	s.counter.Add(shortCode)
	return shortURL.LongURL, nil
}

func (s *shortURLService) RecordVisit(shortCode, ip, userAgent, referer string) error {
	visitLog := &model.VisitLog{
		ShortCode: shortCode,
		IP:        ip,
		UserAgent: userAgent,
		Referer:   referer,
	}
	return s.repo.CreateVisitLog(visitLog)
}

func (s *shortURLService) GetStats(shortCode string) (map[string]interface{}, error) {
	shortURL, err := s.repo.FindByShortCode(shortCode)
	if err != nil {
		return nil, ErrShortCodeNotFound
	}

	todayVisits, err := s.repo.GetTodayVisits(shortCode)
	if err != nil {
		log.Printf("Failed to get today visits: %v", err)
		todayVisits = 0
	}

	return map[string]interface{}{
		"short_code":   shortURL.ShortCode,
		"long_url":     shortURL.LongURL,
		"total_visits": shortURL.VisitCount,
		"today_visits": todayVisits,
		"created_at":   shortURL.CreatedAt,
		"status":       shortURL.Status,
		"expire_at":    shortURL.ExpireAt,
	}, nil
}

func (s *shortURLService) UpdateStatus(shortCode string, status int) error {
	shortURL, err := s.repo.FindByShortCode(shortCode)
	if err != nil {
		return ErrShortCodeNotFound
	}

	shortURL.Status = status
	return s.repo.Update(shortURL)
}

func (s *shortURLService) DeleteShortURL(shortCode string) error {
	err := s.repo.Delete(shortCode)
	if err != nil {
		return err
	}

	return nil
}

func (s *shortURLService) ListShortURLs(userID uint, page, limit int) ([]model.ShortURL, int, error) {
	shortURLs, err := s.repo.List(userID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.Count(userID)
	if err != nil {
		return nil, 0, err
	}

	return shortURLs, total, nil
}

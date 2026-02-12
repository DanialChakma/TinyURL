package services

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"go.mod/initializers"
	"go.mod/models"
	"go.mod/repo"
)

type URLService struct {
	repo       *repo.URLRepository
	tenantRepo repo.TenantRepository
	cache      *Cache
	idGen      *IDGenerator
}

func NewURLService(
	urlRepo *repo.URLRepository,
	tenantRepo repo.TenantRepository,
	cache *Cache,
	idGen *IDGenerator,
) *URLService {
	return &URLService{
		repo:       urlRepo,
		tenantRepo: tenantRepo,
		cache:      cache,
		idGen:      idGen,
	}
}

// ------------------------
// CREATE STATEFUL URL
// ------------------------

func (s *URLService) Create(ctx context.Context, longURL, tenantID string) (string, error) {

	storedURL := longURL

	// Optimize storage if tenant exists
	if tenantID != "" {
		tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
		if err == nil && tenant != nil && tenant.BaseURL != "" {
			if parsed, err := url.Parse(longURL); err == nil {
				storedURL = parsed.RequestURI()
			}
		}
	}

	id := s.idGen.NextID()
	shortCode := EncodeID(id, tenantID)

	urlModel := &models.URL{
		ShortCode: shortCode,
		LongURL:   storedURL,
		TenantID:  tenantID,
		CreatedAt: time.Now().Unix(),
	}

	if err := s.repo.Create(ctx, urlModel); err != nil {
		return "", err
	}

	// Optional: cache full resolved URL instead
	_ = s.cache.Set(shortCode, storedURL, 24*time.Hour)

	// ✅ ALWAYS return your shortener domain
	return initializers.AppBaseURL + "/links/" + shortCode, nil
}

// ------------------------
// RESOLVE STATEFUL URL
// ------------------------
func (s *URLService) Resolve(ctx context.Context, shortCode string) (string, error) {

	// 1️⃣ Try cache
	cachedURL, _ := s.cache.Get(shortCode)
	if cachedURL != "" {
		// If cached value is absolute → return directly
		if strings.HasPrefix(cachedURL, "http://") ||
			strings.HasPrefix(cachedURL, "https://") {
			return cachedURL, nil
		}
		// If cached is relative, we cannot reconstruct tenant safely
		// So fall through to DB
	}

	// 2️⃣ Query DB (single source of truth)
	result, err := s.repo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return "", errors.New("url not found")
	}

	// 3️⃣ Build full URL properly
	fullURL, err := s.buildFullURL(ctx, result)
	if err != nil {
		return "", err
	}

	// 4️⃣ Cache the FULL resolved URL (important improvement)
	_ = s.cache.Set(shortCode, fullURL, 24*time.Hour)

	return fullURL, nil
}

// ------------------------
// BUILD FULL URL WITH TENANT BASE
// ------------------------

func (s *URLService) buildFullURL(ctx context.Context, urlModel *models.URL) (string, error) {

	// If already absolute URL → return directly
	if strings.HasPrefix(urlModel.LongURL, "http://") ||
		strings.HasPrefix(urlModel.LongURL, "https://") {
		return urlModel.LongURL, nil
	}

	// If tenant-based URL → prepend base URL safely
	if urlModel.TenantID != "" {
		tenant, err := s.tenantRepo.GetByID(ctx, urlModel.TenantID)
		if err == nil && tenant != nil && tenant.BaseURL != "" {

			base := strings.TrimRight(tenant.BaseURL, "/")
			path := strings.TrimLeft(urlModel.LongURL, "/")

			return base + "/" + path, nil
		}
	}

	// Fallback: return stored value
	return urlModel.LongURL, nil
}

// ------------------------
// CREATE STATELESS URL
// ------------------------
func (s *URLService) CreateStateless(ctx context.Context, longURL, tenantID string, trim bool) (string, error) {
	urlToEncode := longURL

	if trim && tenantID != "" {
		if tenant, err := s.tenantRepo.GetByID(ctx, tenantID); err == nil && tenant != nil && tenant.BaseURL != "" {
			parsed, err := url.Parse(longURL)
			if err == nil {
				urlToEncode = parsed.RequestURI()
			}
		}
	}

	key := initializers.MasterFeistelSecret
	if tenantID != "" {
		key = tenantID
	}

	compressed, err := Compress([]byte(urlToEncode))
	if err != nil {
		return "", err
	}

	encodedBytes := ObfuscateBytes(compressed, key, initializers.FeistelRounds)
	shortCode := Base62Encode(encodedBytes)

	return initializers.AppBaseURL + "/stateless/" + shortCode, nil
}

// ------------------------
// RESOLVE STATELESS URL
// ------------------------
func (s *URLService) ResolveStateless(shortCode, tenantID string) (string, error) {
	encodedBytes, err := Base62Decode(shortCode)
	if err != nil {
		return "", errors.New("invalid short url")
	}

	key := initializers.MasterFeistelSecret
	if tenantID != "" {
		key = tenantID
	}

	decoded := DeobfuscateBytes(encodedBytes, key, initializers.FeistelRounds)
	originalBytes, err := Decompress(decoded)
	if err != nil {
		return "", errors.New("corrupted data")
	}

	longURL := BytesToStringSafe(originalBytes)

	// ✅ If tenantID provided, prepend base URL
	if tenantID != "" {
		if tenant, err := s.tenantRepo.GetByID(context.Background(), tenantID); err == nil && tenant != nil && tenant.BaseURL != "" {
			longURL = tenant.BaseURL + longURL
		}
	}

	return longURL, nil
}

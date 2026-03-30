package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/codingninja/pob-management/internal/domain"
	"github.com/codingninja/pob-management/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

type AuthService struct {
	userRepository *repository.UserRepository
	tokenManager   *TokenManager
}

type RegisterInput struct {
	FirstName string
	LastName  string
	Email     string
	Password  string
	Role      domain.UserRole
}

type LoginInput struct {
	Email    string
	Password string
}

type UpdateProfileInput struct {
	FirstName string
	LastName  string
}

type ChangePasswordInput struct {
	CurrentPassword string
	NewPassword     string
}

type AuthTokens struct {
	AccessToken     string    `json:"access_token"`
	RefreshToken    string    `json:"refresh_token"`
	AccessTokenTTL  time.Time `json:"access_token_expires_at"`
	RefreshTokenTTL time.Time `json:"refresh_token_expires_at"`
}

func NewAuthService(userRepository *repository.UserRepository, tokenManager *TokenManager) *AuthService {
	return &AuthService{
		userRepository: userRepository,
		tokenManager:   tokenManager,
	}
}

func (s *AuthService) Initialize(ctx context.Context) error {
	return s.userRepository.EnsureIndexes(ctx)
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*domain.User, *AuthTokens, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	_, err := s.userRepository.FindByEmail(ctx, email)
	if err == nil {
		return nil, nil, ErrEmailAlreadyExists
	}
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:           bson.NewObjectID(),
		FirstName:    strings.TrimSpace(input.FirstName),
		LastName:     strings.TrimSpace(input.LastName),
		Email:        email,
		PasswordHash: string(passwordHash),
		Role:         input.Role,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepository.Create(ctx, user); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "e11000") {
			return nil, nil, ErrEmailAlreadyExists
		}
		return nil, nil, err
	}

	tokens, err := s.issueTokens(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*domain.User, *AuthTokens, error) {
	user, err := s.userRepository.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(input.Email)))
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	if !user.IsActive {
		return nil, nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	tokens, err := s.issueTokens(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*domain.User, *AuthTokens, error) {
	claims, err := s.tokenManager.Validate(refreshToken, RefreshTokenType)
	if err != nil {
		return nil, nil, ErrInvalidRefreshToken
	}

	userID, err := bson.ObjectIDFromHex(claims.Subject)
	if err != nil {
		return nil, nil, ErrInvalidRefreshToken
	}

	user, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		return nil, nil, ErrInvalidRefreshToken
	}

	if user.RefreshTokenHash == "" || user.RefreshTokenHash != s.tokenManager.HashToken(refreshToken) {
		return nil, nil, ErrInvalidRefreshToken
	}

	if user.RefreshTokenExpires == nil || user.RefreshTokenExpires.Before(time.Now().UTC()) {
		return nil, nil, ErrInvalidRefreshToken
	}

	tokens, err := s.issueTokens(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *AuthService) Logout(ctx context.Context, userID string) error {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	return s.userRepository.ClearRefreshToken(ctx, objectID)
}

func (s *AuthService) Me(ctx context.Context, userID string) (*domain.User, error) {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	return s.userRepository.FindByID(ctx, objectID)
}

func (s *AuthService) UpdateMe(ctx context.Context, userID string, input UpdateProfileInput) (*domain.User, error) {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	if err := s.userRepository.UpdateProfile(ctx, objectID, strings.TrimSpace(input.FirstName), strings.TrimSpace(input.LastName)); err != nil {
		return nil, err
	}

	return s.userRepository.FindByID(ctx, objectID)
}

func (s *AuthService) ChangePassword(ctx context.Context, userID string, input ChangePasswordInput) error {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	user, err := s.userRepository.FindByID(ctx, objectID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.CurrentPassword)); err != nil {
		return ErrInvalidCredentials
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.userRepository.UpdatePassword(ctx, objectID, string(passwordHash)); err != nil {
		return err
	}

	return s.userRepository.ClearRefreshToken(ctx, objectID)
}

func (s *AuthService) issueTokens(ctx context.Context, user *domain.User) (*AuthTokens, error) {
	accessToken, accessExpiresAt, err := s.tokenManager.GenerateAccessToken(user.ID.Hex(), user.Email, string(user.Role))
	if err != nil {
		return nil, err
	}

	refreshToken, refreshExpiresAt, err := s.tokenManager.GenerateRefreshToken(user.ID.Hex(), user.Email, string(user.Role))
	if err != nil {
		return nil, err
	}

	if err := s.userRepository.SaveRefreshToken(ctx, user.ID, s.tokenManager.HashToken(refreshToken), refreshExpiresAt); err != nil {
		return nil, err
	}

	return &AuthTokens{
		AccessToken:     accessToken,
		RefreshToken:    refreshToken,
		AccessTokenTTL:  accessExpiresAt,
		RefreshTokenTTL: refreshExpiresAt,
	}, nil
}

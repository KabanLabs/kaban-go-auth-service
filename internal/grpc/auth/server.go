package auth

import (
	"context"
	"errors"

	"github.com/VACdotCS/kaban-go-auth-service/internal/domain/models"
	"github.com/VACdotCS/kaban-go-auth-service/internal/services/auth"
	ssov1 "github.com/VACdotCS/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(ctx context.Context,
		email string,
		password string,
		appID int,
	) (tokens auth.TokenData, err error)
	RegisterNewUser(ctx context.Context,
		email string,
		password string,
	) (userID string, err error)
	GetJWK(ctx context.Context, kid string) (key *models.JWK, err error)
	ValidateAccessToken(ctx context.Context,
		accessToken string,
	) (isValid bool, err error)
	RegenerateRefreshToken(ctx context.Context,
		refreshToken string,
	) (accessToken string, err error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{
		auth: auth,
	})
}

const (
	emptyValue = 0
)

func (s *serverAPI) Login(
	ctx context.Context,
	req *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {
	if validateError := validateLogin(req); validateError != nil {
		return nil, validateError
	}

	// TODO: implement login via auth service
	tokens, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))

	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.NotFound, "Wrong email or password")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	req *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	if validateError := validateRegister(req); validateError != nil {
		return nil, validateError
	}

	userID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())

	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.RegisterResponse{
		UserId: userID,
	}, nil
}

func (s *serverAPI) ValidateToken(
	ctx context.Context,
	req *ssov1.ValidateTokenRequest,
) (*ssov1.ValidateTokenResponse, error) {
	//TODO: add validator for JWT string
	isValid, err := s.auth.ValidateAccessToken(ctx, req.GetAccessToken())

	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.ValidateTokenResponse{
		Valid: isValid,
	}, nil
}

func (s *serverAPI) GetJWKs(
	ctx context.Context,
	req *ssov1.GetJWKSRequest,
) (*ssov1.JWKS, error) {
	panic("implement me")
}

func (s *serverAPI) RegenerateRefreshToken(
	ctx context.Context,
	req *ssov1.RegenerateAccessTokenRequest,
) (*ssov1.RegenerateAccessTokenResponse, error) {
	//TODO: add validator for JWT string
	accessToken, err := s.auth.RegenerateRefreshToken(ctx, req.GetRefreshToken())

	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.RegenerateAccessTokenResponse{
		AccessToken: accessToken,
	}, nil
}

// Validation utils
func validateLogin(req *ssov1.LoginRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	if req.GetAppId() == emptyValue {
		return status.Error(codes.InvalidArgument, "app_id is required")
	}

	return nil
}

func validateRegister(req *ssov1.RegisterRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	return nil
}

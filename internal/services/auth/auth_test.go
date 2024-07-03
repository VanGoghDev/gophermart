package auth_test

import (
	"testing"
	"time"

	"github.com/VanGoghDev/gophermart/internal/services/auth"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

func TestGrantToken(t *testing.T) {
	type args struct {
		login       string
		secret      string
		tokenExpire time.Duration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid data",
			args: args{
				login:       "",
				secret:      "",
				tokenExpire: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid login",
			args: args{
				login:       "",
				secret:      "secret",
				tokenExpire: time.Second * 5,
			},
			wantErr: true,
		},
		{
			name: "invalid secret",
			args: args{
				login:       "test",
				secret:      "",
				tokenExpire: time.Second * 5,
			},
			wantErr: true,
		},
		{
			name: "invalid token expire",
			args: args{
				login:       "test",
				secret:      "secret",
				tokenExpire: 0,
			},
			wantErr: true,
		},
		{
			name: "valid token",
			args: args{
				login:       "test",
				secret:      "secret",
				tokenExpire: time.Second * 5,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTokenStr, err := auth.GrantToken(tt.args.login, tt.args.secret, tt.args.tokenExpire)
			if (err != nil) != tt.wantErr {
				t.Errorf("GrantToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				assert.Empty(t, err)
				ok, err := auth.IsAuthorized(gotTokenStr, tt.args.secret)
				assert.Empty(t, err)
				assert.True(t, ok)
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	type args struct {
		login       string
		secret      string
		tokenExpire time.Duration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid data",
			args: args{
				login:       "test",
				secret:      "secret",
				tokenExpire: time.Second * 5,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTokenStr, err := auth.GenerateToken(tt.args.login, tt.args.secret, tt.args.tokenExpire)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Empty(t, err)
				assert.NotEmpty(t, gotTokenStr)
			}
		})
	}
}

func TestIsAuthorized(t *testing.T) {
	type args struct {
		clientSecret string
		serverSecret string
		tokenExpired bool
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "valid secret",
			args: args{
				clientSecret: "secret",
				serverSecret: "secret",
				tokenExpired: false,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "secret differs",
			args: args{
				clientSecret: "secret2",
				serverSecret: "secret",
				tokenExpired: false,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "token expires",
			args: args{
				clientSecret: "secret",
				serverSecret: "secret",
				tokenExpired: true,
			},
			want:    false,
			wantErr: true,
		},
	}
	login := "test"
	tokenExpires := time.Hour * 3
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.tokenExpired {
				// создадим токен который почти тут же протухнет.
				tokenExpires = time.Microsecond
			}
			serverToken, err := auth.GenerateToken(login, tt.args.serverSecret, tokenExpires)
			assert.Empty(t, err)

			ok, err := auth.IsAuthorized(serverToken, tt.args.clientSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsAuthorized() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Empty(t, err)
				assert.True(t, ok)
			}
		})
	}
}

func TestIsAuthorizedTokenDifferAlg(t *testing.T) {
	tests := []struct {
		name       string
		anotherAlg bool
		want       bool
		wantErr    bool
	}{
		{
			name:       "different token alg",
			anotherAlg: true,
			want:       false,
			wantErr:    true,
		},
		{
			name:       "same token gen alg",
			anotherAlg: false,
			want:       true,
			wantErr:    false,
		},
	}
	login := "test"
	secret := "secret"
	tokenExpires := time.Second * 5
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tokenString string
			var err error
			if tt.anotherAlg {
				token := jwt.NewWithClaims(jwt.SigningMethodNone, nil)
				tokenString, _ = token.SignedString(jwt.UnsafeAllowNoneSignatureType)
			} else {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, auth.Claims{
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpires)),
					},
					UserLogin: login,
				})

				tokenString, _ = token.SignedString([]byte(secret))
			}
			ok, err := auth.IsAuthorized(tokenString, secret)

			if !tt.wantErr {
				assert.Empty(t, err)
			} else {
				assert.NotEmpty(t, err)
				assert.Equal(t, tt.want, ok)
			}
		})
	}
}

func TestExtractLoginFromToken(t *testing.T) {
	type args struct {
		login        string
		clientSecret string
		serverSecret string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "valid data",
			args: args{
				login:        "test",
				clientSecret: "secret",
				serverSecret: "secret",
			},
			want:    "test",
			wantErr: false,
		},
		{
			name: "secret differ",
			args: args{
				login:        "test",
				clientSecret: "secret2",
				serverSecret: "secret",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := auth.GenerateToken(tt.args.login, tt.args.serverSecret, time.Second*5)
			assert.Empty(t, err)
			assert.NotEmpty(t, token)

			gotLogin, err := auth.ExtractLoginFromToken(token, tt.args.clientSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractLoginFromToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Equal(t, tt.want, gotLogin)
			}
		})
	}
}

func TestExtractLoginFromTokenDifferAlg(t *testing.T) {
	tests := []struct {
		name       string
		login      string
		want       string
		anotherAlg bool
		wantErr    bool
	}{
		{
			name:       "different token alg",
			login:      "test",
			anotherAlg: true,
			want:       "",
			wantErr:    true,
		},
		{
			name:       "same token gen alg",
			anotherAlg: false,
			login:      "test",
			want:       "test",
			wantErr:    false,
		},
	}
	secret := "secret"
	tokenExpires := time.Second * 5
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tokenString string
			var err error
			if tt.anotherAlg {
				token := jwt.NewWithClaims(jwt.SigningMethodNone, nil)
				tokenString, _ = token.SignedString(jwt.UnsafeAllowNoneSignatureType)
			} else {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, auth.Claims{
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpires)),
					},
					UserLogin: tt.login,
				})

				tokenString, _ = token.SignedString([]byte(secret))
			}
			login, err := auth.ExtractLoginFromToken(tokenString, secret)

			if !tt.wantErr {
				assert.Empty(t, err)
				assert.Equal(t, login, tt.want)
			} else {
				assert.NotEmpty(t, err)
				assert.Equal(t, tt.want, login)
			}
		})
	}
}

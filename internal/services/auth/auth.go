package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserLogin string
}

func GrantToken(login string, secret string, tokenExpire time.Duration) (tokenStr string, err error) {
	if login == "" || secret == "" || tokenExpire == 0 {
		return "", fmt.Errorf("given parameters is not valid: %w", errors.New("invalid token data"))
	}

	tokenString, err := GenerateToken(login, secret, tokenExpire)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenString, nil
}

func GenerateToken(login string, secret string, tokenExpire time.Duration) (tokenStr string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpire)),
		},
		UserLogin: login,
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign string: %w", err)
	}
	return tokenString, nil
}

func IsAuthorized(token string, secret string) (bool, error) {
	_, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method :%v", token)
		}
		return []byte(secret), nil
	})

	if err != nil {
		return false, fmt.Errorf("failed to parse jwt token: %w", err)
	}

	return true, nil
}

func ExtractLoginFromToken(token string, secret string) (string, error) {
	tkn, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method :%v", token)
		}
		return []byte(secret), nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse jwt:%w ", err)
	}

	claims, ok := tkn.Claims.(jwt.MapClaims)

	if !ok && !tkn.Valid {
		return "", fmt.Errorf("invalid token :%w", err)
	}

	login, ok := claims["UserLogin"].(string)
	if !ok {
		return "", fmt.Errorf("failed to cast UserLogin claim to string:%w", err)
	}
	return login, nil
}

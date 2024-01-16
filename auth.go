package permissionbus

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type tokenClaims struct {
	Uid      string `json:"uid"`
	ExpireAt int64  `json:"expireAt"`
}

// 生成token
func GenerateToken(uid string, expireAt time.Time) (string, error) {
	if len(tokenSecretKey) == 0 {
		return "", errors.New("must call SetTokenSecretKey first.")
	}

	claims := tokenClaims{
		Uid:      uid,
		ExpireAt: expireAt.Unix(),
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	hasher := hmac.New(sha256.New, tokenSecretKey)
	hasher.Write(claimsJSON)
	signature := hasher.Sum(nil)

	token := base64.URLEncoding.EncodeToString(claimsJSON) + "." + base64.URLEncoding.EncodeToString(signature)
	return token, nil
}

// 解析token
func ParseToken(tokenString string) (string, error) {
	if len(tokenSecretKey) == 0 {
		return "", errors.New("must call SetTokenSecretKey first.")
	}

	parts := strings.Split(tokenString, ".")

	if len(parts) != 2 {
		return "", errors.New("非法的token格式")
	}

	claimsJSON, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", err
	}

	hasher := hmac.New(sha256.New, tokenSecretKey)
	hasher.Write(claimsJSON)
	expectedSignature := hasher.Sum(nil)

	actualSignature, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}

	if !hmac.Equal(expectedSignature, actualSignature) {
		return "", fmt.Errorf("非法的token签名")
	}

	var claims tokenClaims
	err = json.Unmarshal(claimsJSON, &claims)
	if err != nil {
		return "", err
	}

	nowStamp := time.Now().Unix()
	if nowStamp >= claims.ExpireAt {
		return "", errors.New("token过期，请重新登录")
	}

	return claims.Uid, nil
}

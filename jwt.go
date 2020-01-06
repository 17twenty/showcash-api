package showcash

import "github.com/dgrijalva/jwt-go"

var AuthSigningKey = "someTokenYoudWouldBeProud0f"

func ParseTokenWithDefaultSigningKey(token string, claims jwt.Claims) error {
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (i interface{}, e error) {
		return []byte(AuthSigningKey), nil
	})
	return err
}

func SignClaimsWithDefaultSigningKey(claims jwt.Claims) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString([]byte(AuthSigningKey))
}

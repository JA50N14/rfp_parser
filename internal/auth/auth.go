package auth

import (
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType string `json:"token_type"`
	ExpiresIn int `json:"expires_in"`
}

func GetGraphAccessToken(client *http.Client) (AccessTokenResponse, error) {
	jwt, err := makeJWT()
	if err != nil {
		return AccessTokenResponse{}, err
	}
	return fetchAccessToken(jwt, client)
}

func fetchAccessToken(jwt string, client *http.Client) (AccessTokenResponse, error) {
	tenant_id := os.Getenv("GRAPH_TENANT_ID")
	if tenant_id == "" {
		return AccessTokenResponse{}, errors.New("GRAPH_TENANT_ID .env variable not set")
	}

	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenant_id)
	
	clientID := os.Getenv("GRAPH_CLIENT_ID")
	if clientID == "" {
		return AccessTokenResponse{}, errors.New("GRAPH_CLIENT_ID .env variable not set")
	}

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("scope", "https://graph.microsoft.com/.default")
	data.Set("grant_type", "client_credentials")
	data.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	data.Set("client_assertion", jwt)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return AccessTokenResponse{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return AccessTokenResponse{}, fmt.Errorf("error making request to retrieve access token: %s", err)
	}
	defer resp.Body.Close()
	
	var accessTokenResp AccessTokenResponse

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&accessTokenResp)
	if err != nil {
		return AccessTokenResponse{}, err
	}

	if accessTokenResp.TokenType != "Bearer" {
		return AccessTokenResponse{}, fmt.Errorf("invalid token_type: %s", accessTokenResp.TokenType)
	}
	return accessTokenResp, nil
}


func makeJWT() (string, error) {
	clientID  := os.Getenv("GRAPH_CLIENT_ID")
	if clientID == "" {
		return "", errors.New("GRAPH_CLIENT_ID .env variable not set")
	}

	tenantID := os.Getenv("GRAPH_TENANT_ID")
	if tenantID == "" {
		return "", errors.New("GRAPH_TENANT_ID .env variable not set")
	}

	thumbprint, err := computeX5TFromCert()
	if err != nil {
		return "", err
	}

	privateKey, err := loadPrivateKey()
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims {
		"aud": "https://login.microsoftonline.com/" + tenantID + "/oauth2/v2.0/token",
		"iss": clientID,
		"sub": clientID,
		"jti": uuid.NewString(),
		"nbf": time.Now().UTC(),
		"exp": time.Now().UTC().Add(5 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	token.Header["alg"] = "RS256"
	token.Header["typ"] = "JWT"
	token.Header["x5t"] = thumbprint

	signedJWT, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	return signedJWT, nil
}

func loadPrivateKey() (*rsa.PrivateKey, error) {
	keyBytes, err := os.ReadFile(os.Getenv("PRIVATE_KEY_PATH"))
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, errors.New("invalid PEM file")
	}
	
	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("private key is not RSA")
		}
		return rsaKey, nil
	default:
		return nil, fmt.Errorf("unsupported key type %s", block.Type)
	}
}

func computeX5TFromCert() (string, error) {
	certPEM, err := os.ReadFile(os.Getenv("CERTIFICATE_PATH"))
	if err != nil {
		return "", err
	}

	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return "", fmt.Errorf("unsupported certificate PEM %s", block.Type)
	}
	
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}

	checkSum := sha1.Sum(cert.Raw)
	return base64.RawURLEncoding.EncodeToString(checkSum[:]), nil
}


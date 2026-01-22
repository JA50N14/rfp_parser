package auth

import (
	"context"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
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
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func GetGraphAccessToken(client *http.Client) (AccessTokenResponse, error) {
	tenantID := os.Getenv("GRAPH_TENANT_ID")
	if tenantID == "" {
		return AccessTokenResponse{}, fmt.Errorf("GRAPH_TENANT_ID .env variable not set")
	}

	clientID := os.Getenv("GRAPH_CLIENT_ID")
	if clientID == "" {
		return AccessTokenResponse{}, fmt.Errorf("GRAPH_CLIENT_ID .env variable not set")
	}

	jwt, err := makeJWT(tenantID, clientID)
	if err != nil {
		return AccessTokenResponse{}, fmt.Errorf("make JWT returned: %w", err)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	return fetchAccessToken(ctx, jwt, tenantID, clientID, client)
}

func fetchAccessToken(ctx context.Context, jwt, tenantID, clientID string, client *http.Client) (AccessTokenResponse, error) {
	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID)

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("scope", "https://graph.microsoft.com/.default")
	data.Set("grant_type", "client_credentials")
	data.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	data.Set("client_assertion", jwt)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return AccessTokenResponse{}, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return AccessTokenResponse{}, fmt.Errorf("sending request to Entra ID endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return AccessTokenResponse{}, fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var accessTokenResp AccessTokenResponse

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&accessTokenResp)
	if err != nil {
		return AccessTokenResponse{}, fmt.Errorf("decoding token response: %w", err)
	}

	if !strings.EqualFold(accessTokenResp.TokenType, "Bearer") {
		return AccessTokenResponse{}, fmt.Errorf("invalid token_type: %s", accessTokenResp.TokenType)
	}

////////////////
	fmt.Printf("ACCESS TOKEN: %s / ACCESS TOKEN TYPE: %s / Expires In: %d", accessTokenResp.AccessToken, accessTokenResp.TokenType, accessTokenResp.ExpiresIn)
////////////////
	return accessTokenResp, nil
}

func makeJWT(tenantID, clientID string) (string, error) {
	thumbprint, err := computeX5TFromCert()
	if err != nil {
		return "", fmt.Errorf("thumbprint returned: %w", err)
	}

	privateKey, err := loadPrivateKey()
	if err != nil {
		return "", fmt.Errorf("privatekey returned: %w", err)
	}

	claims := jwt.MapClaims{
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
		return "", fmt.Errorf("signed JWT returned: %w", err)
	}

	return signedJWT, nil
}

func loadPrivateKey() (*rsa.PrivateKey, error) {
	keyBytes, err := os.ReadFile(os.Getenv("PRIVATE_KEY_PATH"))
	if err != nil {
		return nil, fmt.Errorf("private key bytes returned: %w", err)
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM file: %w", err)
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("private key returned from PKCS8: %w", err)
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not RSA")
		}
		return rsaKey, nil
	default:
		return nil, fmt.Errorf("unsupported key type %s", block.Type)
	}
}

func computeX5TFromCert() (string, error) {
	certPEM, err := os.ReadFile(os.Getenv("CERTIFICATE_PATH"))
	if err != nil {
		return "", fmt.Errorf("could not read certificate from path: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return "", fmt.Errorf("unsupported certificate PEM %s", block.Type)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("cert returned from parse certificate: %w", err)
	}

	checkSum := sha1.Sum(cert.Raw)

	return base64.RawURLEncoding.EncodeToString(checkSum[:]), nil
}

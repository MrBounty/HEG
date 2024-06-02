package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/edgedb/edgedb-go"
	"github.com/gofiber/fiber/v2"
)

type DiscoveryDocument struct {
	UserInfoEndpoint string `json:"userinfo_endpoint"`
}

type UserProfile struct {
	Email        string `json:"email"`
	Name         string `json:"name"`
	AvatarGitHub string `json:"avatar_url"`
	AvatarGoogle string `json:"picture"`
}

type TokenResponse struct {
	AuthToken     string `json:"auth_token"`
	IdentityID    string `json:"identity_id"`
	ProviderToken string `json:"provider_token"`
}

func getGoogleUserProfile(providerToken string) (string, string, string) {
	// Fetch the discovery document
	resp, err := http.Get("https://accounts.google.com/.well-known/openid-configuration")
	if err != nil {
		fmt.Println("Error fetching discovery document")
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error fetching discovery document")
		panic(resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading discovery document")
		panic(err)
	}

	var discoveryDocument DiscoveryDocument
	if err := json.Unmarshal(body, &discoveryDocument); err != nil {
		fmt.Println("Error unmarshalling discovery document")
		panic(err)
	}

	// Fetch the user profile
	req, err := http.NewRequest("GET", discoveryDocument.UserInfoEndpoint, nil)
	if err != nil {
		fmt.Println("Error fetching user profile")
		panic(err)
	}
	req.Header.Set("Authorization", "Bearer "+providerToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		fmt.Println("Error fetching user profile")
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic("Error fetching user profile")
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading user profile")
		panic(err)
	}

	var userProfile UserProfile
	if err := json.Unmarshal(body, &userProfile); err != nil {
		fmt.Println("Error unmarshalling user profile")
		panic(err)
	}

	return userProfile.Email, userProfile.Name, userProfile.AvatarGoogle
}

func getGitHubUserProfile(providerToken string) (string, string, string) {
	// Create the request to fetch the user profile
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		fmt.Println("failed to create request: user profile")
		panic(err)
	}
	req.Header.Set("Authorization", "Bearer "+providerToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("failed to execute request")
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("failed to fetch user profile: status code")
		panic(resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("failed to read response body")
		panic(err)
	}

	var userProfile UserProfile
	if err := json.Unmarshal(body, &userProfile); err != nil {
		fmt.Println("failed to unmarshal user profile")
		panic(err)
	}

	return userProfile.Email, userProfile.Name, userProfile.AvatarGitHub
}

func generatePKCE() (string, string) {
	verifier_source := make([]byte, 32)
	_, err := rand.Read(verifier_source)
	if err != nil {
		fmt.Println("failed to generate PKCE")
		panic(err)
	}

	verifier := base64.RawURLEncoding.EncodeToString(verifier_source)
	challenge := sha256.Sum256([]byte(verifier))
	return verifier, base64.RawURLEncoding.EncodeToString(challenge[:])
}

func handleUiSignIn(c *fiber.Ctx) error {
	verifier, challenge := generatePKCE()

	c.Cookie(&fiber.Cookie{
		Name:     "my-cookie-name-verifier",
		Value:    verifier,
		HTTPOnly: true,
		Path:     "/",
		Secure:   true,
	})

	return c.Redirect(fmt.Sprintf("%s/ui/signup?challenge=%s", os.Getenv("EDGEDB_AUTH_BASE_URL"), challenge), fiber.StatusTemporaryRedirect)
}

func handleCallbackSignup(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		err := c.Query("error")
		fmt.Println("OAuth callback is missing 'code'. OAuth provider responded with error")
		panic(err)
	}

	verifier := c.Cookies("my-cookie-name-verifier", "")
	if verifier == "" {
		panic("Could not find 'verifier' in the cookie store. Is this the same user agent/browser that started the authorization flow?")
	}

	codeExchangeURL := fmt.Sprintf("%s/token?code=%s&verifier=%s", os.Getenv("EDGEDB_AUTH_BASE_URL"), code, verifier)
	resp, err := http.Get(codeExchangeURL)
	if err != nil {
		fmt.Println("Error exchanging code for access token")
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("Error exchanging code for access token")
		panic(string(body))
	}

	var tokenResponse TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		fmt.Println("Error decoding auth server response")
		panic(err)
	}

	c.Cookie(&fiber.Cookie{
		Name:     "my-cookie-name-auth-token",
		Value:    tokenResponse.AuthToken,
		HTTPOnly: true,
		Path:     "/",
		Secure:   true,
	})

	// Get the issuer of the identity
	var identity Identity
	identityUUID, err := edgedb.ParseUUID(tokenResponse.IdentityID)
	if err != nil {
		fmt.Println("Error parsing UUID")
		panic(err)
	}
	err = edgeGlobalClient.WithGlobals(map[string]interface{}{"ext::auth::client_token": c.Cookies("jade-edgedb-auth-token")}).QuerySingle(edgeCtx, `
		SELECT ext::auth::Identity {
			issuer
		} FILTER .id = <uuid>$0
	`, &identity, identityUUID)
	if err != nil {
		fmt.Println("Error fetching identity")
		panic(err)
	}

	var (
		providerEmail  string
		providerName   string
		providerAvatar string
	)

	// Get the email and name from the provider
	if identity.Issuer == "https://accounts.google.com" {
		providerEmail, providerName, providerAvatar = getGoogleUserProfile(tokenResponse.ProviderToken)
	} else if identity.Issuer == "https://github.com" {
		providerEmail, providerName, providerAvatar = getGitHubUserProfile(tokenResponse.ProviderToken)
	}

	// Here you handle User creation. I put this as an example
	err = edgeGlobalClient.WithGlobals(map[string]interface{}{"ext::auth::client_token": tokenResponse.AuthToken}).Execute(edgeCtx, `
	INSERT User {
		email := <str>$0,
		name := <str>$1,
		avatar := <str>$2,
		identity := (SELECT ext::auth::Identity FILTER .id = <uuid>$3)
	  }
	`, providerEmail, providerName, providerAvatar, identityUUID)
	if err != nil {
		fmt.Println("Error creating user")
		panic(err)
	}

	return c.Redirect("/", fiber.StatusPermanentRedirect)
}

func handleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		err := c.Query("error")
		fmt.Println("OAuth callback is missing 'code'. OAuth provider responded with error")
		panic(err)
	}

	verifier := c.Cookies("my-cookie-name-verifier", "")
	if verifier == "" {
		panic("Could not find 'verifier' in the cookie store. Is this the same user agent/browser that started the authorization flow?")
	}

	codeExchangeURL := fmt.Sprintf("%s/token?code=%s&verifier=%s", os.Getenv("EDGEDB_AUTH_BASE_URL"), code, verifier)
	resp, err := http.Get(codeExchangeURL)
	if err != nil {
		fmt.Println("Error exchanging code for access token")
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("Error exchanging code for access token")
		panic(string(body))
	}

	var tokenResponse TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		fmt.Println("Error decoding auth server response")
		panic(err)
	}

	c.Cookie(&fiber.Cookie{
		Name:     "my-cookie-name-auth-token",
		Value:    tokenResponse.AuthToken,
		HTTPOnly: true,
		Path:     "/",
		Secure:   true,
		SameSite: "Strict",
	})

	return c.Redirect("/", fiber.StatusPermanentRedirect)
}

func handleSignOut(c *fiber.Ctx) error {
	c.ClearCookie("my-cookie-name-auth-token")
	return c.Redirect("/", fiber.StatusTemporaryRedirect)
}

package security

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Nerzal/gocloak/v13"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt"
	"github.com/javiorfo/go-microservice-lib/response"
	"github.com/javiorfo/go-microservice-lib/response/codes"
	"github.com/javiorfo/go-microservice-lib/tracing"
)

type Securizer interface {
	SecureWithRoles(roles ...string) fiber.Handler
}

type KeycloakConfig struct {
	Keycloak       *gocloak.GoCloak
	Realm          string
	ClientID       string
	ClientSecret   string
	AdminRealmUser *KeycloakAdminRealmUser
	Enabled        bool
}

type KeycloakAdminRealmUser struct {
	Username string
	Password string
}

type SimpleUser struct {
	Username string
	Password string
	FistName string
	LastName string
	Email    string
}

func (kc KeycloakConfig) SecureWithRoles(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log.Infof("%s Keycloak capture: %s", tracing.LogTraceAndSpan(c), c.Path())
		if !kc.Enabled {
			return c.Next()
		}

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			authorizationHeaderError := response.NewRestResponseError(c, response.ResponseError{
				Code:    codes.AUTH_ERROR,
				Message: "Authorization header missing",
			})
			return c.Status(http.StatusUnauthorized).JSON(authorizationHeaderError)
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		rptResult, err := kc.Keycloak.RetrospectToken(c.Context(), token, kc.ClientID, kc.ClientSecret, kc.Realm)
		if err != nil || !*rptResult.Active {
			invalidTokenError := response.NewRestResponseError(c, response.ResponseError{
				Code:    codes.AUTH_ERROR,
				Message: "Invalid or expired token",
			})
			return c.Status(http.StatusUnauthorized).JSON(invalidTokenError)
		}

		if user, err := userHasRole(kc.ClientID, token, roles); err != nil {
			unauthorizedRoleError := response.NewRestResponseError(c, response.ResponseError{
				Code:    codes.AUTH_ERROR,
				Message: err.Error(),
			})
			return c.Status(http.StatusUnauthorized).JSON(unauthorizedRoleError)
		} else {
			c.Locals("tokenUser", *user)
			return c.Next()
		}
	}
}

func (kc KeycloakConfig) CreateUser(c *fiber.Ctx, simpleUser SimpleUser) error {
	log.Infof("%s Creating user: %s", tracing.LogTraceAndSpan(c), simpleUser.Username)

	if kc.AdminRealmUser == nil {
		return errors.New("AdminRealmUser is not set")
	}

	token, err := kc.Keycloak.LoginAdmin(c.Context(), kc.AdminRealmUser.Username, kc.AdminRealmUser.Password, kc.Realm)
	if err != nil {
		return fmt.Errorf("Error logging Admin Realm User: %v\n", err)
	}

	user := gocloak.User{
		Username:  gocloak.StringP(simpleUser.Username),
		FirstName: gocloak.StringP(simpleUser.FistName),
		LastName:  gocloak.StringP(simpleUser.LastName),
		Email:     gocloak.StringP(simpleUser.Email),
		Enabled:   gocloak.BoolP(true),
	}

	createdUserID, err := kc.Keycloak.CreateUser(c.Context(), token.AccessToken, kc.Realm, user)
	if err != nil {
		return fmt.Errorf("Error creating user: %v\n", err)
	}

	err = kc.Keycloak.SetPassword(c.Context(), token.AccessToken, createdUserID, kc.Realm, simpleUser.Password, true)
	if err != nil {
		return fmt.Errorf("Error setting password: %v\n", err)
	}

	log.Infof("%s User created and password set successfully. Keycloak UserID %s", tracing.LogTraceAndSpan(c), createdUserID)
	return nil
}

type customClaims struct {
	ResourceAccess    map[string]any `json:"resource_access"`
	PreferredUsername string         `json:"preferred_username"`
	Aud               []string       `json:"aud"`
	jwt.StandardClaims
}

func userHasRole(clientID, tokenStr string, roles []string) (*string, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, &customClaims{})
	if err != nil {
		return nil, fmt.Errorf("Error parsing token: %v", err)
	}

	claims, ok := token.Claims.(*customClaims)
	if !ok {
		log.Errorf("Error asserting claims")
		return nil, fmt.Errorf("Error asserting claims")
	}

	resourceData, ok := claims.ResourceAccess[clientID]
	if !ok {
		return nil, fmt.Errorf("Resource key %s not found", clientID)
	}

	resourceMap := resourceData.(map[string]any)
	clientRoles := resourceMap["roles"].([]any)
	if len(clientRoles) > 0 {
		for _, cr := range clientRoles {
			for _, r := range roles {
				if r == cr.(string) {
					return &claims.PreferredUsername, nil
				}
			}
		}
		return nil, fmt.Errorf("User does not have permission to access")
	}

	return nil, fmt.Errorf("No roles found for resource key %s", clientID)
}

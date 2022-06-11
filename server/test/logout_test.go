package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/authorizerdev/authorizer/server/constants"
	"github.com/authorizerdev/authorizer/server/db"
	"github.com/authorizerdev/authorizer/server/graph/model"
	"github.com/authorizerdev/authorizer/server/memorystore"
	"github.com/authorizerdev/authorizer/server/resolvers"
	"github.com/authorizerdev/authorizer/server/token"
	"github.com/stretchr/testify/assert"
)

func logoutTests(t *testing.T, s TestSetup) {
	t.Helper()
	t.Run(`should logout user`, func(t *testing.T) {
		req, ctx := createContext(s)
		email := "logout." + s.TestInfo.Email

		_, err := resolvers.MagicLinkLoginResolver(ctx, model.MagicLinkLoginInput{
			Email: email,
		})

		verificationRequest, err := db.Provider.GetVerificationRequestByEmail(email, constants.VerificationTypeMagicLinkLogin)
		verifyRes, err := resolvers.VerifyEmailResolver(ctx, model.VerifyEmailInput{
			Token: verificationRequest.Token,
		})

		accessToken := *verifyRes.AccessToken
		assert.NotEmpty(t, accessToken)

		claims, err := token.ParseJWTToken(accessToken)
		assert.NoError(t, err)
		assert.NotEmpty(t, claims)

		sessionToken, err := memorystore.Provider.GetUserSession(verifyRes.User.ID, constants.TokenTypeSessionToken+"_"+claims["nonce"].(string))
		assert.NoError(t, err)
		assert.NotEmpty(t, sessionToken)

		cookie := fmt.Sprintf("%s=%s;", constants.AppCookieName+"_session", sessionToken)
		cookie = strings.TrimSuffix(cookie, ";")

		req.Header.Set("Cookie", cookie)
		_, err = resolvers.LogoutResolver(ctx)
		assert.Nil(t, err)
		cleanData(email)
	})
}

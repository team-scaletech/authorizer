package resolvers

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/authorizerdev/authorizer/server/cookie"
	"github.com/authorizerdev/authorizer/server/crypto"
	"github.com/authorizerdev/authorizer/server/db"
	"github.com/authorizerdev/authorizer/server/db/models"
	"github.com/authorizerdev/authorizer/server/graph/model"
	"github.com/authorizerdev/authorizer/server/memorystore"
	"github.com/authorizerdev/authorizer/server/memorystore/providers/redis"
	"github.com/authorizerdev/authorizer/server/refs"
	"github.com/authorizerdev/authorizer/server/utils"

	log "github.com/sirupsen/logrus"
)

// TotpInfoResolver retrieves Time-based One-Time Password (TOTP) information for oauth authentication.
func TotpInfoResolver(ctx context.Context) (*model.AuthResponse, error) {
	// Declare a variable to hold the response
	var res *model.AuthResponse

	// Get the GinContext from the context
	gc, err := utils.GinContextFromContext(ctx)
	if err != nil {
		log.Debug("Failed to get GinContext: ", err)
		return res, err
	}

	// Retrieve Multi-Factor Authentication (MFA) session from cookie
	mfaSession, err := cookie.GetMfaSession(gc)
	if err != nil {
		log.Debug("Failed to get oauth mfa session: ", err)
		return res, fmt.Errorf(`invalid session: %s`, err.Error())
	}

	// Retrieve user ID from OAuth MFA session cookie
	userId, err := cookie.GetOAuthMfaSession(gc)
	if err != nil {
		log.Debug("Failed to get oauth mfa session: ", err)
		return res, fmt.Errorf(`invalid session: %s`, err.Error())
	}

	// Get user information by user ID
	var user *models.User
	if userId != "" {
		user, err = db.Provider.GetUserByID(ctx, userId)
		if err != nil {
			log.Debug("Failed to get user by id: ", err)
			return res, fmt.Errorf(`failed to get user by id: %s`, err.Error())
		}
	}

	// If user is not found or an error occurs, return an error
	if user == nil || err != nil {
		return res, fmt.Errorf(`user not found`)
	}

	// Retrieve encrypted TOTP information from the MFA session
	encryptedTotpInfo, err := memorystore.Provider.GetMfaSession(userId, mfaSession, redis.MfaOAuthSessionPrefix)
	if err != nil {
		log.Debug("Failed to get mfa session: ", err)
		return res, fmt.Errorf(`invalid session: %s`, err.Error())
	}

	// Decrypt the TOTP info from Base64
	totpInfoString, err := crypto.DecryptB64(encryptedTotpInfo)
	if err != nil {
		log.Debug("Error while decrypting env data from B64: ", err)
		return res, err
	}

	// Parse the TOTP info string into a map
	queryValues, err := url.ParseQuery(totpInfoString)
	if err != nil {
		fmt.Println("Error parsing totpInfo:", err)
		return nil, err
	}

	// Retrieve the image URL from the query parameters
	image := queryValues.Get("authenticator_scanner_image")

	// Replace spaces with plus signs in the image URL
	scannerImage := strings.Replace(image, " ", "+", -1)

	// Create the AuthResponse object with parsed TOTP info
	res = &model.AuthResponse{
		Message:                    `Proceed to totp flow`,
		ShouldShowTotpScreen:       utils.ParseBool(queryValues.Get("should_show_totp_screen")),
		AuthenticatorScannerImage:  refs.NewStringRef(scannerImage),
		AuthenticatorSecret:        refs.NewStringRef(queryValues.Get("authenticator_secret")),
		AuthenticatorRecoveryCodes: utils.ParseStringArray(queryValues.Get("authenticator_recovery_codes")),
		User:                       user.AsAPIUser(),
	}

	// Respond with TOTP info
	return res, nil
}

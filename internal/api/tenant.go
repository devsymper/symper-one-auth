package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/supabase/auth/internal/api/apierrors"
	mail "github.com/supabase/auth/internal/mailer"
	"github.com/supabase/auth/internal/models"
	"github.com/supabase/auth/internal/storage"
)

// TenantInviteParams are the parameters for inviting a user to a tenant
type TenantInviteParams struct {
	Email    string `json:"email"`
	Role     string `json:"role"`                // admin, member
	TenantID string `json:"tenant_id,omitempty"` // optional, will use user's default tenant if not provided
}

// TenantInviteResponse represents the response after creating an invitation
type TenantInviteResponse struct {
	InvitationID string `json:"invitation_id"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	TenantID     string `json:"tenant_id"`
	ExpiresAt    string `json:"expires_at"`
	Token        string `json:"token"` // Include token for testing/development
}

// TenantInviteAcceptParams are the parameters for accepting a tenant invitation
type TenantInviteAcceptParams struct {
	Token string `json:"token"`
}

// TenantInviteAcceptResponse represents the response after accepting an invitation
type TenantInviteAcceptResponse struct {
	Tenant *models.Tenant       `json:"tenant"`
	Member *models.TenantMember `json:"member"`
}

// InviteToTenant invites a user to join a tenant
func (a *API) InviteToTenant(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := a.db.WithContext(ctx)
	user := getUser(ctx)

	if user == nil {
		return apierrors.NewForbiddenError(apierrors.ErrorCodeNoAuthorization, "Authorization required")
	}

	params := &TenantInviteParams{}
	if err := retrieveRequestParams(r, params); err != nil {
		return err
	}

	// Validate parameters
	if params.Email == "" {
		return apierrors.NewBadRequestError(apierrors.ErrorCodeValidationFailed, "Email is required")
	}

	var err error
	params.Email, err = a.validateEmail(params.Email)
	if err != nil {
		return err
	}

	// Validate role
	if params.Role == "" {
		params.Role = "member"
	}
	if params.Role != "admin" && params.Role != "member" {
		return apierrors.NewBadRequestError(apierrors.ErrorCodeValidationFailed, "Role must be 'admin' or 'member'")
	}

	var tenant *models.Tenant
	if params.TenantID != "" {
		// Use specified tenant
		tenantID, err := uuid.FromString(params.TenantID)
		if err != nil {
			return apierrors.NewBadRequestError(apierrors.ErrorCodeValidationFailed, "Invalid tenant ID")
		}
		tenant, err = models.FindTenantByID(db, tenantID)
		if err != nil {
			if models.IsNotFoundError(err) {
				return apierrors.NewNotFoundError(apierrors.ErrorCodeValidationFailed, "Tenant not found")
			}
			return apierrors.NewInternalServerError("Database error finding tenant").WithInternalError(err)
		}
	} else {
		// Use user's default tenant
		var err error
		tenant, err = models.GetDefaultTenantForUser(db, user.ID)
		if err != nil {
			if models.IsNotFoundError(err) {
				return apierrors.NewNotFoundError(apierrors.ErrorCodeValidationFailed, "No tenant found for user")
			}
			return apierrors.NewInternalServerError("Database error finding tenant").WithInternalError(err)
		}
	}

	// Check if user has permission to invite to this tenant
	if tenant.OwnerID != user.ID {
		// Check if user is admin of this tenant
		member, err := models.FindTenantMemberByUserAndTenant(db, user.ID, tenant.ID)
		if err != nil || member.Role != "admin" {
			return apierrors.NewForbiddenError(apierrors.ErrorCodeNotAdmin, "Insufficient permissions to invite users to this tenant")
		}
	}

	// Check if invited user already exists
	aud := a.requestAud(ctx, r)
	invitedUser, err := models.FindUserByEmailAndAudience(db, params.Email, aud)
	if err != nil && !models.IsNotFoundError(err) {
		return apierrors.NewInternalServerError("Database error finding invited user").WithInternalError(err)
	}

	// If user exists, check if they're already a member of this tenant
	if invitedUser != nil {
		existingMember, err := models.FindTenantMemberByUserAndTenant(db, invitedUser.ID, tenant.ID)
		if err == nil && existingMember != nil {
			return apierrors.NewUnprocessableEntityError(apierrors.ErrorCodeValidationFailed, "User is already a member of this tenant")
		}
	}

	var invitation *models.TenantInvitation
	err = db.Transaction(func(tx *storage.Connection) error {
		// Check for existing pending invitation
		existingInvite := &models.TenantInvitation{}
		err := tx.Where("tenant_id = ? AND email = ? AND accepted_at IS NULL AND expires_at > NOW() AND deleted_at IS NULL",
			tenant.ID, params.Email).First(existingInvite)
		if err == nil {
			return apierrors.NewUnprocessableEntityError(apierrors.ErrorCodeValidationFailed, "Pending invitation already exists for this email")
		}

		// Generate secure token
		token, err := generateInvitationToken()
		if err != nil {
			return apierrors.NewInternalServerError("Failed to generate invitation token").WithInternalError(err)
		}

		// Create invitation with 7 days expiry
		expiresAt := time.Now().Add(7 * 24 * time.Hour)
		invitation = models.NewTenantInvitation(tenant.ID, params.Email, params.Role, user.ID, token, expiresAt)

		if terr := tx.Create(invitation); terr != nil {
			return apierrors.NewInternalServerError("Database error creating invitation").WithInternalError(terr)
		}

		// Send tenant invitation email
		if terr := a.sendTenantInvite(r, tx, invitation, tenant, user); terr != nil {
			return terr
		}

		return nil
	})

	if err != nil {
		return err
	}

	response := &TenantInviteResponse{
		InvitationID: invitation.ID.String(),
		Email:        invitation.Email,
		Role:         invitation.Role,
		TenantID:     invitation.TenantID.String(),
		ExpiresAt:    invitation.ExpiresAt.Format(time.RFC3339),
		Token:        invitation.Token, // Include for testing
	}

	return sendJSON(w, http.StatusOK, response)
}

// AcceptTenantInvitation accepts a tenant invitation
func (a *API) AcceptTenantInvitation(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := a.db.WithContext(ctx)
	user := getUser(ctx)

	if user == nil {
		return apierrors.NewForbiddenError(apierrors.ErrorCodeNoAuthorization, "Authorization required")
	}

	params := &TenantInviteAcceptParams{}
	if err := retrieveRequestParams(r, params); err != nil {
		return err
	}

	if params.Token == "" {
		return apierrors.NewBadRequestError(apierrors.ErrorCodeValidationFailed, "Token is required")
	}

	var tenant *models.Tenant
	var member *models.TenantMember

	err := db.Transaction(func(tx *storage.Connection) error {
		// Find invitation by token
		invitation, err := models.FindTenantInvitationByToken(tx, params.Token)
		if err != nil {
			if models.IsNotFoundError(err) {
				return apierrors.NewNotFoundError(apierrors.ErrorCodeInviteNotFound, "Invalid or expired invitation token")
			}
			return apierrors.NewInternalServerError("Database error finding invitation").WithInternalError(err)
		}

		// Check if invitation is expired
		if invitation.IsExpired() {
			return apierrors.NewUnprocessableEntityError(apierrors.ErrorCodeValidationFailed, "Invitation has expired")
		}

		// Check if invitation is already accepted
		if invitation.IsAccepted() {
			return apierrors.NewUnprocessableEntityError(apierrors.ErrorCodeValidationFailed, "Invitation has already been accepted")
		}

		// Check if email matches current user
		if !strings.EqualFold(invitation.Email, user.GetEmail()) {
			return apierrors.NewForbiddenError(apierrors.ErrorCodeEmailAddressNotAuthorized, "Invitation email does not match current user")
		}

		// Get tenant
		tenant, err = models.FindTenantByID(tx, invitation.TenantID)
		if err != nil {
			return apierrors.NewInternalServerError("Database error finding tenant").WithInternalError(err)
		}

		// Check if user is already a member
		existingMember, err := models.FindTenantMemberByUserAndTenant(tx, user.ID, tenant.ID)
		if err == nil && existingMember != nil {
			return apierrors.NewUnprocessableEntityError(apierrors.ErrorCodeValidationFailed, "User is already a member of this tenant")
		}

		// Create tenant member
		member = models.NewTenantMember(tenant.ID, user.ID, invitation.Role)
		if terr := tx.Create(member); terr != nil {
			return apierrors.NewInternalServerError("Database error creating tenant member").WithInternalError(terr)
		}

		// Update invitation as accepted
		now := time.Now()
		invitation.AcceptedAt = &now
		invitation.AcceptedByID = &user.ID
		if terr := tx.Update(invitation); terr != nil {
			return apierrors.NewInternalServerError("Database error updating invitation").WithInternalError(terr)
		}

		return nil
	})

	if err != nil {
		return err
	}

	response := &TenantInviteAcceptResponse{
		Tenant: tenant,
		Member: member,
	}

	return sendJSON(w, http.StatusOK, response)
}

// generateInvitationToken generates a secure random token for invitations
func generateInvitationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetUserTenants returns all tenants the user belongs to
func (a *API) GetUserTenants(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := a.db.WithContext(ctx)
	user := getUser(ctx)

	if user == nil {
		return apierrors.NewForbiddenError(apierrors.ErrorCodeNoAuthorization, "Authorization required")
	}

	tenants, err := models.GetUserTenants(db, user.ID)
	if err != nil {
		return apierrors.NewInternalServerError("Database error getting user tenants").WithInternalError(err)
	}

	return sendJSON(w, http.StatusOK, map[string]interface{}{
		"tenants": tenants,
	})
}

// sendTenantInvite sends a tenant invitation email to the invited user
func (a *API) sendTenantInvite(r *http.Request, tx *storage.Connection, invitation *models.TenantInvitation, tenant *models.Tenant, inviter *models.User) error {
	config := a.config

	// Create a temporary user-like object for the email system
	// Since the invited user might not exist yet, we create a minimal user object
	tempUser := &models.User{
		ID:    invitation.ID, // Use invitation ID as temporary user ID
		Email: storage.NullString(invitation.Email),
	}

	// Generate OTP for email display (we'll use our custom token for actual verification)
	otpLength := config.Mailer.OtpLength
	otp := invitation.Token[:min(otpLength, len(invitation.Token))]

	// Set confirmation token to our invitation token for email templates
	tempUser.ConfirmationToken = invitation.Token

	// Use the existing sendEmail infrastructure with InviteVerification type
	if err := a.sendEmail(r, tx, tempUser, mail.InviteVerification, otp, "", invitation.Token); err != nil {
		return apierrors.NewInternalServerError("Error sending tenant invitation email").WithInternalError(err)
	}

	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/supabase/auth/internal/conf"
	"github.com/supabase/auth/internal/models"
)

type TenantTestSuite struct {
	suite.Suite
	API    *API
	Config *conf.GlobalConfiguration

	// Test users and tenants
	instanceID uuid.UUID
	testUser   *models.User
	testTenant *models.Tenant
}

func TestTenant(t *testing.T) {
	api, config, err := setupAPIForTest()
	require.NoError(t, err)

	ts := &TenantTestSuite{
		API:        api,
		Config:     config,
		instanceID: uuid.Must(uuid.NewV4()),
	}
	defer api.db.Close()

	suite.Run(t, ts)
}

func (ts *TenantTestSuite) SetupTest() {
	require := require.New(ts.T())

	// Create test user
	user, err := models.NewUser("", "test@example.com", "password", ts.instanceID.String(), nil)
	require.NoError(err)
	require.NoError(ts.API.db.Create(user))
	ts.testUser = user

	// Create test tenant
	tenant := models.NewTenant("Test Tenant", "A test tenant", user.ID)
	require.NoError(ts.API.db.Create(tenant))
	ts.testTenant = tenant

	// Create tenant member relationship
	member := models.NewTenantMember(tenant.ID, user.ID, "owner")
	require.NoError(ts.API.db.Create(member))
}

func (ts *TenantTestSuite) TearDownTest() {
	// Clean up test data
	ts.API.db.RawQuery("DELETE FROM tenant_invitations").Exec()
	ts.API.db.RawQuery("DELETE FROM tenant_members").Exec()
	ts.API.db.RawQuery("DELETE FROM tenants").Exec()
	ts.API.db.RawQuery("DELETE FROM users").Exec()
}

func (ts *TenantTestSuite) TestInviteToTenant() {
	require := require.New(ts.T())

	// Test successful invitation
	reqBody := TenantInviteParams{
		Email: "invite@example.com",
		Role:  "member",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/tenants/invite", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Add authenticated user to context
	token := ts.generateAccessToken(ts.testUser)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)

	require.Equal(http.StatusOK, w.Code)

	var response TenantInviteResponse
	require.NoError(json.NewDecoder(w.Body).Decode(&response))
	require.Equal("invite@example.com", response.Email)
	require.Equal("member", response.Role)
	require.Equal(ts.testTenant.ID.String(), response.TenantID)
	require.NotEmpty(response.Token)

	// Verify invitation was created in database
	invitation, err := models.FindTenantInvitationByToken(ts.API.db, response.Token)
	require.NoError(err)
	require.Equal("invite@example.com", invitation.Email)
	require.Equal("member", invitation.Role)
	require.Equal(ts.testTenant.ID, invitation.TenantID)
	require.Equal(ts.testUser.ID, invitation.InviterID)
}

func (ts *TenantTestSuite) TestInviteToTenantInvalidEmail() {
	require := require.New(ts.T())

	reqBody := TenantInviteParams{
		Email: "invalid-email",
		Role:  "member",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/tenants/invite", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	token := ts.generateAccessToken(ts.testUser)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)

	require.Equal(http.StatusBadRequest, w.Code)
}

func (ts *TenantTestSuite) TestInviteToTenantInvalidRole() {
	require := require.New(ts.T())

	reqBody := TenantInviteParams{
		Email: "test@example.com",
		Role:  "invalid-role",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/tenants/invite", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	token := ts.generateAccessToken(ts.testUser)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)

	require.Equal(http.StatusBadRequest, w.Code)
}

func (ts *TenantTestSuite) TestInviteToTenantUnauthorized() {
	require := require.New(ts.T())

	reqBody := TenantInviteParams{
		Email: "test@example.com",
		Role:  "member",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/tenants/invite", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)

	require.Equal(http.StatusUnauthorized, w.Code)
}

func (ts *TenantTestSuite) TestAcceptTenantInvitation() {
	require := require.New(ts.T())

	// Create another user to accept the invitation
	inviteeUser, err := models.NewUser("", "invitee@example.com", "password", ts.instanceID.String(), nil)
	require.NoError(err)
	require.NoError(ts.API.db.Create(inviteeUser))

	// Create invitation
	token := "test-invitation-token"
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	invitation := models.NewTenantInvitation(ts.testTenant.ID, "invitee@example.com", "member", ts.testUser.ID, token, expiresAt)
	require.NoError(ts.API.db.Create(invitation))

	// Test accepting invitation
	reqBody := TenantInviteAcceptParams{
		Token: token,
	}

	body, err := json.Marshal(reqBody)
	require.NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/tenants/accept-invitation", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Add invitee user to context
	inviteeToken := ts.generateAccessToken(inviteeUser)
	req.Header.Set("Authorization", "Bearer "+inviteeToken)

	w := httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)

	require.Equal(http.StatusOK, w.Code)

	var response TenantInviteAcceptResponse
	require.NoError(json.NewDecoder(w.Body).Decode(&response))
	require.NotNil(response.Tenant)
	require.NotNil(response.Member)
	require.Equal(ts.testTenant.ID, response.Tenant.ID)
	require.Equal(inviteeUser.ID, response.Member.UserID)
	require.Equal("member", response.Member.Role)

	// Verify invitation was marked as accepted
	require.NoError(ts.API.db.Reload(invitation))
	require.NotNil(invitation.AcceptedAt)
	require.Equal(&inviteeUser.ID, invitation.AcceptedByID)

	// Verify tenant member was created
	member, err := models.FindTenantMemberByUserAndTenant(ts.API.db, inviteeUser.ID, ts.testTenant.ID)
	require.NoError(err)
	require.Equal("member", member.Role)
}

func (ts *TenantTestSuite) TestAcceptTenantInvitationInvalidToken() {
	require := require.New(ts.T())

	reqBody := TenantInviteAcceptParams{
		Token: "invalid-token",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/tenants/accept-invitation", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	token := ts.generateAccessToken(ts.testUser)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)

	require.Equal(http.StatusNotFound, w.Code)
}

func (ts *TenantTestSuite) TestAcceptTenantInvitationExpired() {
	require := require.New(ts.T())

	// Create expired invitation
	token := "expired-invitation-token"
	expiresAt := time.Now().Add(-1 * time.Hour) // Expired
	invitation := models.NewTenantInvitation(ts.testTenant.ID, ts.testUser.Email.String(), "member", ts.testUser.ID, token, expiresAt)
	require.NoError(ts.API.db.Create(invitation))

	reqBody := TenantInviteAcceptParams{
		Token: token,
	}

	body, err := json.Marshal(reqBody)
	require.NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/tenants/accept-invitation", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	userToken := ts.generateAccessToken(ts.testUser)
	req.Header.Set("Authorization", "Bearer "+userToken)

	w := httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)

	require.Equal(http.StatusUnprocessableEntity, w.Code)
}

func (ts *TenantTestSuite) TestGetUserTenants() {
	require := require.New(ts.T())

	req := httptest.NewRequest(http.MethodGet, "/tenants/", nil)

	token := ts.generateAccessToken(ts.testUser)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)

	require.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	require.NoError(json.NewDecoder(w.Body).Decode(&response))

	tenants, ok := response["tenants"].([]interface{})
	require.True(ok)
	require.Len(tenants, 1)

	tenant := tenants[0].(map[string]interface{})
	require.Equal(ts.testTenant.ID.String(), tenant["id"])
	require.Equal(ts.testTenant.Name, tenant["name"])
}

func (ts *TenantTestSuite) TestAcceptTenantInvitationEmailMismatch() {
	require := require.New(ts.T())

	// Create another user with different email
	differentUser, err := models.NewUser("", "different@example.com", "password", ts.instanceID.String(), nil)
	require.NoError(err)
	require.NoError(ts.API.db.Create(differentUser))

	// Create invitation for a different email
	token := "mismatched-invitation-token"
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	invitation := models.NewTenantInvitation(ts.testTenant.ID, "other@example.com", "member", ts.testUser.ID, token, expiresAt)
	require.NoError(ts.API.db.Create(invitation))

	reqBody := TenantInviteAcceptParams{
		Token: token,
	}

	body, err := json.Marshal(reqBody)
	require.NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/tenants/accept-invitation", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Try to accept with different user
	differentToken := ts.generateAccessToken(differentUser)
	req.Header.Set("Authorization", "Bearer "+differentToken)

	w := httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)

	require.Equal(http.StatusForbidden, w.Code)
}

// Helper method to generate access token for testing
func (ts *TenantTestSuite) generateAccessToken(user *models.User) string {
	// This is a simplified token generation for testing
	// In a real scenario, you would use the proper JWT generation
	token, _, err := ts.API.generateAccessToken(
		httptest.NewRequest("GET", "/", nil),
		ts.API.db,
		user,
		&uuid.Nil, // session ID
		models.PasswordGrant,
	)
	require.NoError(ts.T(), err)
	return token
}

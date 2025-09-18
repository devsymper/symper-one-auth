package models

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/supabase/auth/internal/conf"
	"github.com/supabase/auth/internal/storage"
	"github.com/supabase/auth/internal/storage/test"
)

type TenantTestSuite struct {
	suite.Suite
	db     *storage.Connection
	user   *User
	tenant *Tenant
}

func TestTenant(t *testing.T) {
	globalConfig, err := conf.LoadGlobal(modelsTestConfig)
	require.NoError(t, err)

	conn, err := test.SetupDBConnection(globalConfig)
	require.NoError(t, err)

	ts := &TenantTestSuite{
		db: conn,
	}
	defer ts.db.Close()

	suite.Run(t, ts)
}

func (ts *TenantTestSuite) SetupTest() {
	TruncateAll(ts.db)

	// Create test user
	var err error
	ts.user, err = NewUser("", "test@example.com", "password", "test", nil)
	require.NoError(ts.T(), err)
	require.NoError(ts.T(), ts.db.Create(ts.user))

	// Create a test tenant
	ts.tenant = NewTenant("Test Tenant", "A test tenant", ts.user.ID)
	require.NoError(ts.T(), ts.db.Create(ts.tenant))
}

func (ts *TenantTestSuite) TestNewTenant() {
	require := require.New(ts.T())

	tenant := NewTenant("New Tenant", "Description", ts.user.ID)
	require.NotNil(tenant)
	require.NotEqual(uuid.Nil, tenant.ID)
	require.Equal("New Tenant", tenant.Name)
	require.Equal("Description", tenant.Description)
	require.Equal(ts.user.ID, tenant.OwnerID)
	require.False(tenant.CreatedAt.IsZero())
	require.False(tenant.UpdatedAt.IsZero())
}

func (ts *TenantTestSuite) TestFindTenantByID() {
	require := require.New(ts.T())

	// Test finding existing tenant
	foundTenant, err := FindTenantByID(ts.db, ts.tenant.ID)
	require.NoError(err)
	require.NotNil(foundTenant)
	require.Equal(ts.tenant.ID, foundTenant.ID)
	require.Equal(ts.tenant.Name, foundTenant.Name)

	// Test finding non-existent tenant
	nonExistentID := uuid.Must(uuid.NewV4())
	_, err = FindTenantByID(ts.db, nonExistentID)
	require.Error(err)
	require.True(IsNotFoundError(err))
}

func (ts *TenantTestSuite) TestFindTenantByOwnerID() {
	require := require.New(ts.T())

	// Test finding tenant by owner
	foundTenant, err := FindTenantByOwnerID(ts.db, ts.user.ID)
	require.NoError(err)
	require.NotNil(foundTenant)
	require.Equal(ts.tenant.ID, foundTenant.ID)
	require.Equal(ts.user.ID, foundTenant.OwnerID)

	// Test finding tenant for non-existent owner
	nonExistentOwnerID := uuid.Must(uuid.NewV4())
	_, err = FindTenantByOwnerID(ts.db, nonExistentOwnerID)
	require.Error(err)
	require.True(IsNotFoundError(err))
}

func (ts *TenantTestSuite) TestTenantMember() {
	require := require.New(ts.T())

	// Create another user
	user2, err := NewUser("", "user2@example.com", "password", "test", nil)
	require.NoError(err)
	require.NoError(ts.db.Create(user2))

	// Create tenant member
	member := NewTenantMember(ts.tenant.ID, user2.ID, "admin")
	require.NotNil(member)
	require.NotEqual(uuid.Nil, member.ID)
	require.Equal(ts.tenant.ID, member.TenantID)
	require.Equal(user2.ID, member.UserID)
	require.Equal("admin", member.Role)

	// Save to database
	require.NoError(ts.db.Create(member))

	// Test finding member
	foundMember, err := FindTenantMemberByUserAndTenant(ts.db, user2.ID, ts.tenant.ID)
	require.NoError(err)
	require.NotNil(foundMember)
	require.Equal(member.ID, foundMember.ID)
	require.Equal("admin", foundMember.Role)
}

func (ts *TenantTestSuite) TestGetUserTenants() {
	require := require.New(ts.T())

	// Create another user
	user2, err := NewUser("", "user2@example.com", "password", "test", nil)
	require.NoError(err)
	require.NoError(ts.db.Create(user2))

	// Create another tenant
	tenant2 := NewTenant("Second Tenant", "Another tenant", user2.ID)
	require.NoError(ts.db.Create(tenant2))

	// Add user2 as member to first tenant
	member := NewTenantMember(ts.tenant.ID, user2.ID, "member")
	require.NoError(ts.db.Create(member))

	// Get tenants for user2 (should have 2 tenants)
	tenants, err := GetUserTenants(ts.db, user2.ID)
	require.NoError(err)
	require.Len(tenants, 2)

	// Get tenants for user1 (should have 1 tenant)
	tenants, err = GetUserTenants(ts.db, ts.user.ID)
	require.NoError(err)
	require.Len(tenants, 1)
	require.Equal(ts.tenant.ID, tenants[0].ID)
}

func (ts *TenantTestSuite) TestGetDefaultTenantForUser() {
	require := require.New(ts.T())

	// Test getting default tenant for owner
	defaultTenant, err := GetDefaultTenantForUser(ts.db, ts.user.ID)
	require.NoError(err)
	require.NotNil(defaultTenant)
	require.Equal(ts.tenant.ID, defaultTenant.ID)

	// Create another user and add as member
	user2, err := NewUser("", "user2@example.com", "password", "test", nil)
	require.NoError(err)
	require.NoError(ts.db.Create(user2))

	member := NewTenantMember(ts.tenant.ID, user2.ID, "member")
	require.NoError(ts.db.Create(member))

	// Test getting default tenant for member
	defaultTenant, err = GetDefaultTenantForUser(ts.db, user2.ID)
	require.NoError(err)
	require.NotNil(defaultTenant)
	require.Equal(ts.tenant.ID, defaultTenant.ID)

	// Test getting default tenant for user with no tenants
	user3, err := NewUser("", "user3@example.com", "password", "test", nil)
	require.NoError(err)
	require.NoError(ts.db.Create(user3))

	_, err = GetDefaultTenantForUser(ts.db, user3.ID)
	require.Error(err)
	require.True(IsNotFoundError(err))
}

func (ts *TenantTestSuite) TestTenantInvitation() {
	require := require.New(ts.T())

	// Create invitation
	token := "test-token-123"
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	invitation := NewTenantInvitation(ts.tenant.ID, "invite@example.com", "member", ts.user.ID, token, expiresAt)

	require.NotNil(invitation)
	require.NotEqual(uuid.Nil, invitation.ID)
	require.Equal(ts.tenant.ID, invitation.TenantID)
	require.Equal("invite@example.com", invitation.Email)
	require.Equal("member", invitation.Role)
	require.Equal(ts.user.ID, invitation.InviterID)
	require.Equal(token, invitation.Token)
	require.False(invitation.InvitedAt.IsZero())

	// Save to database
	require.NoError(ts.db.Create(invitation))

	// Test finding by token
	foundInvitation, err := FindTenantInvitationByToken(ts.db, token)
	require.NoError(err)
	require.NotNil(foundInvitation)
	require.Equal(invitation.ID, foundInvitation.ID)
}

func (ts *TenantTestSuite) TestTenantInvitationExpiry() {
	require := require.New(ts.T())

	// Create expired invitation
	expiredTime := time.Now().Add(-1 * time.Hour)
	expiredInvitation := NewTenantInvitation(ts.tenant.ID, "expired@example.com", "member", ts.user.ID, "expired-token", expiredTime)
	require.True(expiredInvitation.IsExpired())

	// Create future invitation
	futureTime := time.Now().Add(1 * time.Hour)
	futureInvitation := NewTenantInvitation(ts.tenant.ID, "future@example.com", "member", ts.user.ID, "future-token", futureTime)
	require.False(futureInvitation.IsExpired())
}

func (ts *TenantTestSuite) TestTenantInvitationAcceptance() {
	require := require.New(ts.T())

	// Create invitation
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	invitation := NewTenantInvitation(ts.tenant.ID, "accept@example.com", "member", ts.user.ID, "accept-token", expiresAt)
	require.False(invitation.IsAccepted())

	// Mark as accepted
	now := time.Now()
	invitation.AcceptedAt = &now
	require.True(invitation.IsAccepted())
}

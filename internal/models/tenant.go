package models

import (
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/supabase/auth/internal/storage"
)

// Tenant represents a tenant organization that users can belong to
type Tenant struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description,omitempty" db:"description"`
	OwnerID     uuid.UUID `json:"owner_id" db:"owner_id"`

	// Metadata
	AppMetaData  JSONMap `json:"app_metadata" db:"raw_app_meta_data"`
	UserMetaData JSONMap `json:"user_metadata" db:"raw_user_meta_data"`

	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Relationships
	Owner *User `json:"owner,omitempty" belongs_to:"user" fk_id:"OwnerID"`

	DONTUSEINSTANCEID uuid.UUID `json:"-" db:"instance_id"`
}

// TenantMember represents the relationship between a user and tenant
type TenantMember struct {
	ID       uuid.UUID `json:"id" db:"id"`
	TenantID uuid.UUID `json:"tenant_id" db:"tenant_id"`
	UserID   uuid.UUID `json:"user_id" db:"user_id"`
	Role     string    `json:"role" db:"role"` // owner, admin, member

	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Relationships
	Tenant *Tenant `json:"tenant,omitempty" belongs_to:"tenant"`
	User   *User   `json:"user,omitempty" belongs_to:"user"`

	DONTUSEINSTANCEID uuid.UUID `json:"-" db:"instance_id"`
}

// TenantInvitation represents an invitation to join a tenant
type TenantInvitation struct {
	ID       uuid.UUID `json:"id" db:"id"`
	TenantID uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Email    string    `json:"email" db:"email"`
	Role     string    `json:"role" db:"role"` // admin, member

	InviterID    uuid.UUID  `json:"inviter_id" db:"inviter_id"`
	InvitedAt    time.Time  `json:"invited_at" db:"invited_at"`
	AcceptedAt   *time.Time `json:"accepted_at,omitempty" db:"accepted_at"`
	ExpiresAt    time.Time  `json:"expires_at" db:"expires_at"`
	Token        string     `json:"-" db:"token"`
	AcceptedByID *uuid.UUID `json:"accepted_by_id,omitempty" db:"accepted_by_id"`

	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Relationships
	Tenant     *Tenant `json:"tenant,omitempty" belongs_to:"tenant"`
	Inviter    *User   `json:"inviter,omitempty" belongs_to:"user" fk_id:"InviterID"`
	AcceptedBy *User   `json:"accepted_by,omitempty" belongs_to:"user" fk_id:"AcceptedByID"`

	DONTUSEINSTANCEID uuid.UUID `json:"-" db:"instance_id"`
}

// TableName overrides the table name used by pop
func (Tenant) TableName() string {
	return "tenants"
}

// TableName overrides the table name used by pop
func (TenantMember) TableName() string {
	return "tenant_members"
}

// TableName overrides the table name used by pop
func (TenantInvitation) TableName() string {
	return "tenant_invitations"
}

// BeforeSave is invoked before the tenant is saved to the database
func (t *Tenant) BeforeSave(tx *pop.Connection) error {
	if t.DeletedAt != nil && t.DeletedAt.IsZero() {
		t.DeletedAt = nil
	}
	return nil
}

// BeforeSave is invoked before the tenant member is saved to the database
func (tm *TenantMember) BeforeSave(tx *pop.Connection) error {
	if tm.DeletedAt != nil && tm.DeletedAt.IsZero() {
		tm.DeletedAt = nil
	}
	return nil
}

// BeforeSave is invoked before the tenant invitation is saved to the database
func (ti *TenantInvitation) BeforeSave(tx *pop.Connection) error {
	if ti.AcceptedAt != nil && ti.AcceptedAt.IsZero() {
		ti.AcceptedAt = nil
	}
	if ti.DeletedAt != nil && ti.DeletedAt.IsZero() {
		ti.DeletedAt = nil
	}
	return nil
}

// NewTenant creates a new tenant with the given name and owner
func NewTenant(name, description string, ownerID uuid.UUID) *Tenant {
	id := uuid.Must(uuid.NewV4())
	now := time.Now()

	return &Tenant{
		ID:          id,
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewTenantMember creates a new tenant member relationship
func NewTenantMember(tenantID, userID uuid.UUID, role string) *TenantMember {
	id := uuid.Must(uuid.NewV4())
	now := time.Now()

	return &TenantMember{
		ID:        id,
		TenantID:  tenantID,
		UserID:    userID,
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewTenantInvitation creates a new tenant invitation
func NewTenantInvitation(tenantID uuid.UUID, email, role string, inviterID uuid.UUID, token string, expiresAt time.Time) *TenantInvitation {
	id := uuid.Must(uuid.NewV4())
	now := time.Now()

	return &TenantInvitation{
		ID:        id,
		TenantID:  tenantID,
		Email:     email,
		Role:      role,
		InviterID: inviterID,
		InvitedAt: now,
		ExpiresAt: expiresAt,
		Token:     token,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// IsExpired checks if the invitation has expired
func (ti *TenantInvitation) IsExpired() bool {
	return time.Now().After(ti.ExpiresAt)
}

// IsAccepted checks if the invitation has been accepted
func (ti *TenantInvitation) IsAccepted() bool {
	return ti.AcceptedAt != nil
}

// FindTenantByID finds a tenant by its ID
func FindTenantByID(tx *storage.Connection, id uuid.UUID) (*Tenant, error) {
	tenant := &Tenant{}
	if err := tx.Find(tenant, id); err != nil {
		if IsNotFoundError(err) {
			return nil, TenantNotFoundError{}
		}
		return nil, err
	}
	return tenant, nil
}

// FindTenantByOwnerID finds the tenant owned by a specific user
func FindTenantByOwnerID(tx *storage.Connection, ownerID uuid.UUID) (*Tenant, error) {
	tenant := &Tenant{}
	if err := tx.Where("owner_id = ? AND deleted_at IS NULL", ownerID).First(tenant); err != nil {
		if IsNotFoundError(err) {
			return nil, TenantNotFoundError{}
		}
		return nil, err
	}
	return tenant, nil
}

// FindTenantInvitationByToken finds an invitation by its token
func FindTenantInvitationByToken(tx *storage.Connection, token string) (*TenantInvitation, error) {
	invitation := &TenantInvitation{}
	if err := tx.Where("token = ? AND deleted_at IS NULL", token).First(invitation); err != nil {
		if IsNotFoundError(err) {
			return nil, TenantInvitationNotFoundError{}
		}
		return nil, err
	}
	return invitation, nil
}

// FindTenantMemberByUserAndTenant finds a tenant membership by user and tenant ID
func FindTenantMemberByUserAndTenant(tx *storage.Connection, userID, tenantID uuid.UUID) (*TenantMember, error) {
	member := &TenantMember{}
	if err := tx.Where("user_id = ? AND tenant_id = ? AND deleted_at IS NULL", userID, tenantID).First(member); err != nil {
		if IsNotFoundError(err) {
			return nil, TenantMemberNotFoundError{}
		}
		return nil, err
	}
	return member, nil
}

// GetUserTenants gets all tenants for a user
func GetUserTenants(tx *storage.Connection, userID uuid.UUID) ([]Tenant, error) {
	var tenants []Tenant

	query := `
		SELECT t.* FROM tenants t
		JOIN tenant_members tm ON t.id = tm.tenant_id
		WHERE tm.user_id = ? AND t.deleted_at IS NULL AND tm.deleted_at IS NULL
		ORDER BY t.created_at DESC
	`

	if err := tx.RawQuery(query, userID).All(&tenants); err != nil {
		return nil, err
	}

	return tenants, nil
}

// GetDefaultTenantForUser gets the default tenant for a user (their owned tenant or first member tenant)
func GetDefaultTenantForUser(tx *storage.Connection, userID uuid.UUID) (*Tenant, error) {
	// First try to find a tenant they own
	tenant, err := FindTenantByOwnerID(tx, userID)
	if err == nil {
		return tenant, nil
	}

	// If no owned tenant, get their first tenant membership
	tenants, err := GetUserTenants(tx, userID)
	if err != nil {
		return nil, err
	}

	if len(tenants) == 0 {
		return nil, TenantNotFoundError{}
	}

	return &tenants[0], nil
}

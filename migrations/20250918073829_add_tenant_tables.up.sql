-- Create tenants table
CREATE TABLE IF NOT EXISTS {{ index .Options "Namespace" }}.tenants (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    instance_id uuid NULL,
    name varchar(255) NOT NULL,
    description text NULL,
    owner_id uuid NOT NULL,
    raw_app_meta_data jsonb NULL,
    raw_user_meta_data jsonb NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL,
    CONSTRAINT tenants_pkey PRIMARY KEY (id),
    CONSTRAINT tenants_owner_id_fkey FOREIGN KEY (owner_id) REFERENCES {{ index .Options "Namespace" }}.users(id) ON DELETE CASCADE
);

-- Create tenant_members table
CREATE TABLE IF NOT EXISTS {{ index .Options "Namespace" }}.tenant_members (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    instance_id uuid NULL,
    tenant_id uuid NOT NULL,
    user_id uuid NOT NULL,
    role varchar(50) NOT NULL DEFAULT 'member',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL,
    CONSTRAINT tenant_members_pkey PRIMARY KEY (id),
    CONSTRAINT tenant_members_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES {{ index .Options "Namespace" }}.tenants(id) ON DELETE CASCADE,
    CONSTRAINT tenant_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES {{ index .Options "Namespace" }}.users(id) ON DELETE CASCADE,
    CONSTRAINT tenant_members_tenant_user_unique UNIQUE (tenant_id, user_id)
);

-- Create tenant_invitations table
CREATE TABLE IF NOT EXISTS {{ index .Options "Namespace" }}.tenant_invitations (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    instance_id uuid NULL,
    tenant_id uuid NOT NULL,
    email varchar(255) NOT NULL,
    role varchar(50) NOT NULL DEFAULT 'member',
    inviter_id uuid NOT NULL,
    invited_at timestamptz NOT NULL DEFAULT now(),
    accepted_at timestamptz NULL,
    expires_at timestamptz NOT NULL,
    token varchar(255) NOT NULL UNIQUE,
    accepted_by_id uuid NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL,
    CONSTRAINT tenant_invitations_pkey PRIMARY KEY (id),
    CONSTRAINT tenant_invitations_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES {{ index .Options "Namespace" }}.tenants(id) ON DELETE CASCADE,
    CONSTRAINT tenant_invitations_inviter_id_fkey FOREIGN KEY (inviter_id) REFERENCES {{ index .Options "Namespace" }}.users(id) ON DELETE CASCADE,
    CONSTRAINT tenant_invitations_accepted_by_id_fkey FOREIGN KEY (accepted_by_id) REFERENCES {{ index .Options "Namespace" }}.users(id) ON DELETE SET NULL
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS tenant_members_tenant_id_idx ON {{ index .Options "Namespace" }}.tenant_members USING btree (tenant_id);
CREATE INDEX IF NOT EXISTS tenant_members_user_id_idx ON {{ index .Options "Namespace" }}.tenant_members USING btree (user_id);
CREATE INDEX IF NOT EXISTS tenant_invitations_tenant_id_idx ON {{ index .Options "Namespace" }}.tenant_invitations USING btree (tenant_id);
CREATE INDEX IF NOT EXISTS tenant_invitations_email_idx ON {{ index .Options "Namespace" }}.tenant_invitations USING btree (email);
CREATE INDEX IF NOT EXISTS tenant_invitations_token_idx ON {{ index .Options "Namespace" }}.tenant_invitations USING btree (token);
CREATE INDEX IF NOT EXISTS tenant_invitations_expires_at_idx ON {{ index .Options "Namespace" }}.tenant_invitations USING btree (expires_at);
CREATE INDEX IF NOT EXISTS tenants_owner_id_idx ON {{ index .Options "Namespace" }}.tenants USING btree (owner_id);

-- Add comments for documentation
COMMENT ON TABLE {{ index .Options "Namespace" }}.tenants IS 'Auth: Stores tenant organizations that users can belong to';
COMMENT ON TABLE {{ index .Options "Namespace" }}.tenant_members IS 'Auth: Stores the relationship between users and tenants';
COMMENT ON TABLE {{ index .Options "Namespace" }}.tenant_invitations IS 'Auth: Stores invitations for users to join tenants';

-- Add constraint to ensure valid roles
ALTER TABLE {{ index .Options "Namespace" }}.tenant_members 
ADD CONSTRAINT tenant_members_role_check CHECK (role IN ('owner', 'admin', 'member'));

ALTER TABLE {{ index .Options "Namespace" }}.tenant_invitations 
ADD CONSTRAINT tenant_invitations_role_check CHECK (role IN ('admin', 'member'));

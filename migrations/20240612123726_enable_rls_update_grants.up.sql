do $$ begin
    -- enable RLS policy on auth tables
    alter table {{ index .Options "Namespace" }}.schema_migrations enable row level security;
    alter table {{ index .Options "Namespace" }}.instances enable row level security;
    alter table {{ index .Options "Namespace" }}.users enable row level security;
    alter table {{ index .Options "Namespace" }}.audit_log_entries enable row level security;
    alter table {{ index .Options "Namespace" }}.saml_relay_states enable row level security;
    alter table {{ index .Options "Namespace" }}.refresh_tokens enable row level security;
    alter table {{ index .Options "Namespace" }}.mfa_factors enable row level security;
    alter table {{ index .Options "Namespace" }}.sessions enable row level security;
    alter table {{ index .Options "Namespace" }}.sso_providers enable row level security;
    alter table {{ index .Options "Namespace" }}.sso_domains enable row level security;
    alter table {{ index .Options "Namespace" }}.mfa_challenges enable row level security;
    alter table {{ index .Options "Namespace" }}.mfa_amr_claims enable row level security;
    alter table {{ index .Options "Namespace" }}.saml_providers enable row level security;
    alter table {{ index .Options "Namespace" }}.flow_state enable row level security;
    alter table {{ index .Options "Namespace" }}.identities enable row level security;
    alter table {{ index .Options "Namespace" }}.one_time_tokens enable row level security;
    -- allow symper_one_auth role to select from auth tables and allow it to grant select to other roles
    grant select on {{ index .Options "Namespace" }}.schema_migrations to symper_one_auth with grant option;
    grant select on {{ index .Options "Namespace" }}.instances to symper_one_auth with grant option;
    grant select on {{ index .Options "Namespace" }}.users to symper_one_auth with grant option;
    grant select on {{ index .Options "Namespace" }}.audit_log_entries to symper_one_auth with grant option;
    grant select on {{ index .Options "Namespace" }}.saml_relay_states to symper_one_auth with grant option;
    grant select on {{ index .Options "Namespace" }}.refresh_tokens to symper_one_auth with grant option;
    grant select on {{ index .Options "Namespace" }}.mfa_factors to symper_one_auth with grant option;
    grant select on {{ index .Options "Namespace" }}.sessions to symper_one_auth with grant option;
    grant select on {{ index .Options "Namespace" }}.sso_providers to symper_one_auth with grant option;
    grant select on {{ index .Options "Namespace" }}.sso_domains to symper_one_auth with grant option;
    grant select on {{ index .Options "Namespace" }}.mfa_challenges to symper_one_auth with grant option;
    grant select on {{ index .Options "Namespace" }}.mfa_amr_claims to symper_one_auth with grant option;
    grant select on {{ index .Options "Namespace" }}.saml_providers to symper_one_auth with grant option;
    grant select on {{ index .Options "Namespace" }}.flow_state to symper_one_auth with grant option;
    grant select on {{ index .Options "Namespace" }}.identities to symper_one_auth with grant option;
    grant select on {{ index .Options "Namespace" }}.one_time_tokens to symper_one_auth with grant option;
end $$;

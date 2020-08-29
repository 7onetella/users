package dbutil

var DBSchema = `
DROP TABLE IF EXISTS users CASCADE;
CREATE TABLE users (
    user_id              UUID PRIMARY KEY,
    platform_name        CHARACTER VARYING(64),
    email                CHARACTER VARYING(128) NOT NULL,
    passhash             CHARACTER VARYING(128) NOT NULL,
    firstname            CHARACTER VARYING(64),
    lastname             CHARACTER VARYING(64),
    created_date         BIGINT DEFAULT 0,
    totp_enabled         BOOL DEFAULT FALSE,
    totp_secret_current  CHARACTER VARYING(32) DEFAULT '',
    totp_secret_tmp      CHARACTER VARYING(32) DEFAULT '',
    totp_secret_tmp_exp  BIGINT DEFAULT 0,
	jwt_secret           CHARACTER VARYING(32) DEFAULT '',
    webauthn_enabled     BOOL DEFAULT FALSE,
    webauthn_session     CHARACTER VARYING(2048) DEFAULT '',
    CONSTRAINT           unique_user UNIQUE (platform_name, email)
);

DROP TABLE IF EXISTS auth_event CASCADE;
CREATE TABLE auth_event (
    event_id            UUID PRIMARY KEY,
    user_id             CHARACTER VARYING(40) DEFAULT '',
    event               CHARACTER VARYING(64) DEFAULT '',
    event_timestamp     BIGINT DEFAULT 0,
    ip_v4               CHARACTER VARYING(15) DEFAULT '',
    ip_v6               CHARACTER VARYING(38) DEFAULT '',
    agent               CHARACTER VARYING(128) DEFAULT ''
);

DROP TABLE IF EXISTS user_credential CASCADE;
CREATE TABLE user_credential (
    credential_id        UUID PRIMARY KEY,
    user_id              CHARACTER VARYING(40) DEFAULT '',
    credential           CHARACTER VARYING(1024) DEFAULT '',
    CONSTRAINT           unique_cred_user UNIQUE (credential_id, user_id)
);

DROP TABLE IF EXISTS clients CASCADE;
CREATE TABLE clients (
    client_id            UUID PRIMARY KEY,
    name                 CHARACTER VARYING(40) DEFAULT '',
    secret               CHARACTER VARYING(1024) DEFAULT '',
    CONSTRAINT           unique_client_name UNIQUE (name)
);

DROP TABLE IF EXISTS user_grants CASCADE;
CREATE TABLE user_grants (
    user_id              CHARACTER VARYING(40)   DEFAULT '',
    client_id            CHARACTER VARYING(40)   DEFAULT '',
    scope                CHARACTER VARYING(1024) DEFAULT '',
    CONSTRAINT           unique_user_to_client_grants UNIQUE (user_id, client_id)
);

DROP TABLE IF EXISTS authorization_code CASCADE;
CREATE TABLE authorization_code (
    code                 CHARACTER VARYING(1024) DEFAULT '',
    client_id            CHARACTER VARYING(40)   DEFAULT '',
    created_at           BIGINT DEFAULT 0,
    user_id              CHARACTER VARYING(40)   DEFAULT '',
	CONSTRAINT           unique_authorization_code UNIQUE (code, client_id)
);

DROP TABLE IF EXISTS access_token CASCADE;
CREATE TABLE access_token (
    token_id             CHARACTER VARYING(40)   DEFAULT '',
    user_id              CHARACTER VARYING(40)   DEFAULT '',
    token                CHARACTER VARYING(1024) DEFAULT '',
	CONSTRAINT           unique_access_token UNIQUE (token_id, user_id)
);

INSERT INTO authorization_code (code, client_id, created_at, user_id) VALUES (
	'f7cd9875-8386-4d16-97ef-7ae858ebe4c2', 
	'352b6e64-e498-4307-b64d-ec9e5b9da65c', 
	1598737919, 
	'a2aee5e6-05a0-438c-9276-4ba406b7bf9e'
);

INSERT INTO user_grants (user_id, client_id, scope) VALUES (
	'a2aee5e6-05a0-438c-9276-4ba406b7bf9e', 
	'352b6e64-e498-4307-b64d-ec9e5b9da65c', 
	'read:profile,write:profile'
);

INSERT INTO users 
            (user_id, 
             platform_name, 
             email, 
             passhash, 
             firstname, 
             lastname, 
             created_date, 
             totp_enabled, 
             totp_secret_current, 
             totp_secret_tmp, 
             totp_secret_tmp_exp, 
             jwt_secret, 
             webauthn_enabled, 
             webauthn_session) 
VALUES      ('ee288e8c-0b2a-41b5-937c-9a355c0483b4', 
             'web', 
             'totp_user@example.com', 
             'users91234', 
             'Foo', 
             'Bar', 
             1597042574, 
             true, 
             'FPTUDIF2KSQAKREU', 
             '', 
             0, 
             'FPTUDIF2KSQAKREU', 
             false, 
             ''); 

INSERT INTO users 
            (user_id, 
             platform_name, 
             email, 
             passhash, 
             firstname, 
             lastname, 
             created_date, 
             totp_enabled, 
             totp_secret_current, 
             totp_secret_tmp, 
             totp_secret_tmp_exp, 
             jwt_secret, 
             webauthn_enabled, 
             webauthn_session) 
VALUES      ('a2aee5e6-05a0-438c-9276-4ba406b7bf9e', 
             'web', 
             'John.Smith@example.com', 
             'users91234', 
             'John', 
             'Smith', 
             1596747095, 
             false, 
             'SVVEC5VTQBMNE3DH', 
             'C56BRBHMW3YC4XPA', 
             1597089055, 
             'SVVEC5VTQBMNE3DH', 
             false, 
             ''); 

INSERT INTO users 
            (user_id, 
             platform_name, 
             email, 
             passhash, 
             firstname, 
             lastname, 
             created_date, 
             totp_enabled, 
             totp_secret_current, 
             totp_secret_tmp, 
             totp_secret_tmp_exp, 
             jwt_secret, 
             webauthn_enabled, 
             webauthn_session) 
VALUES      ('a2aee5e6-05a0-438c-9276-4ba406b7bf9f', 
             'web', 
             'webauth_user@example.com', 
             'users91234', 
             'Mary', 
             'Smith', 
             1596747095, 
             false, 
             'SVVEC5VTQBMNE3DH', 
             'C56BRBHMW3YC4XPA', 
             1597089055, 
             'SVVEC5VTQBMNE3DH', 
             true, 
             ''); 

INSERT INTO user_credential 
            (credential_id, 
             user_id, 
             credential) 
VALUES      ('113af9e8-810d-45fe-81d7-0eefb40390bf', 
             'a2aee5e6-05a0-438c-9276-4ba406b7bf9f', 
'{"ID":"L1Mfkf2IL/mzT4xbGheLCG/5dAPA487hIU+bfRWTAu4EQ50AtdON99F09d9EebHBesLmYCU3rU/Czdv0Bcopmg==","PublicKey":"pQMmIAEhWCCljDUqf+Ug6KgYIG/mo7PDBo9x6rVFQt4rPZ1lJhvrmyJYIHdLOCFt0Tv4buqgILjGng6KbxuBpafmhWHePNKIyw/tAQI=","AttestationType":"none","Authenticator":{"AAGUID":"AAAAAAAAAAAAAAAAAAAAAA==","SignCount":3,"CloneWarning":false}}' 
);

INSERT INTO clients (client_id, name, secret) VALUES ('352b6e64-e498-4307-b64d-ec9e5b9da65c', 'my password app', '9191');

`

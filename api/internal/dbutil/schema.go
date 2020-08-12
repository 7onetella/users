package dbutil

var DBSchema = `
DROP TABLE users;

CREATE TABLE users (
    user_id             UUID PRIMARY KEY,
    platform_name       CHARACTER VARYING(64),
    email               CHARACTER VARYING(128) NOT NULL,
    passhash            CHARACTER VARYING(128) NOT NULL,
    firstname           CHARACTER VARYING(64),
    lastname            CHARACTER VARYING(64),
    created_date        BIGINT DEFAULT 0,
    mfa_enabled         BOOL DEFAULT FALSE,
    mfa_secret_current  CHARACTER VARYING(32) DEFAULT '',
    mfa_secret_tmp      CHARACTER VARYING(32) DEFAULT '',
    mfa_secret_tmp_exp  BIGINT DEFAULT 0,
	jwt_secret          CHARACTER VARYING(32) DEFAULT '',
    CONSTRAINT          unique_user UNIQUE (platform_name, email)
);

INSERT INTO users 
		(user_id, platform_name, email, passhash, firstname, lastname, created_date, mfa_enabled, mfa_secret_current, mfa_secret_tmp, mfa_secret_tmp_exp, jwt_secret) 
VALUES 
		('ee288e8c-0b2a-41b5-937c-9a355c0483b4', 'web', 'scott@example.com', 'password', 'scott', 'bar', 1597042574, true, 'FPTUDIF2KSQAKREU', '', 0, 'FPTUDIF2KSQAKREU');

INSERT INTO users 
		(user_id, platform_name, email, passhash, firstname, lastname, created_date, mfa_enabled, mfa_secret_current, mfa_secret_tmp, mfa_secret_tmp_exp, jwt_secret) 
VALUES 
		('a2aee5e6-05a0-438c-9276-4ba406b7bf9e', 'web', 'user8az28y@example.com', 'password', 'scott', 'bar', 1596747095, false, 'SVVEC5VTQBMNE3DH', 'C56BRBHMW3YC4XPA', 1597089055, 'SVVEC5VTQBMNE3DH');

DROP TABLE auth_event;

CREATE TABLE auth_event (
    event_id            UUID PRIMARY KEY,
    user_id             CHARACTER(40) DEFAULT '',
    event               CHARACTER VARYING(64) DEFAULT '',
    event_timestamp     BIGINT DEFAULT 0,
    ip_v4               CHARACTER(15) DEFAULT '',
    ip_v6               CHARACTER(38) DEFAULT '',
    agent               CHARACTER VARYING(128) DEFAULT ''
);
`

CREATE TABLE org_invitations(
	id bigserial NOT NULL PRIMARY KEY,
	org_id integer NOT NULL REFERENCES orgs (id),
	sender_user_id integer NOT NULL REFERENCES users (id),
	recipient_user_id integer NOT NULL REFERENCES users (id),
	created_at timestamp with time zone NOT NULL DEFAULT now(),
	notified_at timestamp with time zone,
	responded_at timestamp with time zone,
	response_type boolean,
	revoked_at timestamp with time zone,
	deleted_at timestamp with time zone
);

CREATE INDEX org_invitations_org_id ON org_invitations(org_id) WHERE deleted_at IS NULL;
CREATE INDEX org_invitations_recipient_user_id ON org_invitations(recipient_user_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX org_invitations_singleflight ON org_invitations(org_id, recipient_user_id) WHERE responded_at IS NULL AND revoked_at IS NULL AND deleted_at IS NULL;
ALTER TABLE org_invitations ADD CONSTRAINT check_atomic_response CHECK((responded_at IS NULL) = (response_type IS NULL));
ALTER TABLE org_invitations ADD CONSTRAINT check_single_use CHECK((responded_at IS NULL AND response_type IS NULL) OR revoked_at IS NULL);

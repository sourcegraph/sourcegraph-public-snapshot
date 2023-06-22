ALTER TABLE lsif_configuration_policies_repository_pattern_lookup
    ADD CONSTRAINT fk_policy_id
        FOREIGN KEY (policy_id)
            REFERENCES lsif_configuration_policies (id)
            ON DELETE CASCADE
        NOT VALID;

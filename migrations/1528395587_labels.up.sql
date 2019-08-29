BEGIN;

-- See https://gitlab.com/gitlab-org/gitlab-ce/issues/19997#note_16081366 for why we don't have
-- group labels, only per-repository. labels. Instead of group labels, we will support batch
-- editing of labels across multiple repositories.
CREATE TABLE labels (
	id bigserial PRIMARY KEY,
    repository_id integer REFERENCES repo(id) ON DELETE CASCADE,
	name citext NOT NULL,
	description text,
	color text NOT NULL
);
CREATE INDEX labels_repository_id ON labels(repository_id);
CREATE INDEX labels_name ON labels(name);
CREATE UNIQUE INDEX labels_name_project_uniq ON labels(name, repository_id);

CREATE TABLE labels_objects (
	label_id bigint NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
	thread_id bigint REFERENCES threads(id) ON DELETE CASCADE
);
CREATE INDEX labels_objects_label_id ON labels_objects(label_id);
CREATE INDEX labels_objects_thread_id ON labels_objects(thread_id) WHERE thread_id IS NOT NULL;
CREATE UNIQUE INDEX labels_objects_uniq ON labels_objects(label_id, thread_id);

COMMIT;

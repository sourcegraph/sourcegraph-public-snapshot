/**
 * The directory relative to the storage where SQLite databases are located.
 */
export const DBS_DIR = 'dbs'

/**
 * The directory relative to the storage where scratch SQLite databases create created during processing.
 */
export const TEMP_DIR = 'temp'

/**
 * The directory relative to the storage where raw dumps are uploaded.
 */
export const UPLOADS_DIR = 'uploads'

/**
 * The maximum number of rows to bulk insert in Postgres.
 */
export const MAX_POSTGRES_BATCH_SIZE = 5000

/**
 * The maximum number of commits to visit breadth-first style when when finding
 * the closest commit.
 */
export const MAX_TRAVERSAL_LIMIT = 100

/**
 * A random integer specific to the xrepo database used to generate advisory lock ids.
 */
export const ADVISORY_LOCK_ID_SALT = 1688730858

/**
 * The number of commits to ask gitserver for when updating commit data for
 * a particular repository. This should be just slightly above the max traversal
 * limit.
 */
export const MAX_COMMITS_PER_UPDATE = MAX_TRAVERSAL_LIMIT * 1.5

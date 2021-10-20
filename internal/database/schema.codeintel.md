# Table "public.codeintel_schema_migrations"
```
 Column  |  Type   | Collation | Nullable | Default 
---------+---------+-----------+----------+---------
 version | bigint  |           | not null | 
 dirty   | boolean |           | not null | 
Indexes:
    "codeintel_schema_migrations_pkey" PRIMARY KEY, btree (version)

```

# Table "public.lsif_data_apidocs_num_dumps"
```
 Column |  Type  | Collation | Nullable | Default 
--------+--------+-----------+----------+---------
 count  | bigint |           |          | 

```

# Table "public.lsif_data_apidocs_num_dumps_indexed"
```
 Column |  Type  | Collation | Nullable | Default 
--------+--------+-----------+----------+---------
 count  | bigint |           |          | 

```

# Table "public.lsif_data_apidocs_num_pages"
```
 Column |  Type  | Collation | Nullable | Default 
--------+--------+-----------+----------+---------
 count  | bigint |           |          | 

```

# Table "public.lsif_data_apidocs_num_search_results_private"
```
 Column |  Type  | Collation | Nullable | Default 
--------+--------+-----------+----------+---------
 count  | bigint |           |          | 

```

# Table "public.lsif_data_apidocs_num_search_results_public"
```
 Column |  Type  | Collation | Nullable | Default 
--------+--------+-----------+----------+---------
 count  | bigint |           |          | 

```

# Table "public.lsif_data_definitions"
```
     Column     |  Type   | Collation | Nullable | Default 
----------------+---------+-----------+----------+---------
 dump_id        | integer |           | not null | 
 scheme         | text    |           | not null | 
 identifier     | text    |           | not null | 
 data           | bytea   |           |          | 
 schema_version | integer |           | not null | 
 num_locations  | integer |           | not null | 
Indexes:
    "lsif_data_definitions_pkey" PRIMARY KEY, btree (dump_id, scheme, identifier)
    "lsif_data_definitions_dump_id_schema_version" btree (dump_id, schema_version)
Triggers:
    lsif_data_definitions_schema_versions_insert AFTER INSERT ON lsif_data_definitions REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_lsif_data_definitions_schema_versions_insert()

```

Associates (document, range) pairs with the import monikers attached to the range.

**data**: A gob-encoded payload conforming to an array of [LocationData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L106:6) types.

**dump_id**: The identifier of the associated dump in the lsif_uploads table (state=completed).

**identifier**: The moniker identifier.

**num_locations**: The number of locations stored in the data field.

**schema_version**: The schema version of this row - used to determine presence and encoding of data.

**scheme**: The moniker scheme.

# Table "public.lsif_data_definitions_schema_versions"
```
       Column       |  Type   | Collation | Nullable | Default 
--------------------+---------+-----------+----------+---------
 dump_id            | integer |           | not null | 
 min_schema_version | integer |           |          | 
 max_schema_version | integer |           |          | 
Indexes:
    "lsif_data_definitions_schema_versions_pkey" PRIMARY KEY, btree (dump_id)
    "lsif_data_definitions_schema_versions_dump_id_schema_version_bo" btree (dump_id, min_schema_version, max_schema_version)

```

Tracks the range of schema_versions for each upload in the lsif_data_definitions table.

**dump_id**: The identifier of the associated dump in the lsif_uploads table.

**max_schema_version**: An upper-bound on the `lsif_data_definitions.schema_version` where `lsif_data_definitions.dump_id = dump_id`.

**min_schema_version**: A lower-bound on the `lsif_data_definitions.schema_version` where `lsif_data_definitions.dump_id = dump_id`.

# Table "public.lsif_data_docs_search_current_private"
```
        Column        |           Type           | Collation | Nullable | Default 
----------------------+--------------------------+-----------+----------+---------
 repo_id              | integer                  |           | not null | 
 dump_root            | text                     |           | not null | 
 lang_name_id         | integer                  |           | not null | 
 dump_id              | integer                  |           | not null | 
 last_cleanup_scan_at | timestamp with time zone |           | not null | 
Indexes:
    "lsif_data_docs_search_current_private_pkey" PRIMARY KEY, btree (repo_id, dump_root, lang_name_id)

```

A table indicating the most current search index for a unique repository, root, and language.

**dump_id**: The most recent dump identifier for this key. See associated content in the lsif_data_docs_search_private table.

**dump_root**: The root of the associated dump.

**lang_name_id**: The interned index name of the associated dump.

**last_cleanup_scan_at**: The last time outdated records in the lsif_data_docs_search_private table have been cleaned.

**repo_id**: The repository identifier of the associated dump.

# Table "public.lsif_data_docs_search_current_public"
```
        Column        |           Type           | Collation | Nullable | Default 
----------------------+--------------------------+-----------+----------+---------
 repo_id              | integer                  |           | not null | 
 dump_root            | text                     |           | not null | 
 lang_name_id         | integer                  |           | not null | 
 dump_id              | integer                  |           | not null | 
 last_cleanup_scan_at | timestamp with time zone |           | not null | 
Indexes:
    "lsif_data_docs_search_current_public_pkey" PRIMARY KEY, btree (repo_id, dump_root, lang_name_id)

```

A table indicating the most current search index for a unique repository, root, and language.

**dump_id**: The most recent dump identifier for this key. See associated content in the lsif_data_docs_search_public table.

**dump_root**: The root of the associated dump.

**lang_name_id**: The interned index name of the associated dump.

**last_cleanup_scan_at**: The last time outdated records in the lsif_data_docs_search_public table have been cleaned.

**repo_id**: The repository identifier of the associated dump.

# Table "public.lsif_data_docs_search_lang_names_private"
```
  Column   |   Type   | Collation | Nullable |                               Default                                
-----------+----------+-----------+----------+----------------------------------------------------------------------
 id        | bigint   |           | not null | nextval('lsif_data_docs_search_lang_names_private_id_seq'::regclass)
 lang_name | text     |           | not null | 
 tsv       | tsvector |           | not null | 
Indexes:
    "lsif_data_docs_search_lang_names_private_pkey" PRIMARY KEY, btree (id)
    "lsif_data_docs_search_lang_names_private_lang_name_key" UNIQUE CONSTRAINT, btree (lang_name)
    "lsif_data_docs_search_lang_names_private_tsv_idx" gin (tsv)
Referenced by:
    TABLE "lsif_data_docs_search_private" CONSTRAINT "lsif_data_docs_search_private_lang_name_id_fk" FOREIGN KEY (lang_name_id) REFERENCES lsif_data_docs_search_lang_names_private(id)

```

Each unique language name being stored in the API docs search index.

**id**: The ID of the language name.

**lang_name**: The lowercase language name (go, java, etc.) OR, if unknown, the LSIF indexer name.

**tsv**: Indexed tsvector for the lang_name field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.

# Table "public.lsif_data_docs_search_lang_names_public"
```
  Column   |   Type   | Collation | Nullable |                               Default                               
-----------+----------+-----------+----------+---------------------------------------------------------------------
 id        | bigint   |           | not null | nextval('lsif_data_docs_search_lang_names_public_id_seq'::regclass)
 lang_name | text     |           | not null | 
 tsv       | tsvector |           | not null | 
Indexes:
    "lsif_data_docs_search_lang_names_public_pkey" PRIMARY KEY, btree (id)
    "lsif_data_docs_search_lang_names_public_lang_name_key" UNIQUE CONSTRAINT, btree (lang_name)
    "lsif_data_docs_search_lang_names_public_tsv_idx" gin (tsv)
Referenced by:
    TABLE "lsif_data_docs_search_public" CONSTRAINT "lsif_data_docs_search_public_lang_name_id_fk" FOREIGN KEY (lang_name_id) REFERENCES lsif_data_docs_search_lang_names_public(id)

```

Each unique language name being stored in the API docs search index.

**id**: The ID of the language name.

**lang_name**: The lowercase language name (go, java, etc.) OR, if unknown, the LSIF indexer name.

**tsv**: Indexed tsvector for the lang_name field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.

# Table "public.lsif_data_docs_search_private"
```
         Column         |   Type   | Collation | Nullable |                          Default                          
------------------------+----------+-----------+----------+-----------------------------------------------------------
 id                     | bigint   |           | not null | nextval('lsif_data_docs_search_private_id_seq'::regclass)
 repo_id                | integer  |           | not null | 
 dump_id                | integer  |           | not null | 
 dump_root              | text     |           | not null | 
 path_id                | text     |           | not null | 
 detail                 | text     |           | not null | 
 lang_name_id           | integer  |           | not null | 
 repo_name_id           | integer  |           | not null | 
 tags_id                | integer  |           | not null | 
 search_key             | text     |           | not null | 
 search_key_tsv         | tsvector |           | not null | 
 search_key_reverse_tsv | tsvector |           | not null | 
 label                  | text     |           | not null | 
 label_tsv              | tsvector |           | not null | 
 label_reverse_tsv      | tsvector |           | not null | 
Indexes:
    "lsif_data_docs_search_private_pkey" PRIMARY KEY, btree (id)
    "lsif_data_docs_search_private_dump_id_idx" btree (dump_id)
    "lsif_data_docs_search_private_dump_root_idx" btree (dump_root)
    "lsif_data_docs_search_private_label_reverse_tsv_idx" btree (label_reverse_tsv)
    "lsif_data_docs_search_private_label_tsv_idx" btree (label_tsv)
    "lsif_data_docs_search_private_repo_id_idx" btree (repo_id)
    "lsif_data_docs_search_private_search_key_reverse_tsv_idx" btree (search_key_reverse_tsv)
    "lsif_data_docs_search_private_search_key_tsv_idx" btree (search_key_tsv)
Foreign-key constraints:
    "lsif_data_docs_search_private_lang_name_id_fk" FOREIGN KEY (lang_name_id) REFERENCES lsif_data_docs_search_lang_names_private(id)
    "lsif_data_docs_search_private_repo_name_id_fk" FOREIGN KEY (repo_name_id) REFERENCES lsif_data_docs_search_repo_names_private(id)
    "lsif_data_docs_search_private_tags_id_fk" FOREIGN KEY (tags_id) REFERENCES lsif_data_docs_search_tags_private(id)
Triggers:
    lsif_data_docs_search_private_delete AFTER DELETE ON lsif_data_docs_search_private REFERENCING OLD TABLE AS oldtbl FOR EACH STATEMENT EXECUTE FUNCTION lsif_data_docs_search_private_delete()
    lsif_data_docs_search_private_insert AFTER INSERT ON lsif_data_docs_search_private REFERENCING NEW TABLE AS newtbl FOR EACH STATEMENT EXECUTE FUNCTION lsif_data_docs_search_private_insert()

```

A tsvector search index over API documentation (private repos only)

**detail**: The detail string (e.g. the full function signature and its docs). See protocol/documentation.go:Documentation

**dump_id**: The identifier of the associated dump in the lsif_uploads table (state=completed).

**dump_root**: Identical to lsif_dumps.root; The working directory of the indexer image relative to the repository root.

**id**: The row ID of the search result.

**label**: The label string of the result, e.g. a one-line function signature. See protocol/documentation.go:Documentation

**label_reverse_tsv**: Indexed tsvector for the reverse of the label field, for suffix lexeme/word matching. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.

**label_tsv**: Indexed tsvector for the label field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.

**lang_name_id**: The programming language (or indexer name) that produced the result. Foreign key into lsif_data_docs_search_lang_names_private.

**path_id**: The fully qualified documentation page path ID, e.g. including "#section". See GraphQL codeintel.schema:documentationPage for what this is.

**repo_id**: The repo ID, from the main app DB repo table. Used to search over a select set of repos by ID.

**repo_name_id**: The repository name that produced the result. Foreign key into lsif_data_docs_search_repo_names_private.

**search_key**: The search key generated by the indexer, e.g. mux.Router.ServeHTTP. It is language-specific, and likely unique within a repository (but not always.) See protocol/documentation.go:Documentation.SearchKey

**search_key_reverse_tsv**: Indexed tsvector for the reverse of the search_key field, for suffix lexeme/word matching. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.

**search_key_tsv**: Indexed tsvector for the search_key field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.

**tags_id**: The tags from the documentation node. Foreign key into lsif_data_docs_search_tags_private.

# Table "public.lsif_data_docs_search_public"
```
         Column         |   Type   | Collation | Nullable |                         Default                          
------------------------+----------+-----------+----------+----------------------------------------------------------
 id                     | bigint   |           | not null | nextval('lsif_data_docs_search_public_id_seq'::regclass)
 repo_id                | integer  |           | not null | 
 dump_id                | integer  |           | not null | 
 dump_root              | text     |           | not null | 
 path_id                | text     |           | not null | 
 detail                 | text     |           | not null | 
 lang_name_id           | integer  |           | not null | 
 repo_name_id           | integer  |           | not null | 
 tags_id                | integer  |           | not null | 
 search_key             | text     |           | not null | 
 search_key_tsv         | tsvector |           | not null | 
 search_key_reverse_tsv | tsvector |           | not null | 
 label                  | text     |           | not null | 
 label_tsv              | tsvector |           | not null | 
 label_reverse_tsv      | tsvector |           | not null | 
Indexes:
    "lsif_data_docs_search_public_pkey" PRIMARY KEY, btree (id)
    "lsif_data_docs_search_public_dump_id_idx" btree (dump_id)
    "lsif_data_docs_search_public_dump_root_idx" btree (dump_root)
    "lsif_data_docs_search_public_label_reverse_tsv_idx" btree (label_reverse_tsv)
    "lsif_data_docs_search_public_label_tsv_idx" btree (label_tsv)
    "lsif_data_docs_search_public_repo_id_idx" btree (repo_id)
    "lsif_data_docs_search_public_search_key_reverse_tsv_idx" btree (search_key_reverse_tsv)
    "lsif_data_docs_search_public_search_key_tsv_idx" btree (search_key_tsv)
Foreign-key constraints:
    "lsif_data_docs_search_public_lang_name_id_fk" FOREIGN KEY (lang_name_id) REFERENCES lsif_data_docs_search_lang_names_public(id)
    "lsif_data_docs_search_public_repo_name_id_fk" FOREIGN KEY (repo_name_id) REFERENCES lsif_data_docs_search_repo_names_public(id)
    "lsif_data_docs_search_public_tags_id_fk" FOREIGN KEY (tags_id) REFERENCES lsif_data_docs_search_tags_public(id)
Triggers:
    lsif_data_docs_search_public_delete AFTER DELETE ON lsif_data_docs_search_public REFERENCING OLD TABLE AS oldtbl FOR EACH STATEMENT EXECUTE FUNCTION lsif_data_docs_search_public_delete()
    lsif_data_docs_search_public_insert AFTER INSERT ON lsif_data_docs_search_public REFERENCING NEW TABLE AS newtbl FOR EACH STATEMENT EXECUTE FUNCTION lsif_data_docs_search_public_insert()

```

A tsvector search index over API documentation (public repos only)

**detail**: The detail string (e.g. the full function signature and its docs). See protocol/documentation.go:Documentation

**dump_id**: The identifier of the associated dump in the lsif_uploads table (state=completed).

**dump_root**: Identical to lsif_dumps.root; The working directory of the indexer image relative to the repository root.

**id**: The row ID of the search result.

**label**: The label string of the result, e.g. a one-line function signature. See protocol/documentation.go:Documentation

**label_reverse_tsv**: Indexed tsvector for the reverse of the label field, for suffix lexeme/word matching. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.

**label_tsv**: Indexed tsvector for the label field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.

**lang_name_id**: The programming language (or indexer name) that produced the result. Foreign key into lsif_data_docs_search_lang_names_public.

**path_id**: The fully qualified documentation page path ID, e.g. including "#section". See GraphQL codeintel.schema:documentationPage for what this is.

**repo_id**: The repo ID, from the main app DB repo table. Used to search over a select set of repos by ID.

**repo_name_id**: The repository name that produced the result. Foreign key into lsif_data_docs_search_repo_names_public.

**search_key**: The search key generated by the indexer, e.g. mux.Router.ServeHTTP. It is language-specific, and likely unique within a repository (but not always.) See protocol/documentation.go:Documentation.SearchKey

**search_key_reverse_tsv**: Indexed tsvector for the reverse of the search_key field, for suffix lexeme/word matching. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.

**search_key_tsv**: Indexed tsvector for the search_key field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.

**tags_id**: The tags from the documentation node. Foreign key into lsif_data_docs_search_tags_public.

# Table "public.lsif_data_docs_search_repo_names_private"
```
   Column    |   Type   | Collation | Nullable |                               Default                                
-------------+----------+-----------+----------+----------------------------------------------------------------------
 id          | bigint   |           | not null | nextval('lsif_data_docs_search_repo_names_private_id_seq'::regclass)
 repo_name   | text     |           | not null | 
 tsv         | tsvector |           | not null | 
 reverse_tsv | tsvector |           | not null | 
Indexes:
    "lsif_data_docs_search_repo_names_private_pkey" PRIMARY KEY, btree (id)
    "lsif_data_docs_search_repo_names_private_repo_name_key" UNIQUE CONSTRAINT, btree (repo_name)
    "lsif_data_docs_search_repo_names_private_reverse_tsv_idx" gin (reverse_tsv)
    "lsif_data_docs_search_repo_names_private_tsv_idx" gin (tsv)
Referenced by:
    TABLE "lsif_data_docs_search_private" CONSTRAINT "lsif_data_docs_search_private_repo_name_id_fk" FOREIGN KEY (repo_name_id) REFERENCES lsif_data_docs_search_repo_names_private(id)

```

Each unique repository name being stored in the API docs search index.

**id**: The ID of the repository name.

**repo_name**: The fully qualified name of the repository.

**reverse_tsv**: Indexed tsvector for the reverse of the lang_name field, for suffix lexeme/word matching. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.

**tsv**: Indexed tsvector for the lang_name field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.

# Table "public.lsif_data_docs_search_repo_names_public"
```
   Column    |   Type   | Collation | Nullable |                               Default                               
-------------+----------+-----------+----------+---------------------------------------------------------------------
 id          | bigint   |           | not null | nextval('lsif_data_docs_search_repo_names_public_id_seq'::regclass)
 repo_name   | text     |           | not null | 
 tsv         | tsvector |           | not null | 
 reverse_tsv | tsvector |           | not null | 
Indexes:
    "lsif_data_docs_search_repo_names_public_pkey" PRIMARY KEY, btree (id)
    "lsif_data_docs_search_repo_names_public_repo_name_key" UNIQUE CONSTRAINT, btree (repo_name)
    "lsif_data_docs_search_repo_names_public_reverse_tsv_idx" gin (reverse_tsv)
    "lsif_data_docs_search_repo_names_public_tsv_idx" gin (tsv)
Referenced by:
    TABLE "lsif_data_docs_search_public" CONSTRAINT "lsif_data_docs_search_public_repo_name_id_fk" FOREIGN KEY (repo_name_id) REFERENCES lsif_data_docs_search_repo_names_public(id)

```

Each unique repository name being stored in the API docs search index.

**id**: The ID of the repository name.

**repo_name**: The fully qualified name of the repository.

**reverse_tsv**: Indexed tsvector for the reverse of the lang_name field, for suffix lexeme/word matching. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.

**tsv**: Indexed tsvector for the lang_name field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.

# Table "public.lsif_data_docs_search_tags_private"
```
 Column |   Type   | Collation | Nullable |                            Default                             
--------+----------+-----------+----------+----------------------------------------------------------------
 id     | bigint   |           | not null | nextval('lsif_data_docs_search_tags_private_id_seq'::regclass)
 tags   | text     |           | not null | 
 tsv    | tsvector |           | not null | 
Indexes:
    "lsif_data_docs_search_tags_private_pkey" PRIMARY KEY, btree (id)
    "lsif_data_docs_search_tags_private_tags_key" UNIQUE CONSTRAINT, btree (tags)
    "lsif_data_docs_search_tags_private_tsv_idx" gin (tsv)
Referenced by:
    TABLE "lsif_data_docs_search_private" CONSTRAINT "lsif_data_docs_search_private_tags_id_fk" FOREIGN KEY (tags_id) REFERENCES lsif_data_docs_search_tags_private(id)

```

Each uniques sequence of space-separated tags being stored in the API docs search index.

**id**: The ID of the tags.

**tags**: The full sequence of space-separated tags. See protocol/documentation.go:Documentation

**tsv**: Indexed tsvector for the tags field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.

# Table "public.lsif_data_docs_search_tags_public"
```
 Column |   Type   | Collation | Nullable |                            Default                            
--------+----------+-----------+----------+---------------------------------------------------------------
 id     | bigint   |           | not null | nextval('lsif_data_docs_search_tags_public_id_seq'::regclass)
 tags   | text     |           | not null | 
 tsv    | tsvector |           | not null | 
Indexes:
    "lsif_data_docs_search_tags_public_pkey" PRIMARY KEY, btree (id)
    "lsif_data_docs_search_tags_public_tags_key" UNIQUE CONSTRAINT, btree (tags)
    "lsif_data_docs_search_tags_public_tsv_idx" gin (tsv)
Referenced by:
    TABLE "lsif_data_docs_search_public" CONSTRAINT "lsif_data_docs_search_public_tags_id_fk" FOREIGN KEY (tags_id) REFERENCES lsif_data_docs_search_tags_public(id)

```

Each uniques sequence of space-separated tags being stored in the API docs search index.

**id**: The ID of the tags.

**tags**: The full sequence of space-separated tags. See protocol/documentation.go:Documentation

**tsv**: Indexed tsvector for the tags field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.

# Table "public.lsif_data_documentation_mappings"
```
  Column   |  Type   | Collation | Nullable | Default 
-----------+---------+-----------+----------+---------
 dump_id   | integer |           | not null | 
 path_id   | text    |           | not null | 
 result_id | integer |           | not null | 
 file_path | text    |           |          | 
Indexes:
    "lsif_data_documentation_mappings_pkey" PRIMARY KEY, btree (dump_id, path_id)
    "lsif_data_documentation_mappings_inverse_unique_idx" UNIQUE, btree (dump_id, result_id)

```

Maps documentation path IDs to their corresponding integral documentationResult vertex IDs, which are unique within a dump.

**dump_id**: The identifier of the associated dump in the lsif_uploads table (state=completed).

**file_path**: The document file path for the documentationResult, if any. e.g. the path to the file where the symbol described by this documentationResult is located, if it is a symbol.

**path_id**: The documentation page path ID, see see GraphQL codeintel.schema:documentationPage for what this is.

**result_id**: The documentationResult vertex ID.

# Table "public.lsif_data_documentation_pages"
```
     Column     |  Type   | Collation | Nullable | Default 
----------------+---------+-----------+----------+---------
 dump_id        | integer |           | not null | 
 path_id        | text    |           | not null | 
 data           | bytea   |           |          | 
 search_indexed | boolean |           |          | false
Indexes:
    "lsif_data_documentation_pages_pkey" PRIMARY KEY, btree (dump_id, path_id)
    "lsif_data_documentation_pages_dump_id_unindexed" btree (dump_id) WHERE NOT search_indexed
Triggers:
    lsif_data_documentation_pages_delete AFTER DELETE ON lsif_data_documentation_pages REFERENCING OLD TABLE AS oldtbl FOR EACH STATEMENT EXECUTE FUNCTION lsif_data_documentation_pages_delete()
    lsif_data_documentation_pages_insert AFTER INSERT ON lsif_data_documentation_pages REFERENCING NEW TABLE AS newtbl FOR EACH STATEMENT EXECUTE FUNCTION lsif_data_documentation_pages_insert()
    lsif_data_documentation_pages_update AFTER UPDATE ON lsif_data_documentation_pages REFERENCING OLD TABLE AS oldtbl NEW TABLE AS newtbl FOR EACH STATEMENT EXECUTE FUNCTION lsif_data_documentation_pages_update()

```

Associates documentation pathIDs to their documentation page hierarchy chunk.

**data**: A gob-encoded payload conforming to a `type DocumentationPageData struct` pointer (lib/codeintel/semantic/types.go)

**dump_id**: The identifier of the associated dump in the lsif_uploads table (state=completed).

**path_id**: The documentation page path ID, see see GraphQL codeintel.schema:documentationPage for what this is.

# Table "public.lsif_data_documentation_path_info"
```
 Column  |  Type   | Collation | Nullable | Default 
---------+---------+-----------+----------+---------
 dump_id | integer |           | not null | 
 path_id | text    |           | not null | 
 data    | bytea   |           |          | 
Indexes:
    "lsif_data_documentation_path_info_pkey" PRIMARY KEY, btree (dump_id, path_id)

```

Associates documentation page pathIDs to information about what is at that pathID, its immediate children, etc.

**data**: A gob-encoded payload conforming to a `type DocumentationPathInoData struct` pointer (lib/codeintel/semantic/types.go)

**dump_id**: The identifier of the associated dump in the lsif_uploads table (state=completed).

**path_id**: The documentation page path ID, see see GraphQL codeintel.schema:documentationPage for what this is.

# Table "public.lsif_data_documents"
```
     Column      |  Type   | Collation | Nullable | Default 
-----------------+---------+-----------+----------+---------
 dump_id         | integer |           | not null | 
 path            | text    |           | not null | 
 data            | bytea   |           |          | 
 schema_version  | integer |           | not null | 
 num_diagnostics | integer |           | not null | 
 ranges          | bytea   |           |          | 
 hovers          | bytea   |           |          | 
 monikers        | bytea   |           |          | 
 packages        | bytea   |           |          | 
 diagnostics     | bytea   |           |          | 
Indexes:
    "lsif_data_documents_pkey" PRIMARY KEY, btree (dump_id, path)
    "lsif_data_documents_dump_id_schema_version" btree (dump_id, schema_version)
Triggers:
    lsif_data_documents_schema_versions_insert AFTER INSERT ON lsif_data_documents REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_lsif_data_documents_schema_versions_insert()

```

Stores reference, hover text, moniker, and diagnostic data about a particular text document witin a dump.

**data**: A gob-encoded payload conforming to the [DocumentData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L13:6) type. This field is being migrated across ranges, hovers, monikers, packages, and diagnostics columns and will be removed in a future release of Sourcegraph.

**diagnostics**: A gob-encoded payload conforming to the [Diagnostics](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L18:2) field of the DocumentDatatype.

**dump_id**: The identifier of the associated dump in the lsif_uploads table (state=completed).

**hovers**: A gob-encoded payload conforming to the [HoversResults](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L15:2) field of the DocumentDatatype.

**monikers**: A gob-encoded payload conforming to the [Monikers](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L16:2) field of the DocumentDatatype.

**num_diagnostics**: The number of diagnostics stored in the data field.

**packages**: A gob-encoded payload conforming to the [PackageInformation](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L17:2) field of the DocumentDatatype.

**path**: The path of the text document relative to the associated dump root.

**ranges**: A gob-encoded payload conforming to the [Ranges](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L14:2) field of the DocumentDatatype.

**schema_version**: The schema version of this row - used to determine presence and encoding of data.

# Table "public.lsif_data_documents_schema_versions"
```
       Column       |  Type   | Collation | Nullable | Default 
--------------------+---------+-----------+----------+---------
 dump_id            | integer |           | not null | 
 min_schema_version | integer |           |          | 
 max_schema_version | integer |           |          | 
Indexes:
    "lsif_data_documents_schema_versions_pkey" PRIMARY KEY, btree (dump_id)
    "lsif_data_documents_schema_versions_dump_id_schema_version_boun" btree (dump_id, min_schema_version, max_schema_version)

```

Tracks the range of schema_versions for each upload in the lsif_data_documents table.

**dump_id**: The identifier of the associated dump in the lsif_uploads table.

**max_schema_version**: An upper-bound on the `lsif_data_documents.schema_version` where `lsif_data_documents.dump_id = dump_id`.

**min_schema_version**: A lower-bound on the `lsif_data_documents.schema_version` where `lsif_data_documents.dump_id = dump_id`.

# Table "public.lsif_data_metadata"
```
      Column       |  Type   | Collation | Nullable | Default 
-------------------+---------+-----------+----------+---------
 dump_id           | integer |           | not null | 
 num_result_chunks | integer |           |          | 
Indexes:
    "lsif_data_metadata_pkey" PRIMARY KEY, btree (dump_id)

```

Stores the number of result chunks associated with a dump.

**dump_id**: The identifier of the associated dump in the lsif_uploads table (state=completed).

**num_result_chunks**: A bound of populated indexes in the lsif_data_result_chunks table for the associated dump. This value is used to hash identifiers into the result chunk index to which they belong.

# Table "public.lsif_data_references"
```
     Column     |  Type   | Collation | Nullable | Default 
----------------+---------+-----------+----------+---------
 dump_id        | integer |           | not null | 
 scheme         | text    |           | not null | 
 identifier     | text    |           | not null | 
 data           | bytea   |           |          | 
 schema_version | integer |           | not null | 
 num_locations  | integer |           | not null | 
Indexes:
    "lsif_data_references_pkey" PRIMARY KEY, btree (dump_id, scheme, identifier)
    "lsif_data_references_dump_id_schema_version" btree (dump_id, schema_version)
Triggers:
    lsif_data_references_schema_versions_insert AFTER INSERT ON lsif_data_references REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_lsif_data_references_schema_versions_insert()

```

Associates (document, range) pairs with the export monikers attached to the range.

**data**: A gob-encoded payload conforming to an array of [LocationData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L106:6) types.

**dump_id**: The identifier of the associated dump in the lsif_uploads table (state=completed).

**identifier**: The moniker identifier.

**num_locations**: The number of locations stored in the data field.

**schema_version**: The schema version of this row - used to determine presence and encoding of data.

**scheme**: The moniker scheme.

# Table "public.lsif_data_references_schema_versions"
```
       Column       |  Type   | Collation | Nullable | Default 
--------------------+---------+-----------+----------+---------
 dump_id            | integer |           | not null | 
 min_schema_version | integer |           |          | 
 max_schema_version | integer |           |          | 
Indexes:
    "lsif_data_references_schema_versions_pkey" PRIMARY KEY, btree (dump_id)
    "lsif_data_references_schema_versions_dump_id_schema_version_bou" btree (dump_id, min_schema_version, max_schema_version)

```

Tracks the range of schema_versions for each upload in the lsif_data_references table.

**dump_id**: The identifier of the associated dump in the lsif_uploads table.

**max_schema_version**: An upper-bound on the `lsif_data_references.schema_version` where `lsif_data_references.dump_id = dump_id`.

**min_schema_version**: A lower-bound on the `lsif_data_references.schema_version` where `lsif_data_references.dump_id = dump_id`.

# Table "public.lsif_data_result_chunks"
```
 Column  |  Type   | Collation | Nullable | Default 
---------+---------+-----------+----------+---------
 dump_id | integer |           | not null | 
 idx     | integer |           | not null | 
 data    | bytea   |           |          | 
Indexes:
    "lsif_data_result_chunks_pkey" PRIMARY KEY, btree (dump_id, idx)

```

Associates result set identifiers with the (document path, range identifier) pairs that compose the set.

**data**: A gob-encoded payload conforming to the [ResultChunkData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L76:6) type.

**dump_id**: The identifier of the associated dump in the lsif_uploads table (state=completed).

**idx**: The unique result chunk index within the associated dump. Every result set identifier present should hash to this index (modulo lsif_data_metadata.num_result_chunks).

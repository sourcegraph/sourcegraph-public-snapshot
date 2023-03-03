---------------------------------------------------------
-- PACKAGE REPO FILTERS
---------------------------------------------------------

CREATE TABLE IF NOT EXISTS package_repo_filters (
    id SERIAL PRIMARY KEY NOT NULL,
    behaviour TEXT NOT NULL,
    scheme TEXT NOT NULL,
    matcher JSONB NOT NULL,
    internal_regex TEXT NOT NULL,
    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT statement_timestamp()
);

CREATE OR REPLACE FUNCTION func_package_repo_filters_globtoregex() RETURNS TRIGGER AS $$
BEGIN
    IF NOT(NEW.matcher ? 'VersionGlob') AND NOT(NEW.matcher ? 'PackageGlob') THEN
        RAISE check_violation USING
            MESSAGE = 'new row for relation "package_repo_filters" must provide either "VersionGlob" or "PackageGlob" for column "matcher" of type JSONB',
            DETAIL = 'Failing row contains' || format('%s', NEW) || '.';
    END IF;

    IF NEW.matcher ? 'VersionGlob' AND NEW.matcher ? 'PackageGlob' THEN
        RAISE check_violation USING
            MESSAGE = 'new row for relation "package_repo_filters" must provide one, not both, of "VersionGlob" or "PackageGlob" for column "matcher" of type JSONB',
            DETAIL = 'Failing row contains' || format('%s', NEW) || '.';
    END IF;

    IF NEW.matcher ? 'VersionGlob' THEN
        IF NEW.matcher->>'VersionGlob' = '' THEN
            RAISE check_violation USING
                MESSAGE = 'new row for relation "package_repo_filters" must provide non-empty value for "VersionGlob" for column "matcher" of type JSONB',
                DETAIL = 'Failing row contains' || format('%s', NEW) || '.';
        END IF;
        NEW.internal_regex := glob_to_regex(NEW.matcher->>'VersionGlob');
    ELSE
        IF NEW.matcher->>'PackageGlob' = '' THEN
            RAISE check_violation USING
                MESSAGE = 'new row for relation "package_repo_filters" must provide non-empty value for "PackageGlob" for column "matcher" of type JSONB',
                DETAIL = 'Failing row contains' || format('%s', NEW) || '.';
        END IF;
        NEW.internal_regex := glob_to_regex(NEW.matcher->>'PackageGlob');
    END IF;
    RETURN NEW;
END $$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_package_repo_filters_globtoregex ON package_repo_filters;
CREATE TRIGGER trigger_package_repo_filters_insert_globtoregex
BEFORE INSERT ON package_repo_filters
FOR EACH ROW
EXECUTE PROCEDURE func_package_repo_filters_globtoregex();

DROP TRIGGER IF EXISTS trigger_package_repo_filters_update_globtoregex ON package_repo_filters;
CREATE TRIGGER trigger_package_repo_filters_update_globtoregex
BEFORE UPDATE ON package_repo_filters
FOR EACH ROW
WHEN (OLD.matcher <> NEW.matcher)
EXECUTE PROCEDURE func_package_repo_filters_globtoregex();

CREATE OR REPLACE FUNCTION func_package_repo_filters_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = statement_timestamp();
    RETURN NEW;
END $$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_package_repo_filters_updated_at ON package_repo_filters;
CREATE TRIGGER trigger_package_repo_filters_updated_at
BEFORE UPDATE ON package_repo_filters
FOR EACH ROW
WHEN (OLD.* IS DISTINCT FROM NEW.*)
EXECUTE PROCEDURE func_package_repo_filters_updated_at();

-- because creating types is unnecessarily awkward with idempotency
ALTER TABLE package_repo_filters
    DROP CONSTRAINT IF EXISTS package_repo_filters_is_pkgrepo_scheme,
    ADD CONSTRAINT package_repo_filters_is_pkgrepo_scheme CHECK (
        scheme = ANY('{"semanticdb","npm","go","python","rust-analyzer","scip-ruby"}')
    );

ALTER TABLE package_repo_filters
    DROP CONSTRAINT IF EXISTS package_repo_filters_behaviour_is_allow_or_block,
    ADD CONSTRAINT package_repo_filters_behaviour_is_allow_or_block CHECK (
        behaviour = ANY('{"BLOCK","ALLOW"}')
    );

CREATE UNIQUE INDEX IF NOT EXISTS package_repo_filters_unique_matcher_per_scheme
ON package_repo_filters (scheme, matcher);

---------------------------------------------------------
-- PACKAGE REPOS & VERSIONS BLOCK AND LAST CHECK DATES
---------------------------------------------------------

ALTER TABLE lsif_dependency_repos
ADD COLUMN IF NOT EXISTS blocked BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS last_checked_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS lsif_dependency_repos_blocked
ON lsif_dependency_repos USING btree (blocked);

CREATE INDEX IF NOT EXISTS lsif_dependency_repos_last_checked_at
ON lsif_dependency_repos USING btree (last_checked_at);

ALTER TABLE package_repo_versions
ADD COLUMN IF NOT EXISTS blocked BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS last_checked_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS package_repo_versions_blocked
ON package_repo_versions USING btree (blocked);

CREATE INDEX IF NOT EXISTS package_repo_versions_last_checked_at
ON package_repo_versions USING btree (last_checked_at);

---------------------------------------------------------
-- BIG BAG O' FUNCTIONS
---------------------------------------------------------

CREATE OR REPLACE FUNCTION is_unversioned_package_allowed(package text, pkgscheme text) RETURNS boolean AS $$
DECLARE
    blocked boolean;
    allowed boolean;
BEGIN
    blocked := (
        SELECT COALESCE(bool_or(m.matches), FALSE)
        FROM (
            SELECT TRUE AS matches
            FROM package_repo_filters f
            WHERE f.behaviour = 'BLOCK'
            AND f.scheme = pkgscheme
            AND (
                (
                    f.matcher ? 'PackageGlob'
                    AND package ~ f.internal_regex
                ) OR (
                    -- non-all-encompassing version globs don't apply to unversioned packages,
		            -- likely we're at too-early point in the syncing process to know, but also
		            -- we may still want the package to browse versions that _dont_ match this
                    f.matcher->>'PackageName' = package
                    AND f.matcher->>'VersionGlob' = '*'
                )
            )
        ) as m
    );

    -- blacklist takes priority, can't poke holes out
    IF blocked = TRUE THEN
        RETURN FALSE;
    END IF;

    -- default allow if no allowlist and not blocked so far
    allowed := (
        SELECT COUNT(*) = 0
        FROM package_repo_filters f
        WHERE f.behaviour = 'ALLOW'
        AND f.scheme = pkgscheme
    );

    allowed := allowed OR (
        SELECT COALESCE(bool_or(m.matches), FALSE)
        FROM (
            SELECT TRUE AS matches
            FROM package_repo_filters f
            WHERE f.behaviour = 'ALLOW'
            AND f.scheme = pkgscheme
            AND (
                    (
                    f.matcher ? 'PackageGlob'
                    AND package ~ f.internal_regex
                ) OR (
                    f.matcher->>'PackageName' = package
                    AND f.matcher->>'VersionGlob' = '*'
                )
            )
        ) AS m
    );

    RETURN allowed;
END;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION is_versioned_package_allowed(package text, version text, pkgscheme text) RETURNS boolean AS $$
DECLARE
    blocked boolean;
    allowed boolean;
BEGIN
    blocked := (
        SELECT COALESCE(bool_or(m.matches), FALSE)
        FROM (
            SELECT TRUE AS matches
            FROM package_repo_filters f
            WHERE f.behaviour = 'BLOCK'
            AND f.scheme = 'rust-analyzer'
            AND
            (
                (
                    f.matcher ? 'PackageGlob'
                    AND package ~ f.internal_regex
                ) OR (
                    f.matcher->>'PackageName' = package
                    AND version ~ f.internal_regex
                )
            )
        ) as m
    );

    -- blacklist takes priority, can't poke holes out
    IF blocked = TRUE THEN
        RETURN FALSE;
    END IF;

    -- default allow if no allowlist and not blocked so far
    allowed := (
        SELECT COUNT(*) = 0
        FROM package_repo_filters f
        WHERE behaviour = 'ALLOW'
        AND f.scheme = pkgscheme
    );

    allowed := allowed OR (
        SELECT COALESCE(bool_or(m.matches), FALSE)
        FROM (
            SELECT TRUE AS matches
            FROM package_repo_filters f
            WHERE f.behaviour = 'ALLOW'
            AND f.scheme = pkgscheme
            AND
            (
                (
                    f.matcher ? 'PackageGlob'
                    AND package ~ f.internal_regex
                ) OR (
                    f.matcher->>'PackageName' = package
                    AND version ~ f.internal_regex
                )
            )
        ) AS m
    );

    RETURN allowed;
END;
$$ LANGUAGE plpgsql;

-- Transformed from Python stdlib fnmatch.translate function
-- as licensed by https://sourcegraph.com/github.com/python/cpython@54dfa14c5a94b893b67a4d9e9e403ff538ce9023/-/blob/LICENSE
CREATE OR REPLACE FUNCTION glob_to_regex(pat text) RETURNS text AS $$
DECLARE
    res text[] := array[]::text[];
    i int := 0;
    n int := length(pat) + 1;
BEGIN
    WHILE i < n LOOP
        DECLARE
            c text := substring(pat, i + 1, 1);
        BEGIN
            i := i + 1;
            IF c = '*' THEN
                -- compress consecutive `*` into one
                IF (array_length(res, 1) IS NULL OR res[array_length(res, 1)] IS NOT NULL) THEN
                    res := array_append(res, NULL);
                END IF;
            ELSIF c = '?' THEN
                res := array_append(res, '.');
            ELSIF c = '[' THEN
                DECLARE
                    j int := i;
                BEGIN
                    IF substring(pat, j + 1, 1) = '!' THEN
                        j := j + 1;
                    END IF;
                    IF substring(pat, j + 1, 1) = ']' THEN
                        j := j + 1;
                    END IF;
                    WHILE substring(pat, j + 1, 1) <> ']' LOOP
                        j := j + 1;
                    END LOOP;
                    IF j >= n THEN
                        res := array_append(res, '\[');
                    ELSE
                        DECLARE
                            stuff text := substring(pat, i, j - i + 1);
                        BEGIN
                            IF position('-' IN stuff) = 0 THEN
                                stuff := replace(stuff, '\\', '\\\\');
                            ELSE
                                DECLARE
                                    chunks text[] := '{}';
                                    k int := i + 2;
                                BEGIN
                                    IF substring(pat, i + 1, 1) = '!' THEN
                                        k := k + 1;
                                    END IF;
                                    WHILE TRUE LOOP
                                        k := position('-' in substring(pat, k, j - k + 1));
                                        IF k = 0 THEN
                                            EXIT;
                                        END IF;
                                        chunks := array_append(chunks, substring(pat, i, k - i));
                                        i := k + 1;
                                        k := k + 3;
                                    END LOOP;
                                    DECLARE
                                        chunk text := substring(pat, i, j - i + 1);
                                    BEGIN
                                        IF chunk <> '' THEN
                                            chunks := array_append(chunks, chunk);
                                        ELSE
                                            chunks[array_length(chunks, 1)] := chunks[array_length(chunks, 1)] || '-';
                                        END IF;
                                        -- Remove empty ranges -- invalid in RE.
                                        FOR k IN REVERSE 1 .. array_length(chunks, 1) - 1 LOOP
                                            IF substring(chunks[k - 1], array_length(chunks[k - 1], 1), 1) > substring(chunks[k], 1, 1) THEN
                                                chunks[k - 1] := substring(chunks[k - 1], 1, array_length(chunks[k - 1], 1) - 1) || substring(chunks[k], 2);
                                                chunks := array_remove(chunks, k);
                                            END IF;
                                        END LOOP;
                                        -- Escape backslashes and hyphens for set difference (--).
                                        -- Hyphens that create ranges shouldn't be escaped.
                                        stuff := array_to_string(chunks, '-', true);
                                    END;
                                END;
                            END IF;
                            -- Escape set operations (&&, ~~ and ||).
                            stuff := replace(stuff, '([&~|])', '\\\\\1');
                            i := j + 1;
                            IF stuff = '' THEN
                                -- Empty range: never match.
                                res := array_append(res, '(?!)');
                            ELSIF stuff = '!' THEN
                                -- Negated empty range: match any character.
                                res := array_append(res, '.');
                            ELSE
                                IF substring(stuff, 1, 1) = '!' THEN
                                    stuff := '^' || substring(stuff, 2);
                                ELSIF substring(stuff, 1, 1) IN ('^', '[') THEN
                                    stuff := '\\' || stuff;
                                END IF;
                                res := array_append(res, format('[%s]', stuff));
                            END IF;
                        END;
                    END IF;
                END;
            ELSE
                -- regex escape
                res := array_append(res, replace(replace(replace(replace(replace(
                replace(replace(replace(replace(replace(
                replace(replace(replace(replace(replace(
                replace(replace(replace(replace(replace(
                replace(replace(replace(replace(c,
                '(', '\('), ')', '\)'), '[', '\['), ']', '\]'), '{', '\{'),
                '}', '\}'), '?', '\?'), '*', '\*'), '+', '\+'), '-', '\-'),
                '|', '\|'), '^', '\^'), '$', '\$'), '\', '\\'), '.', '\.'),
                '&', '\&'), '~', '\~'), '#', '\#'), ' ', '\ '), chr(9), '\t'),
                chr(10), '\n'), chr(13), '\r'), chr(11), '\x0b'), chr(12), '\x0c'));
            END IF;
        END;
    END LOOP;
    ASSERT i = n, format('%s != %s', i, n);
    -- Deal with STARs.
    DECLARE
        inp text[] := res;
        res text[] := array[]::text[];
        i int := 1;
        n int := array_length(inp, 1)+1;
    BEGIN
        -- Fixed pieces at the start?
        WHILE i < n AND inp[i] IS NOT NULL LOOP
            res := array_append(res, inp[i]);
            i := i + 1;
        END LOOP;
        -- Now deal with STAR fixed STAR fixed ...
        -- For an interior `STAR fixed` pairing, we want to do a minimal
        -- .*? match followed by `fixed`, with no possibility of backtracking.
        -- Atomic groups ("(?>...)") allow us to spell that directly.
        -- Note: people rely on the undocumented ability to join multiple
        -- translate() results together via "|" to build large regexps matching
        -- "one of many" shell patterns.
        WHILE i < n LOOP
            ASSERT inp[i] IS NULL, format('%s at %s IS NOT NULL', inp[i], i);
            i := i + 1;
            IF i = n THEN
                res := array_append(res, '.*');
                EXIT;
            END IF;
            ASSERT inp[i] IS NOT NULL, format('%s IS NULL', i);
            DECLARE
                fixed text[] := '{}';
                fixed1 text := '';
            BEGIN
                WHILE i < n AND inp[i] IS NOT NULL LOOP
                    fixed := array_append(fixed, inp[i]);
                    i := i + 1;
                END LOOP;
                fixed1 := array_to_string(fixed, '');
                IF i = n THEN
                    res := array_append(res, '.*');
                    res := array_append(res, fixed1);
                ELSE
                    res := array_append(res, format('(.*?%s)', fixed1));
                END IF;
            END;
        END LOOP;
        ASSERT i = n, format('%s != %s', i, n);
        RETURN format('^%s$', array_to_string(res, ''));
    END;
END;
$$ LANGUAGE plpgsql;

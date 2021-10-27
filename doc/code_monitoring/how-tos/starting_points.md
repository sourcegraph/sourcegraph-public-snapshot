# Start monitoring your code

This page lists code monitors that are commonly used and can be used across most codebases.

## Watch for potential secrets

Code monitors can help you react quickly to the secrets being accidentally committed to your codebase.

```
patterntype:regexp ("[a-z0-9+/]{32,}=?"|'[a-z0-9+/]{32,}=?'|`[a-z0-9+/]{32,}=?`) or -----BEGIN RSA PRIVATE KEY----- or token.+[a-z0-9+/]{32,}=['"]?\n
```

The above query will return matches for:
* Alphanumeric or base64-encoded strings longer than 32 characters
* RSA private key headers
* Patterns suggesting the association of “token” with a base64-encoded value on the same line

## Watch for consumers of deprecated endpoints

```
f:\.tsx?$ patterntype:regexp fetch\(['"`]/deprecated-endpoint
```

If you’re deprecating an API or an endpoint, you may find it useful to set up a code monitor watching for new consumers. As an example, the above query will surface fetch() calls to `/deprecated-endpoint` within TypeScript files. Replace `/deprecated-endpoint` with the actual path of the endpoint being deprecated.

## Get notified when a file changes

```
patterntype:regexp repo:^github\.com/sourcegraph/sourcegraph$ file:SourcegraphWebApp\.tsx$ type:diff
```

You may want to get notified when a given file is changed, regardless of the diff contents of the change: the above query will return all changes to the `SourcegraphWebApp.tsx` file on the `github.com/sourcegraph/sourcegraph` repo.

## Get notified when a specific function call is added

```
repo:^github\.com/sourcegraph/sourcegraph$ type:diff select:commit.diff.added Sprintf
```

You may want to monitor new additions of a specific function call, for example a deprecated function or a function that introduces a security concern.  This query will notify you whenever a new addition of `Sprintf` is added to the `sourcegraph/sourcegraph` repository.  This query selects all diff additions marked as "+".  If a call of `Sprintf` is both added and removed from a file, this query will still notify due to the addition.

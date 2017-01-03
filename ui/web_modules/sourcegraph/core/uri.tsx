import URI from "vs/base/common/uri";

// A URI in Sourcegraph refers to a (file) path and revision in a
// repository. For example:
//
//   git://github.com/facebook/react?commitid#file
//
// Use pathInRepo to generate this URI, and use repoParams to
// extract the repo, rev, and path parameters from it.
export class URIUtils {
	// pathInRepo returns a URI to a file at a specific (optional)
	// revision in a Git repository. It is a Sourcegraph-specific
	// convention.
	static pathInRepo(repo: string, rev: string | null, path: string): URI {
		if (!rev) {
			rev = "HEAD";
		}
		return URI.parse(`git:\/\/${repo}`).with({
			fragment: path.replace(/^\//, ""),
			query: rev ? encodeURIComponent(rev) : "",
		});
	}

	// repoParams extracts the repo, rev, and path parameters
	static repoParams(uri: URI): { repo: string, rev: string | null, path: string } {
		if (uri.scheme !== "git") {
			throw new Error(`expected git URI scheme in ${uri.toString()}`);
		}
		return {
			repo: `${uri.authority}${uri.path.replace(/\.git$/, "")}`,
			rev: decodeURIComponent(uri.query),
			path: uri.fragment.replace(/^\//, ""),
		};
	}

	// repoParamsExt mirrors the functionality of repoParams, but is
	// meant to be called outside of Monaco (or when Monaco has not
	// loaded).
	static repoParamsExt(uri: string): { repo: string, rev: string | null, path: string } {
		let a = document.createElement("a");
		uri = uri.replace(/^git/, "http");
		a.href = uri;
		return {
			repo: `${a.hostname}/${a.pathname.replace(/\.git$/, "")}`,
			rev: decodeURIComponent(a.search.replace(/^\?/, "")),
			path: a.hash.replace(/^\#/, "").replace(/^\//, ""),
		};
	}

	// withoutFilePath returns the URI without the file path (the URL
	// fragment).
	static withoutFilePath(uri: URI): URI {
		if (!uri.fragment) {
			return uri;
		}
		return URI.from({
			scheme: uri.scheme,
			authority: uri.authority,
			path: uri.path,
			query: uri.query,
			// Omit the fragment (file path).
		});
	}

}

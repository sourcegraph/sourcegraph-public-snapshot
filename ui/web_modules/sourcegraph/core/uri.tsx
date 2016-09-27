// tslint:disable no-namespace

// A URI in Sourcegraph refers to a (file) path and revision in a
// repository. For example:
//
//   git://github.com/facebook/react?commitid#file
//
// Use pathInRepo to generate this URI, and use repoParams to
// extract the repo, rev, and path parameters from it.
export class URI {
	// Implementation note: This is a class so that callers refer to
	// it as "URI.whatever", which makes it easier to scan. We aren't
	// able to make it inherit from monaco.Uri because we load Monaco
	// asynchronously, but if that changes in the future, we can make
	// this a true subclass of monaco.Uri.

	// pathInRepo returns a URI to a file at a specific (optional)
	// revision in a Git repository. It is a Sourcegraph-specific
	// convention.
	static pathInRepo(repo: string, rev: string | null, path: string): monaco.Uri {
		return monaco.Uri.parse(`git:\/\/${repo}`).with({
			fragment: path.replace(/^\//, ""),
			query: rev ? encodeURIComponent(rev) : "",
		});
	}

	// repoParams extracts the repo, rev, and path parameters
	static repoParams(uri: monaco.Uri): {repo: string, rev: string | null, path: string} {
		uri = this.fromRefsDisplayURIMaybe(uri);
		if (uri.scheme !== "git") {
			throw new Error(`expected git URI scheme in ${uri.toString()}`);
		}
		return {
			repo: `${uri.authority}${uri.path.replace(/\.git$/, "")}`,
			rev: decodeURIComponent(uri.query),
			path: uri.fragment.replace(/^\//, ""),
		};
	}

	// withoutFilePath returns the URI without the file path (the URL
	// fragment).
	static withoutFilePath(uri: monaco.Uri): monaco.Uri {
		if (!uri.fragment) {
			return uri;
		}
		return monaco.Uri.from({
			scheme: uri.scheme,
			authority: uri.authority,
			path: uri.path,
			query: uri.query,
			// Omit the fragment (file path).
		});
	}

	// toRefsDisplayURI converts a URI into a form in which it will be displayed
	// nicely in the Monaco references sidebar. This is a kludge, because the
	// display logic for that component is not easily overridable and the only
	// other alternative would be to change our own URI scheme for files.
	static toRefsDisplayURI(serverURI: monaco.Uri): monaco.Uri {
		return monaco.Uri.parse(`${serverURI.scheme}:\/\/${serverURI.authority}${serverURI.path}\/%20${serverURI.fragment}?q=${serverURI.query}&refsDisp=t`);
	}

	// fromDisplayURI converts a URI back from the Monaco-find-references form
	// to its standard form. If the URI is not in that form to begin with, this
	// just returns the original URI.
	static fromRefsDisplayURIMaybe(dispURI: monaco.Uri): monaco.Uri {
		if (dispURI.query.indexOf("refsDisp") === -1) {
			return dispURI;
		}
		let query = new (global as any).URLSearchParams(dispURI.query);
		let rev = query.get("q") ? query.get("q") : "";
		let sepIdx = dispURI.path.indexOf("/ ");
		let repo = dispURI.path.slice(0, sepIdx);
		let filepath = dispURI.path.slice(sepIdx + 2);
		return monaco.Uri.from({
			scheme: dispURI.scheme,
			authority: dispURI.authority,
			path: repo,
			query: rev,
			fragment: filepath,
		});
	}
}

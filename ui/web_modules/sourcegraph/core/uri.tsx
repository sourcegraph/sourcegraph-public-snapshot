import URI from "vs/base/common/uri";

export class URIUtils {

	/**
	 * tryConvertGitToFileURI converts a git-scheme or zap-scheme
	 * URI to the equivalent file-scheme URI. For non git/zap-scheme
	 * URIs, it is the identity function.
	 */
	static tryConvertGitToFileURI(uri: URI): URI {
		if (uri.scheme !== "git" && uri.scheme !== "zap") {
			return uri;
		}
		return URI.parse(`file://${uri.authority}${uri.path}${uri.fragment ? `/${uri.fragment}` : ""}`);
	}

	/**
	 * createResourceURI returns a file-scheme URI for a
	 * workspace or document. A workspace is identified by a
	 * repo; a document by a repo *and* path.
	 */
	static createResourceURI(repo: string, path?: string): URI {
		if (path && path !== "/") {
			return URI.parse(`file://${repo}/${path}`);
		}
		return URI.parse(`file://${repo}`);
	}
}

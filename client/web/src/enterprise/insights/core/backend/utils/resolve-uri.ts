/**
 * A resolved URI without an identified path.
 */
export interface ResolvedRootURI {
    repo: string
    rev: string
}

/**
 * A resolved URI with an identified path in a repository at a specific revision.
 */
export interface ResolvedDocumentURI extends ResolvedRootURI {
    path: string
}

/**
 * Resolve a URI of the form git://github.com/owner/repo?rev to an absolute reference.
 */
export function resolveRootURI(uri: URL): ResolvedRootURI {
    if (uri.protocol !== 'git:') {
        throw new Error(`Unsupported protocol: ${uri.protocol}`)
    }

    const repo = (uri.host + decodeURIComponent(uri.pathname)).replace(/^\/*/, '')
    const revision = decodeURIComponent(uri.search.slice(1))

    if (!revision) {
        throw new Error('Could not determine revision')
    }

    return { repo, rev: revision }
}

/**
 * Resolve a URI of the form git://github.com/owner/repo?rev#path to an absolute reference.
 */
export function resolveDocumentURI(uri: URL): ResolvedDocumentURI {
    return {
        ...resolveRootURI(uri),
        path: uri.hash.slice(1),
    }
}

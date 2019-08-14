import RelateUrl from 'relateurl'

/**
 * Options to make sure that RelateUrl only outputs relative URLs and performs not other "smart" modifications.
 */
const RELATE_URL_OPTIONS: RelateUrl.Options = {
    // Make sure RelateUrl does not prefer root-relative URLs if shorter
    output: RelateUrl.PATH_RELATIVE,
    // Make sure RelateUrl does not remove trailing slash if present
    removeRootTrailingSlash: false,
    // Make sure RelateUrl does not remove default ports
    defaultPorts: {},
}

/**
 * Like `path.relative()` but for URLs.
 * Inverse of `url.resolve()` or `new URL(relative, base)`.
 */
export const relativeUrl = (from: URL, to: URL): string => RelateUrl.relate(from.href, to.href, RELATE_URL_OPTIONS)

export const buildGitUri = ({
    repository,
    commit,
    path,
}: {
    repository: string
    commit: string
    path: string
}): URL => {
    const url = new URL(`git://${repository}`)
    url.search = commit
    url.hash = path
    return url
}

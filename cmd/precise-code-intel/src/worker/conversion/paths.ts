import * as path from 'path'
import RelateUrl from 'relateurl'

/**
 * Return the dirname of the given path. Returns empty string
 * if the path denotes a file in the current directory.
 */
export function dirnameWithoutDot(pathname: string): string {
    return path.dirname(pathname) === '.' ? '' : path.dirname(pathname)
}

/**
 * Construct a root-relative path within a dump.
 *
 * @param projectRoot The absolute URI to the dump root.
 * @param documentURI The absolute URI to the document.
 */
export function relativePath(projectRoot: URL, documentURI: URL): string {
    return RelateUrl.relate(`${projectRoot.href}/`, documentURI.href, {
        defaultPorts: {},
        output: RelateUrl.PATH_RELATIVE,
        removeRootTrailingSlash: false,
    })
}

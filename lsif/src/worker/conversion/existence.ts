import * as path from 'path'
import { TracingContext, logAndTraceCall } from '../../shared/tracing'
import { getDirectoryChildren } from '../../shared/gitserver/gitserver'
import { createSilentLogger } from '../../shared/logging'

/**
 * Determines whether or not a document path within an LSIF upload should be visible
 * within the generated dump. This allows us to prune documents which are not inside of
 * the root (which will never be queried from within this dump), and references to paths
 * that do not occur in the git tree at this commit.
 *
 * This class caches results to efficiently query gitserver for the contents of directories
 * so that it neither has to:
 *
 *   - request all files recursively at once (bad for large repos and mono repos), nor
 *   - make a request for every unique path in the index.
 */
export class PathExistenceChecker {
    private repositoryId: number
    private commit: string
    private root: string
    private frontendUrl?: string
    private ctx?: TracingContext
    private mockGetDirectoryChildren?: typeof getDirectoryChildren
    private directoryContents = new Map<string, Set<string>>()
    private numGitserverRequests = 0

    /**
     * Create a new PathExistenceChecker.
     *
     * @param args Parameter bag.
     */
    constructor({
        repositoryId,
        commit,
        root,
        frontendUrl,
        ctx,
        mockGetDirectoryChildren,
    }: {
        /** The repository identifier. */
        repositoryId: number
        /** The commit from which the gitserver queries should start. */
        commit: string
        /** The root of all files in the dump. */
        root: string
        /**  The url of the frontend internal API. */
        frontendUrl?: string
        /** The tracing context. */
        ctx?: TracingContext
        /** A mock implementation of the gitserver function. */
        mockGetDirectoryChildren?: typeof getDirectoryChildren
    }) {
        this.repositoryId = repositoryId
        this.commit = commit
        this.root = root
        this.frontendUrl = frontendUrl
        this.ctx = ctx
        this.mockGetDirectoryChildren = mockGetDirectoryChildren
    }

    /**
     * Warms the git directory cache by determining if each of the supplied paths
     * exist in git. This function batches queries to gitserver to minimize the
     * number of roundtrips during conversion.
     *
     * @param documentPaths A set of dump root-relative paths.
     */
    public warmCache(documentPaths: string[]): Promise<void> {
        return logAndTraceCall(
            this.ctx || {},
            'Warming git directory cache',
            async ({ logger = createSilentLogger() }) => {
                // TODO - batch requests. Must do this in a separate PR as the frontend
                // gitserver proxy currently only accepts a single ExecRequest payload.
                // Tracked in https://github.com/sourcegraph/sourcegraph/issues/8555.
                for (const documentPath of documentPaths) {
                    await this.isInGitTree(documentPath)
                }

                logger.debug(`Performed ${this.numGitserverRequests} gitserver requests`)
            }
        )
    }

    /**
     * Determines if the given file path should be included in the generated dump.
     *
     * @param documentPath The path of the file relative to the dump root.
     * @param requireDocumentDump Whether or not we require the path to be within the dump root.
     */
    public async shouldIncludePath(documentPath: string, requireDocumentDump: boolean = true): Promise<boolean> {
        return (await this.isInGitTree(documentPath)) && (!requireDocumentDump || !documentPath.startsWith('..'))
    }

    /**
     * Determine if the given path is known by git. If no frontend url is configured,
     * this method returns true (assumes it's in the tree).
     *
     * @param documentPath The path of the file relative to the dump root.
     */
    private async isInGitTree(documentPath: string): Promise<boolean> {
        if (!this.frontendUrl) {
            // Integration tests do not set a frontend url for conversion.
            // We early out here as we can just include everything in the
            // index.
            return true
        }

        const relativePath = path.join(this.root, documentPath)
        const dirname = dirnameWithoutDot(relativePath)
        return (await this.getChildrenFromRoot(dirname)).has(relativePath)
    }

    /**
     * Returns the set of root-relative paths of the immediate children of the
     * given directory. If no frontend url is configured or the directory is outside of
     * the repository root, this method returns an empty set.
     *
     * @param dirname The repo-root-relative directory.
     */
    private async getChildrenFromRoot(dirname: string): Promise<Set<string>> {
        // Not in git tree. Do not make a query for this. Not just because
        // it would be useless, but git ls-tree will blow up pretty hard.
        if (dirname.startsWith('..')) {
            return new Set()
        }

        for (const ancestor of properAncestors(dirname)) {
            // Calculate the children of all ancestors of this directory that are also
            // in the repo. We do this from the root down to the leaf so that we can prune
            // large chunks of untracked files with one request (e.g. a node_modules dir).
            const children = await this.getChildren(ancestor)
            if (children.size === 0) {
                // This directory doesn't exist or there are no children. Either way we can
                // early out with an empty set of children as there are no descendants.
                return new Set()
            }
        }

        return this.getChildren(dirname)
    }

    /**
     * Returns the set of root-relative paths of the immediate children of the
     * given directory. If no frontend url is configured, this method returns an empty
     * set.
     *
     * This method memoizes the results so a dump conversion will make only one request
     * to gitserver per directory.
     *
     * @param dirname The repo-root-relative directory.
     */
    private async getChildren(dirname: string): Promise<Set<string>> {
        if (!this.frontendUrl) {
            return new Set()
        }

        let children = this.directoryContents.get(dirname)
        if (children) {
            return children
        }

        children = await (this.mockGetDirectoryChildren || getDirectoryChildren)({
            frontendUrl: this.frontendUrl,
            repositoryId: this.repositoryId,
            commit: this.commit,
            dirname,
            ctx: this.ctx,
        })

        this.numGitserverRequests++
        this.directoryContents.set(dirname, children)
        return children
    }
}

/**
 * Return the dirname of the given path. Returns empty string
 * if the path denotes a file in the current directory.
 */
function dirnameWithoutDot(pathname: string): string {
    const dirname = path.dirname(pathname)
    return dirname === '.' ? '' : dirname
}

/**
 * Return all ancestor paths of the given directory.
 */
export function properAncestors(dirname: string): string[] {
    const ancestors = []
    const pathSegments = dirname.split('/')
    for (let i = 0; i < pathSegments.length; i++) {
        ancestors.push(pathSegments.slice(0, i).join('/'))
    }
    return ancestors
}

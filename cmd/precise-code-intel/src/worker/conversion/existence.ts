import * as path from 'path'
import { TracingContext, logAndTraceCall } from '../../shared/tracing'
import { getDirectoryChildren } from '../../shared/gitserver/gitserver'
import { createBatcher } from './batch'
import { dirnameWithoutDot } from './paths'

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
     * Determines if the given file path should be included in the generated dump.
     *
     * @param documentPath The path of the file relative to the dump root.
     * @param requireDocumentDump Whether or not we require the path to be within the dump root.
     */
    public shouldIncludePath(documentPath: string, requireDocumentDump: boolean = true): boolean {
        if (this.frontendUrl) {
            // Determine if the given path is known by git. Integration
            // tests do not set a frontend url for conversion, in which case
            // we consider every file to be in the index.
            const relativePath = path.join(this.root, documentPath)
            if (!this.directoryContents.get(dirnameWithoutDot(relativePath))?.has(relativePath)) {
                return false
            }
        }

        return !requireDocumentDump || !documentPath.startsWith('..')
    }

    /**
     * Warms the git directory cache by determining if each of the supplied paths exist
     * in git. This function batches queries to gitserver to minimize the number of
     * roundtrips during conversion. If no frontend url is configured, this method does
     * nothing.
     *
     * @param documentPaths A set of dump root-relative paths.
     */
    public warmCache(documentPaths: string[]): Promise<void> {
        return logAndTraceCall(this.ctx || {}, 'Warming git directory cache', async () => {
            if (!this.frontendUrl) {
                return
            }

            const batcher = createBatcher(this.root, documentPaths)
            let exists: string[] = []

            while (true) {
                const { value: batch, done } = batcher.next(exists)
                if (done || !batch) {
                    break
                }

                const childMap = await (this.mockGetDirectoryChildren || getDirectoryChildren)({
                    frontendUrl: this.frontendUrl,
                    repositoryId: this.repositoryId,
                    commit: this.commit,
                    dirnames: batch,
                    ctx: this.ctx,
                })

                for (const [dirname, children] of childMap.entries()) {
                    this.directoryContents.set(dirname, children)
                }

                exists = []
                for (const dirname of batch) {
                    const children = this.directoryContents.get(dirname)
                    if (children && children.size > 0) {
                        exists.push(dirname)
                    }
                }
            }
        })
    }
}

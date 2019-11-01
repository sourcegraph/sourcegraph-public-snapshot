import rmfr from 'rmfr'
import { Backend, ReferencePaginationCursor } from '../backend'
import { Configuration } from '../config'
import { createCleanPostgresDatabase, createStorageRoot, convertTestData } from '../test-utils'
import { XrepoDatabase } from '../xrepo'
import { lsp } from 'lsif-protocol'

// TODO
export class BackendTestContext {
    public backend?: Backend
    public xrepoDatabase?: XrepoDatabase

    private storageRoot?: string
    private cleanup?: () => Promise<void>

    // TODO
    public async init(): Promise<void> {
        this.storageRoot = await createStorageRoot()

        const { connection, cleanup } = await createCleanPostgresDatabase()
        this.cleanup = cleanup

        this.xrepoDatabase = new XrepoDatabase(this.storageRoot, connection)
        this.backend = new Backend(this.storageRoot, this.xrepoDatabase, () => ({} as Configuration))
    }

    // TODO
    // * @param updateCommits Whether not to update commits.
    public convertTestData(
        repository: string,
        commit: string,
        root: string,
        filename: string,
        updateCommits: boolean = true
    ): Promise<void> {
        if (!this.xrepoDatabase || !this.storageRoot) {
            return Promise.resolve()
        }

        return convertTestData(this.xrepoDatabase, this.storageRoot, repository, commit, root, filename, updateCommits)
    }

    // TODO
    public async teardown(): Promise<void> {
        if (this.storageRoot) {
            await rmfr(this.storageRoot)
        }

        if (this.cleanup) {
            await this.cleanup()
        }
    }
}

// TODO
export function filterNodeModules<T>({
    locations,
    cursor,
}: {
    locations: lsp.Location[]
    cursor?: ReferencePaginationCursor
}): { locations: lsp.Location[]; cursor?: ReferencePaginationCursor } {
    return { locations: locations.filter(l => !l.uri.includes('node_modules')), cursor }
}

import rmfr from 'rmfr'
import { Backend, ReferencePaginationContext } from './backend'
import { Configuration } from './config'
import { Connection } from 'typeorm'
import { convertTestData, createCleanPostgresDatabase, createCommit, createStorageRoot } from './test-utils'
import { XrepoDatabase } from './xrepo'
import { lsp } from 'lsif-protocol'

describe('Database', () => {
    let connection!: Connection
    let cleanup!: () => Promise<void>
    let storageRoot!: string
    let xrepoDatabase!: XrepoDatabase
    let backend!: Backend

    beforeAll(async () => {
        ;({ connection, cleanup } = await createCleanPostgresDatabase())
        storageRoot = await createStorageRoot('typescript')
        xrepoDatabase = new XrepoDatabase(storageRoot, connection)
        backend = new Backend(storageRoot, xrepoDatabase, () => ({} as Configuration))

        // Prepare test data
        for (const repository of ['a', 'b1', 'b2', 'b3', 'b4', 'b5', 'b6', 'b7', 'b8', 'b9', 'b10']) {
            await convertTestData(
                xrepoDatabase,
                storageRoot,
                repository,
                createCommit(repository),
                '',
                `typescript/reference-pagination/data/${repository}.lsif.gz`
            )
        }
    })

    afterAll(async () => {
        await rmfr(storageRoot)

        if (cleanup) {
            await cleanup()
        }
    })

    it('should find all defs of `add` from repo a', async () => {
        const fetch = (paginationContext?: ReferencePaginationContext) =>
            backend.references(
                'a',
                createCommit('a'),
                'src/index.ts',
                {
                    line: 0,
                    character: 17,
                },
                paginationContext
            )

        const { locations: page0, cursor: cursor0 } = await fetch() // everything
        const { locations: page1, cursor: cursor1 } = await fetch({ limit: 3 }) // a, b1, b10, b2
        const { locations: page2, cursor: cursor2 } = await fetch({ limit: 3, cursor: cursor1 }) // b3, b4, b5
        const { locations: page3, cursor: cursor3 } = await fetch({ limit: 3, cursor: cursor2 }) // b6, b7, b8
        const { locations: page4, cursor: cursor4 } = await fetch({ limit: 3, cursor: cursor3 }) // b9

        // TODO - (FIXME) why are these garbage results in the index
        const references0 = page0.filter(l => !l.uri.includes('node_modules'))
        const references1 = page1.filter(l => !l.uri.includes('node_modules'))
        const references2 = page2.filter(l => !l.uri.includes('node_modules'))
        const references3 = page3.filter(l => !l.uri.includes('node_modules'))
        const references4 = page4.filter(l => !l.uri.includes('node_modules'))

        // Ensure paging through result sets gets us everything
        expect(references0).toEqual(references1.concat(...references2, ...references3, ...references4))

        // Ensure cursor is not provided at the end of a result set
        expect(cursor0).toBeUndefined()
        expect(cursor4).toBeUndefined()

        const extractRepos = (references: lsp.Location[]): string[] =>
            // extract the repo name from git://{repo}?{commit}#{path}, or return '' (indicating a local repo)
            Array.from(new Set(references.map(r => (r.uri.match(/git:\/\/([^?]+)\?.+/) || ['', ''])[1]))).sort()

        // Ensure paging gets us expected results per page
        expect(extractRepos(references1)).toEqual(['', 'b1', 'b10', 'b2'])
        expect(extractRepos(references2)).toEqual(['b3', 'b4', 'b5'])
        expect(extractRepos(references3)).toEqual(['b6', 'b7', 'b8'])
        expect(extractRepos(references4)).toEqual(['b9'])
    })
})

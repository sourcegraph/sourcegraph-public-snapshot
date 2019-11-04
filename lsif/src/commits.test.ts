import { hashmod, flattenCommitParents, getCommitsNear } from './commits'
import nock from 'nock'
import { XrepoDatabase } from './xrepo'
import { createCleanPostgresDatabase, createCommit } from './test-utils'

describe('discoverAndUpdateCommit', () => {
    it('should update tracked commits', async () => {
        const ca = createCommit('a')
        const cb = createCommit('b')
        const cc = createCommit('c')

        nock('http://gitserver1')
            .post('/exec')
            .reply(200, `${ca}\n${cb} ${ca}\n${cc} ${cb}`)

        const { connection, cleanup } = await createCleanPostgresDatabase()

        try {
            const xrepoDatabase = new XrepoDatabase('', connection)
            await xrepoDatabase.insertDump('test-repo', ca, '')

            await xrepoDatabase.discoverAndUpdateCommit({
                repository: 'test-repo', // hashes to gitserver1
                commit: cc,
                gitserverUrls: ['gitserver0', 'gitserver1', 'gitserver2'],
                ctx: {},
            })

            // Ensure all commits are now tracked
            expect(await xrepoDatabase.isCommitTracked('test-repo', ca)).toBeTruthy()
            expect(await xrepoDatabase.isCommitTracked('test-repo', cb)).toBeTruthy()
            expect(await xrepoDatabase.isCommitTracked('test-repo', cc)).toBeTruthy()
        } finally {
            await cleanup()
        }
    })

    it('should early-out if commit is tracked', async () => {
        const ca = createCommit('a')
        const cb = createCommit('b')

        const { connection, cleanup } = await createCleanPostgresDatabase()

        try {
            const xrepoDatabase = new XrepoDatabase('', connection)
            await xrepoDatabase.insertDump('test-repo', ca, '')
            await xrepoDatabase.updateCommits('test-repo', [[cb, '']])

            // This test ensures the following does not make a gitserver request.
            // As we did not register a nock interceptor, any request will result
            // in an exception being thrown.

            await xrepoDatabase.discoverAndUpdateCommit({
                repository: 'test-repo', // hashes to gitserver1
                commit: cb,
                gitserverUrls: ['gitserver0', 'gitserver1', 'gitserver2'],
                ctx: {},
            })
        } finally {
            await cleanup()
        }
    })

    it('should early-out if repository is unknown', async () => {
        const ca = createCommit('a')

        const { connection, cleanup } = await createCleanPostgresDatabase()

        try {
            const xrepoDatabase = new XrepoDatabase('', connection)

            // This test ensures the following does not make a gitserver request.
            // As we did not register a nock interceptor, any request will result
            // in an exception being thrown.

            await xrepoDatabase.discoverAndUpdateCommit({
                repository: 'test-repo', // hashes to gitserver1
                commit: ca,
                gitserverUrls: ['gitserver0', 'gitserver1', 'gitserver2'],
                ctx: {},
            })
        } finally {
            await cleanup()
        }
    })
})

describe('discoverAndUpdateTips', () => {
    it('should update tips', async () => {
        const ca = createCommit('a')
        const cb = createCommit('b')
        const cc = createCommit('c')
        const cd = createCommit('d')
        const ce = createCommit('e')

        nock('http://gitserver0')
            .post('/exec', { repo: 'test-repo', args: ['git', 'rev-parse', 'HEAD'] })
            .reply(200, ce)

        const { connection, cleanup } = await createCleanPostgresDatabase()

        try {
            const xrepoDatabase = new XrepoDatabase('', connection)
            await xrepoDatabase.updateCommits('test-repo', [[ca, ''], [cb, ca], [cc, cb], [cd, cc], [ce, cd]])
            await xrepoDatabase.insertDump('test-repo', ca, 'foo')
            await xrepoDatabase.insertDump('test-repo', cb, 'foo')
            await xrepoDatabase.insertDump('test-repo', cc, 'bar')

            await xrepoDatabase.discoverAndUpdateTips({
                gitserverUrls: ['gitserver0'],
                ctx: {},
            })

            expect((await xrepoDatabase.getDump('test-repo', ca, 'foo/test.ts'))!.visibleAtTip).toBeFalsy()
            expect((await xrepoDatabase.getDump('test-repo', cb, 'foo/test.ts'))!.visibleAtTip).toBeTruthy()
            expect((await xrepoDatabase.getDump('test-repo', cc, 'bar/test.ts'))!.visibleAtTip).toBeTruthy()
        } finally {
            await cleanup()
        }
    })
})

describe('discoverTips', () => {
    it('should route requests to correct gitserver', async () => {
        // Distribution of repository names to gitservers
        const requests = {
            'http://gitserver0': [1, 4, 5, 9, 10, 11, 13],
            'http://gitserver1': [0, 3, 6, 7, 12, 14],
            'http://gitserver2': [2, 8],
        }

        // Setup gitsever responses
        for (const [addr, suffixes] of Object.entries(requests)) {
            for (const i of suffixes) {
                nock(addr)
                    .post('/exec', { repo: `test-repo-${i}`, args: ['git', 'rev-parse', 'HEAD'] })
                    .reply(200, `c${i}`)
            }
        }

        // Map repo to the payloads above
        const expected = new Map<string, string>()
        for (let i = 0; i < 15; i++) {
            expected.set(`test-repo-${i}`, `c${i}`)
        }

        const { connection, cleanup } = await createCleanPostgresDatabase()

        try {
            const xrepoDatabase = new XrepoDatabase('', connection)

            for (let i = 0; i < 15; i++) {
                await xrepoDatabase.insertDump(`test-repo-${i}`, createCommit('c'), '')
            }

            const tips = await xrepoDatabase.discoverTips({
                gitserverUrls: ['gitserver0', 'gitserver1', 'gitserver2'],
                ctx: {},
                batchSize: 5,
            })

            expect(tips).toEqual(expected)
        } finally {
            await cleanup()
        }
    })
})

describe('getCommitsNear', () => {
    it('should parse response from gitserver', async () => {
        nock('http://gitserver0')
            .post('/exec', { repo: 'r', args: ['log', '--pretty=%H %P', 'l', '-150'] })
            .reply(200, 'a\nb c\nd e f\ng h i j k l')

        expect(await getCommitsNear('gitserver0', 'r', 'l')).toEqual([
            ['a', ''],
            ['b', 'c'],
            ['d', 'e'],
            ['d', 'f'],
            ['g', 'h'],
            ['g', 'i'],
            ['g', 'j'],
            ['g', 'k'],
            ['g', 'l'],
        ])
    })

    it('should handle request for unknown repository', async () => {
        nock('http://gitserver0')
            .post('/exec')
            .reply(404)

        expect(await getCommitsNear('gitserver0', 'r', 'l')).toEqual([])
    })
})

describe('flattenCommitParents', () => {
    it('should handle multiple commits', () => {
        expect(flattenCommitParents(['a', 'b c', 'd e f', 'g h i j k l'])).toEqual([
            ['a', ''],
            ['b', 'c'],
            ['d', 'e'],
            ['d', 'f'],
            ['g', 'h'],
            ['g', 'i'],
            ['g', 'j'],
            ['g', 'k'],
            ['g', 'l'],
        ])
    })
})

describe('hashmod', () => {
    /**
     * These values were generated from the following program,
     * where {text} is substituted for the values under test here.
     *
     * ```go
     * package main
     *
     * import (
     *     "crypto/md5"
     *     "encoding/binary"
     *     "fmt"
     * )
     *
     * func main() {
     *     sum := md5.Sum([]byte("{text}"))
     *     val := binary.BigEndian.Uint64(sum[:])
     *
     *     for i := 1; i <= 64; i++ {
     *         fmt.Printf("%d\n", val%uint64(i))
     *     }
     * }
     * ```
     */
    it('should hash the same as pkg/gitserver/client.go', () => {
        const testCases: { value: string; resultString: string }[] = [
            {
                value: 'foobar',
                resultString:
                    '0 1 0 1 4 3 2 1 0 9 7 9 12 9 9 1 ' +
                    '16 9 12 9 9 7 15 9 19 25 9 9 26 9 21 ' +
                    '17 18 33 9 9 17 31 12 9 6 9 22 29 9 15 ' +
                    '9 33 16 19 33 25 8 9 29 9 12 55 0 9 46 21 9 17',
            },
            {
                value: 'github.com/sourcegraph/sourcegraph',
                resultString:
                    '0 0 1 0 3 4 6 4 1 8 10 4 11 6 13 12 ' +
                    '1 10 0 8 13 10 3 4 13 24 19 20 4 28 3 28 ' +
                    '10 18 13 28 22 0 37 28 15 34 4 32 28 26 30 28 ' +
                    '13 38 1 24 41 46 43 20 19 4 20 28 29 34 55 60',
            },
        ]

        for (const { value, resultString } of testCases) {
            const results = resultString.split(' ').map(x => parseInt(x, 10))

            for (let i = 0; i < 64; i++) {
                expect(hashmod(value, i + 1)).toEqual(results[i])
            }
        }
    })
})

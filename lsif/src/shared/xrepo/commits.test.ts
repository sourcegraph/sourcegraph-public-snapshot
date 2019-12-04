import nock from 'nock'
import { flattenCommitParents, getCommitsNear, hashmod, gitserverExecLines } from './commits'

describe('getCommitsNear', () => {
    it('should parse response from gitserver', async () => {
        nock('http://gitserver0')
            .post('/exec', { repo: 'r', args: ['log', '--pretty=%H %P', 'l', '-150'] })
            .reply(200, 'a\nb c\nd e f\ng h i j k l')

        expect(await getCommitsNear('gitserver0', 'r', 'l')).toEqual(
            new Map([
                ['a', new Set()],
                ['b', new Set(['c'])],
                ['d', new Set(['e', 'f'])],
                ['g', new Set(['h', 'i', 'j', 'k', 'l'])],
            ])
        )
    })

    it('should handle request for unknown repository', async () => {
        nock('http://gitserver0')
            .post('/exec')
            .reply(404)

        expect(await getCommitsNear('gitserver0', 'r', 'l')).toEqual(new Map())
    })
})

describe('flattenCommitParents', () => {
    it('should handle multiple commits', () => {
        expect(flattenCommitParents(['a', 'b c', 'd e f', '', 'g h i j k l', 'm '])).toEqual(
            new Map([
                ['a', new Set()],
                ['b', new Set(['c'])],
                ['d', new Set(['e', 'f'])],
                ['g', new Set(['h', 'i', 'j', 'k', 'l'])],
                ['m', new Set()],
            ])
        )
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

describe('gitserverExec', () => {
    it('should not allow git as first argument', async () => {
        await expect(gitserverExecLines('', 'r', ['git', 'log'])).rejects.toThrowError(
            new Error('Gitserver commands should not be prefixed with `git`')
        )
    })
})

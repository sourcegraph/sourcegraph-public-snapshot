import { readFile } from 'mz/fs'
import { afterAll, beforeAll, vi, beforeEach, describe, expect, it, test } from 'vitest'
import createFetchMock from 'vitest-fetch-mock'

import { disableFetchCache, enableFetchCache, fetchCache, type LineOrPositionOrRange } from '@sourcegraph/common'

import { testCodeHostMountGetters as testMountGetters, testToolbarMountGetter } from '../shared/codeHostTestUtils'

import {
    getToolbarMount,
    gitlabCodeHost,
    isPrivateRepository,
    parseHash,
    windowLocation__testingOnly,
} from './codeHost'
import { repoNameOnSourcegraph } from './scrape'

const fetch = createFetchMock(vi)

describe('gitlab/codeHost', () => {
    describe('gitlabCodeHost', () => {
        testMountGetters(gitlabCodeHost, `${__dirname}/__fixtures__/repository.html`)
    })
    describe('getToolbarMount()', () => {
        testToolbarMountGetter(`${__dirname}/__fixtures__/code-views/merge-request/unified.html`, getToolbarMount)
    })

    describe('urlToFile()', () => {
        const { urlToFile } = gitlabCodeHost
        const sourcegraphURL = 'https://sourcegraph.my.org'

        beforeAll(async () => {
            document.documentElement.innerHTML = await readFile(__dirname + '/__fixtures__/merge-request.html', 'utf-8')
            windowLocation__testingOnly.value = new URL(
                'https://gitlab.com/SourcegraphCody/jsonrpc2/merge_requests/1/diffs'
            )
            globalThis.gon = { gitlab_url: 'https://gitlab.com' }
        })

        afterAll(() => {
            windowLocation__testingOnly.value = null
            // Reset resolved Sourcegraph repo name value
            repoNameOnSourcegraph.next('')
        })

        it('returns an URL to the Sourcegraph instance if the location has a viewState', () => {
            const rawRepoName = 'gitlab.com/sourcegraph/sourcegraph'
            // Update the resolved Sourcegraph repo name value
            repoNameOnSourcegraph.next(rawRepoName)

            expect(
                urlToFile(
                    sourcegraphURL,
                    {
                        repoName: 'sourcegraph/sourcegraph',
                        rawRepoName,
                        revision: 'master',
                        filePath: 'browser/src/shared/code-hosts/code_intelligence.tsx',
                        position: {
                            line: 5,
                            character: 12,
                        },
                        viewState: 'references',
                    },
                    { part: undefined }
                )
            ).toBe(
                'https://sourcegraph.my.org/sourcegraph/sourcegraph@master/-/blob/browser/src/shared/code-hosts/code_intelligence.tsx?L5:12#tab=references'
            )
        })

        it('returns an absolute URL if the location is not on the same code host', () => {
            const rawRepoName = 'gitlab.sgdev.org/sourcegraph/sourcegraph'
            // Update the resolved Sourcegraph repo name value
            repoNameOnSourcegraph.next(rawRepoName)

            expect(
                urlToFile(
                    sourcegraphURL,
                    {
                        repoName: 'sourcegraph/sourcegraph',
                        rawRepoName,
                        revision: 'master',
                        filePath: 'browser/src/shared/code-hosts/code_intelligence.tsx',
                        position: {
                            line: 5,
                            character: 12,
                        },
                    },
                    { part: undefined }
                )
            ).toBe(
                'https://sourcegraph.my.org/sourcegraph/sourcegraph@master/-/blob/browser/src/shared/code-hosts/code_intelligence.tsx?L5:12'
            )
        })
        it('returns an URL to a blob on the same code host if possible', () => {
            const rawRepoName = 'gitlab.com/sourcegraph/sourcegraph'
            // Update the resolved Sourcegraph repo name value
            repoNameOnSourcegraph.next(rawRepoName)

            expect(
                urlToFile(
                    sourcegraphURL,
                    {
                        repoName: 'sourcegraph/sourcegraph',
                        rawRepoName,
                        revision: 'main',
                        filePath: 'browser/src/shared/code-hosts/code_intelligence.tsx',
                        position: {
                            line: 5,
                            character: 12,
                        },
                    },
                    { part: undefined }
                )
            ).toBe(
                'https://gitlab.com/sourcegraph/sourcegraph/blob/main/browser/src/shared/code-hosts/code_intelligence.tsx#L5'
            )
        })
        it('returns an URL to the file on the same merge request if possible', () => {
            const rawRepoName = 'gitlab.com/SourcegraphCody/jsonrpc2'
            // Update the resolved Sourcegraph repo name value
            repoNameOnSourcegraph.next(rawRepoName)

            expect(
                urlToFile(
                    sourcegraphURL,
                    {
                        repoName: 'sourcegraph/jsonrpc2',
                        rawRepoName,
                        revision: 'changes',
                        filePath: 'call_opt.go',
                        position: {
                            line: 5,
                            character: 12,
                        },
                    },
                    { part: 'head' }
                )
            ).toBe(
                'https://gitlab.com/SourcegraphCody/jsonrpc2/merge_requests/1/diffs#9e1d3828a925c1eca74b74c20b58a9138f886d29_3_5'
            )
        })
    })
})

describe('isPrivateRepository', () => {
    beforeAll(() => {
        disableFetchCache()
    })

    afterAll(() => {
        enableFetchCache()
    })

    it('returns [private=true] if not on "gitlab.com"', async () => {
        expect(await isPrivateRepository('test-org/test-repo', fetchCache)).toBeTruthy()
    })

    describe('when on "gitlab.com"', () => {
        const EMPTY_JSON = JSON.stringify({})

        beforeAll(() => {
            fetch.enableMocks()

            windowLocation__testingOnly.value = new URL('https://gitlab.com')
        })

        beforeEach(() => {
            fetch.resetMocks()
        })

        afterAll(() => {
            fetch.disableMocks()

            windowLocation__testingOnly.value = null
        })

        it('makes request without credentials', async () => {
            fetch.mockResponseOnce(EMPTY_JSON)

            await isPrivateRepository('test-org/test-repo', fetchCache)
            expect(fetch).toHaveBeenCalledWith(expect.anything(), expect.objectContaining({ credentials: 'omit' }))
            expect(fetch).toHaveBeenCalledTimes(1)
        })

        it('returns [private=true] on unsuccessful request', async () => {
            fetch.mockRejectOnce(new Error('fake error happened for unsuccessful request'))

            expect(await isPrivateRepository('test-org/test-repo', fetchCache)).toBeTruthy()
            expect(fetch).toHaveBeenCalledTimes(1)
        })

        it('returns [private=true] if rate-limit exceeded', async () => {
            fetch.mockResponseOnce(EMPTY_JSON, {
                headers: { 'ratelimit-remaining': '0' },
            })

            expect(await isPrivateRepository('test-org/test-repo', fetchCache)).toBeTruthy()
            expect(fetch).toHaveBeenCalledTimes(1)
        })

        it('returns [private=true] when NOT 200 status response', async () => {
            fetch.mockResponseOnce(EMPTY_JSON, {
                status: 404,
            })

            expect(await isPrivateRepository('test-org/test-repo', fetchCache)).toBeTruthy()
            expect(fetch).toHaveBeenCalledTimes(1)
        })

        it('returns [private=false] when 200 status response', async () => {
            fetch.mockResponseOnce(EMPTY_JSON, { status: 200 })
            expect(await isPrivateRepository('test-org/test-repo', fetchCache)).toBeFalsy()
            expect(fetch).toHaveBeenCalledTimes(1)
        })
    })
})

describe('parseHash', () => {
    const entries: [string, LineOrPositionOrRange][] = [
        ['#L1-32', { line: 1, endLine: 32 }],
        ['#L1+32', {}],
        ['#L1-32hello', {}],
        ['#L14', { line: 14 }],
    ]

    for (const [hash, expectedValue] of entries) {
        test(`given "${hash}" as argument returns ${JSON.stringify(expectedValue)}`, () => {
            expect(parseHash(hash)).toEqual(expectedValue)
        })
    }
})

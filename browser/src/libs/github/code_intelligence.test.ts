import { existsSync, readdirSync } from 'fs'
import { startCase } from 'lodash'
import { testCodeHostMountGetters, testToolbarMountGetter } from '../code_intelligence/code_intelligence_test_utils'
import { CodeView } from '../code_intelligence/code_views'
import { createFileActionsToolbarMount, createFileLineContainerToolbarMount, githubCodeHost } from './code_intelligence'

const testCodeHost = (fixturePath: string): void => {
    if (existsSync(fixturePath)) {
        describe('githubCodeHost', () => {
            testCodeHostMountGetters(githubCodeHost, fixturePath)
        })
    }
}

const testMountGetter = (
    mountGetter: NonNullable<CodeView['getToolbarMount']>,
    mountGetterName: string,
    codeViewFixturePath: string
): void => {
    if (existsSync(codeViewFixturePath)) {
        describe(mountGetterName, () => {
            testToolbarMountGetter(codeViewFixturePath, mountGetter)
        })
    }
}

describe('github/code_intelligence', () => {
    for (const version of ['github.com', 'ghe-2.14.11']) {
        describe(version, () => {
            for (const page of readdirSync(`${__dirname}/__fixtures__/${version}`)) {
                describe(`${startCase(page)} page`, () => {
                    for (const extension of ['vanilla', 'refined-github']) {
                        describe(startCase(extension), () => {
                            // no split/unified view on blobs, and pull-request-discussion is always unified
                            if (page === 'blob' || page === 'pull-request-discussion') {
                                const directory = `${__dirname}/__fixtures__/${version}/${page}/${extension}`
                                testCodeHost(`${directory}/page.html`)
                                if (page !== 'pull-request-discussion') {
                                    testMountGetter(
                                        createFileLineContainerToolbarMount,
                                        'createSingleFileToolbarMount()',
                                        `${directory}/code-view.html`
                                    )
                                }
                            } else {
                                for (const view of ['split', 'unified']) {
                                    describe(`${startCase(view)} view`, () => {
                                        const directory = `${__dirname}/__fixtures__/${version}/${page}/${extension}/${view}`
                                        testCodeHost(`${directory}/page.html`)
                                        describe('createFileActionsToolbarMount()', () => {
                                            testMountGetter(
                                                createFileActionsToolbarMount,
                                                'createFileActionsToolbarMount()',
                                                `${directory}/code-view.html`
                                            )
                                        })
                                    })
                                }
                            }
                        })
                    }
                })
            }
        })
    }

    describe('githubCodeHost.urlToFile()', () => {
        const urlToFile = githubCodeHost.urlToFile!
        const sourcegraphURL = 'https://sourcegraph.my.org'

        beforeAll(() => {
            jsdom.reconfigure({
                url:
                    'https://github.com/sourcegraph/sourcegraph/blob/master/browser/src/libs/code_intelligence/code_intelligence.tsx',
            })
        })
        it('returns an URL to the Sourcegraph instance if the location has a viewState', () => {
            expect(
                urlToFile(
                    sourcegraphURL,
                    {
                        repoName: 'sourcegraph/sourcegraph',
                        rawRepoName: 'github.com/sourcegraph/sourcegraph',
                        rev: 'master',
                        filePath: 'browser/src/libs/code_intelligence/code_intelligence.tsx',
                        position: {
                            line: 5,
                            character: 12,
                        },
                        viewState: 'references',
                    },
                    { part: undefined }
                )
            ).toBe(
                'https://sourcegraph.my.org/sourcegraph/sourcegraph@master/-/blob/browser/src/libs/code_intelligence/code_intelligence.tsx#L5:12&tab=references'
            )
        })

        it('returns an absolute URL if the location is not on the same code host', () => {
            expect(
                urlToFile(
                    sourcegraphURL,
                    {
                        repoName: 'sourcegraph/sourcegraph',
                        rawRepoName: 'ghe.sgdev.org/sourcegraph/sourcegraph',
                        rev: 'master',
                        filePath: 'browser/src/libs/code_intelligence/code_intelligence.tsx',
                        position: {
                            line: 5,
                            character: 12,
                        },
                    },
                    { part: undefined }
                )
            ).toBe(
                'https://sourcegraph.my.org/sourcegraph/sourcegraph@master/-/blob/browser/src/libs/code_intelligence/code_intelligence.tsx#L5:12'
            )
        })
        it('returns an URL to a blob on the same code host if possible', () => {
            expect(
                urlToFile(
                    sourcegraphURL,
                    {
                        repoName: 'sourcegraph/sourcegraph',
                        rawRepoName: 'github.com/sourcegraph/sourcegraph',
                        rev: 'master',
                        filePath: 'browser/src/libs/code_intelligence/code_intelligence.tsx',
                        position: {
                            line: 5,
                            character: 12,
                        },
                    },
                    { part: undefined }
                )
            ).toBe(
                'https://github.com/sourcegraph/sourcegraph/blob/master/browser/src/libs/code_intelligence/code_intelligence.tsx#L5:12'
            )
        })
    })
})

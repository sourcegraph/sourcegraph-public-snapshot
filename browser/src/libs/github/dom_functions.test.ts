import { startCase } from 'lodash'
import { Omit } from 'utility-types'
import {
    DiffDOMFunctionsTest,
    getFixtureBody,
    testDOMFunctions,
} from '../code_intelligence/code_intelligence_test_utils'
import { diffDomFunctions, isDomSplitDiff } from './dom_functions'

type GitHubVersion = 'github.com' | 'ghe-2.14.11'
type GitHubPage = 'pull-request' | 'pull-request-discussion' | 'commit'

interface GitHubCodeViewFixture
    extends Omit<DiffDOMFunctionsTest, 'htmlFixturePath' | 'firstCharacterIsDiffIndicator'> {}

const DIFF_FIXTURES: Record<GitHubVersion, Record<GitHubPage, GitHubCodeViewFixture>> = {
    'ghe-2.14.11': {
        commit: {
            url: 'https://ghe.sgdev.org/beyang/mux/commit/1fddf523893b7475951631ed0f7e09edd9ce50d0',
            diffLineCases: [
                { diffPart: 'head', lineNumber: 80 }, // not changed
                { diffPart: 'head', lineNumber: 82 }, // added
                { diffPart: 'base', lineNumber: 82 }, // removed
            ],
        },
        'pull-request': {
            url: 'http://ghe.sgdev.org/beyang/mux/pull/1',
            diffLineCases: [
                { diffPart: 'head', lineNumber: 63 }, // not changed
                { diffPart: 'head', lineNumber: 64 }, // added
            ],
        },
        'pull-request-discussion': {
            url: 'http://ghe.sgdev.org/beyang/mux/pull/1',
            diffLineCases: [
                { diffPart: 'head', lineNumber: 64 }, // added
            ],
        },
    },
    'github.com': {
        commit: {
            url: 'https://github.com/sourcegraph/sourcegraph/commit/d3d0fe7fad2c909e3a2e4de2259dc6604983a092',
            diffLineCases: [
                { diffPart: 'head', lineNumber: 41 }, // not changed
                { diffPart: 'base', lineNumber: 42 }, // removeed
                { diffPart: 'head', lineNumber: 42 }, // added
            ],
        },
        'pull-request': {
            url: 'https://github.com/sourcegraph/sourcegraph/pull/3272/files',
            diffLineCases: [
                { diffPart: 'head', lineNumber: 570 }, // not changed
                { diffPart: 'base', lineNumber: 572 }, // removed
                { diffPart: 'head', lineNumber: 572 }, // added
            ],
        },
        'pull-request-discussion': {
            url: 'https://github.com/sourcegraph/sourcegraph/pull/3221',
            diffLineCases: [
                { diffPart: 'head', lineNumber: 62 }, // added
            ],
        },
    },
}

describe('GitHub DOM functions', () => {
    describe('diffDomFunctions', () => {
        for (const [version, pages] of Object.entries(DIFF_FIXTURES)) {
            describe(version, () => {
                for (const [page, { diffLineCases, url }] of Object.entries(pages)) {
                    describe(`${startCase(page)} page`, () => {
                        for (const extension of ['vanilla', 'refined-github']) {
                            const firstCharacterIsDiffIndicator = version !== 'github.com'

                            describe(startCase(extension), () => {
                                if (page === 'pull-request-discussion') {
                                    const htmlFixturePath = `${__dirname}/__fixtures__/${version}/${page}/${extension}/code-view.html`
                                    testDOMFunctions(diffDomFunctions, {
                                        url,
                                        htmlFixturePath,
                                        diffLineCases,
                                        firstCharacterIsDiffIndicator,
                                    })
                                } else {
                                    for (const view of ['split', 'unified']) {
                                        const htmlFixturePath = `${__dirname}/__fixtures__/${version}/${page}/${extension}/${view}/code-view.html`
                                        describe(`${startCase(view)} view`, () => {
                                            testDOMFunctions(diffDomFunctions, {
                                                url,
                                                htmlFixturePath,
                                                diffLineCases,
                                                firstCharacterIsDiffIndicator,
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
    })

    describe('isDomSplitDiff()', () => {
        for (const version of ['github.com', 'ghe-2.14.11']) {
            describe(`Version ${version}`, () => {
                const views = [
                    {
                        view: 'pull-request',
                        url: 'https://github.com/sourcegraph/sourcegraph/pull/2672/files',
                    },
                    {
                        view: 'commit',
                        url:
                            'https://github.com/sourcegraph/sourcegraph/commit/2c74f329fd03008fa0b446cd5e53234715dae3dc',
                    },
                    {
                        view: 'pull-request-discussion',
                        url: 'https://github.com/sourcegraph/sourcegraph/pull/2672/',
                    },
                ]
                for (const { view, url } of views) {
                    describe(`${startCase(view)} page`, () => {
                        beforeEach(() => {
                            // TODO ideally DOM functions would not look at global state like the URL.
                            jsdom.reconfigure({ url })
                        })
                        for (const extension of ['vanilla', 'refined-github']) {
                            describe(startCase(extension), () => {
                                if (view === 'pull-request-discussion') {
                                    it('should return false', async () => {
                                        const codeViewElement = await getFixtureBody({
                                            htmlFixturePath: `${__dirname}/__fixtures__/${version}/${view}/${extension}/code-view.html`,
                                            isFullDocument: false,
                                        })
                                        expect(isDomSplitDiff(codeViewElement)).toBe(false)
                                    })
                                } else {
                                    it('should return true for split view', async () => {
                                        const codeViewElement = await getFixtureBody({
                                            htmlFixturePath: `${__dirname}/__fixtures__/${version}/${view}/${extension}/split/code-view.html`,
                                            isFullDocument: false,
                                        })
                                        expect(isDomSplitDiff(codeViewElement)).toBe(true)
                                    })
                                    it('should return false for unified view', async () => {
                                        const codeViewElement = await getFixtureBody({
                                            htmlFixturePath: `${__dirname}/__fixtures__/${version}/${view}/${extension}/unified/code-view.html`,
                                            isFullDocument: false,
                                        })
                                        expect(isDomSplitDiff(codeViewElement)).toBe(false)
                                    })
                                }
                            })
                        }
                    })
                }
            })
        }
    })
})

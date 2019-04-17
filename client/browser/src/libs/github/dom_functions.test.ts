import { DOMFunctions } from '@sourcegraph/codeintellify'
import { DiffDOMFunctionsTest, testDOMFunctions } from '../code_intelligence/code_intelligence_test_utils'
import { diffDomFunctions } from './dom_functions'

type GitHubVersion = 'github.com' | 'ghe-2.14.11'
type GitHubPage = 'pull-request' | 'commit'
interface Fixture
    extends Pick<DiffDOMFunctionsTest, Exclude<keyof DiffDOMFunctionsTest, 'htmlFixturePath' | 'getCodeView'>> {
    version: GitHubVersion
    page: GitHubPage
}

/**
 * Creates a test suite based on an array of `Fixture` objects and a `DOMFunctions` instance.
 *
 */
function testFixtures(fixtures: Fixture[], testSuiteName: string, domFunctions: DOMFunctions): void {
    for (const { version, page, ...rest } of fixtures) {
        describe(`${version} ${page}`, () => {
            for (const extension of ['vanilla', 'refined-github']) {
                describe(extension, () => {
                    for (const view of ['split', 'unified']) {
                        describe(view, () => {
                            const htmlFixturePath = `${__dirname}/__fixtures__/${version}/${page}/${extension}/${view}/page.html`
                            testDOMFunctions(testSuiteName, domFunctions, {
                                htmlFixturePath,
                                getCodeView: () => document.querySelector('.file') as HTMLElement,
                                ...rest,
                            })
                        })
                    }
                })
            }
        })
    }
}

const DIFF_FIXTURES: Fixture[] = [
    {
        version: 'ghe-2.14.11',
        page: 'pull-request',
        url: 'http://ghe.sgdev.org/beyang/mux/pull/1',
        firstCharacterIsDiffIndicator: true,
        codeElements: [
            {
                getElement: () =>
                    [...document.querySelectorAll('.blob-code-inner')].find(e =>
                        e.textContent!.includes('-# test')
                    ) as HTMLElement,
                lineNumber: 1,
                diffPart: 'base',
            },
            {
                getElement: () =>
                    [...document.querySelectorAll('.blob-code-inner')].find(e =>
                        e.textContent!.includes('+# test')
                    ) as HTMLElement,
                lineNumber: 1,
                diffPart: 'head',
            },
        ],
    },
    {
        version: 'ghe-2.14.11',
        page: 'commit',
        url: 'https://ghe.sgdev.org/beyang/mux/commit/1fddf523893b7475951631ed0f7e09edd9ce50d0',
        firstCharacterIsDiffIndicator: true,
        codeElements: [
            {
                getElement: () =>
                    [...document.querySelectorAll('.blob-code-inner')].find(e =>
                        e.textContent!.includes('+# test')
                    ) as HTMLElement,
                lineNumber: 1,
                diffPart: 'head',
            },
        ],
    },
    {
        version: 'github.com',
        page: 'pull-request',
        url: 'http://github.com/sourcegraph/sourcegraph/pull/1',
        firstCharacterIsDiffIndicator: false,
        codeElements: [
            {
                getElement: () =>
                    [...document.querySelectorAll('.blob-code-inner')].find(e =>
                        e.textContent!.includes('golang.org/x/net/trace')
                    ) as HTMLElement,
                lineNumber: 43,
                diffPart: 'base',
            },
            {
                getElement: () =>
                    [...document.querySelectorAll('.blob-code-inner')].find(e =>
                        e.textContent!.includes('errorString returns the error string')
                    ) as HTMLElement,
                lineNumber: 1333,
                diffPart: 'head',
            },
        ],
    },
    {
        version: 'github.com',
        page: 'commit',
        url: 'https://github.com/sourcegraph/sourcegraph/commit/1fddf523893b7475951631ed0f7e09edd9ce50d0',
        firstCharacterIsDiffIndicator: false,
        codeElements: [
            {
                getElement: () => document.querySelector('.blob-code-deletion .blob-code-inner') as HTMLElement,
                lineNumber: 42,
                diffPart: 'base',
            },
            {
                getElement: () => document.querySelector('.blob-code-addition .blob-code-inner') as HTMLElement,
                lineNumber: 42,
                diffPart: 'head',
            },
        ],
    },
]

describe('GitHub DOM Functions', () => {
    testFixtures(DIFF_FIXTURES, 'diffDOMFunctions', diffDomFunctions)
})

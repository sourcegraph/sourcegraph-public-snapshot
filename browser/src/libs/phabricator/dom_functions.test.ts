import { startCase } from 'lodash'
import { DOMFunctionsTest, testDOMFunctions } from '../code_intelligence/code_intelligence_test_utils'
import { diffDomFunctions, diffusionDOMFns } from './dom_functions'

type PhabricatorPage = 'commit' | 'differential'

interface PhabricatorCodeViewFixture
    extends Omit<DOMFunctionsTest, 'htmlFixturePath' | 'firstCharacterIsDiffIndicator'> {}

describe('Phabricator DOM functions', () => {
    describe('diffDOMFunctions', () => {
        const DIFF_FIXTURES: Record<PhabricatorPage, PhabricatorCodeViewFixture> = {
            commit: {
                lineCases: [
                    { diffPart: 'head', lineNumber: 3 }, // not changed
                    { diffPart: 'head', lineNumber: 7 }, // added
                    { diffPart: 'base', lineNumber: 10 }, // removed
                ],
            },
            differential: {
                lineCases: [
                    { diffPart: 'head', lineNumber: 9 }, // not changed
                    { diffPart: 'head', lineNumber: 10 }, // added
                    // TODO test case for removed line
                ],
            },
        }
        for (const [page, testCase] of Object.entries(DIFF_FIXTURES)) {
            describe(`${startCase(page)} Page`, () => {
                for (const view of ['split', 'unified']) {
                    const htmlFixturePath = `${__dirname}/__fixtures__/code-views/${page}/${view}.html`
                    describe(`${startCase(view)} view`, () => {
                        // https://phabricator.sgdev.org/D3#diff-7ddfb3e0
                        testDOMFunctions(diffDomFunctions, {
                            ...testCase,
                            htmlFixturePath,
                            firstCharacterIsDiffIndicator: false,
                        })
                    })
                }
            })
        }
    })

    describe('diffusionDOMFns', () => {
        const htmlFixturePath = `${__dirname}/__fixtures__/code-views/diffusion.html`
        // https://phabricator.sgdev.org/source/test/browse/master/main.go;48600480cce9f832f7daacab256fbfdeb7002603
        testDOMFunctions(diffusionDOMFns, {
            htmlFixturePath,
            lineCases: [{ lineNumber: 1 }, { lineNumber: 10 }],
            firstCharacterIsDiffIndicator: false,
        })
    })
})

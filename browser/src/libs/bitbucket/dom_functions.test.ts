import { startCase } from 'lodash'
import { testDOMFunctions } from '../code_intelligence/code_intelligence_test_utils'
import { diffDOMFunctions, singleFileDOMFunctions } from './dom_functions'

describe('Bitbucket DOM functions', () => {
    describe('diffDOMFunctions', () => {
        for (const view of ['split', 'unified']) {
            describe(`${startCase(view)} view`, () => {
                testDOMFunctions(diffDOMFunctions, {
                    htmlFixturePath: `${__dirname}/__fixtures__/code-views/pull-request/${view}/modified.html`,
                    lineCases: [
                        { diffPart: 'head', lineNumber: 54 }, // not changed
                        { diffPart: 'head', lineNumber: 60 }, // added
                        { diffPart: 'base', lineNumber: 102 }, // removed
                    ],
                })
            })
        }
    })

    describe('singleFileDOMFunctions', () => {
        const htmlFixturePath = `${__dirname}/__fixtures__/code-views/single-file-source.html`
        testDOMFunctions(singleFileDOMFunctions, {
            htmlFixturePath,
            lineCases: [{ lineNumber: 1 }, { lineNumber: 18 }],
        })
    })
})

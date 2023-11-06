import { afterAll, describe, expect, it } from '@jest/globals'
import { of } from 'rxjs'

import { findPositionsFromEvents } from './positions'
import { DOM } from './testutils/dom'
import { createMouseEvent } from './testutils/mouse'

describe('positions', () => {
    const dom = new DOM()
    const testcases = dom.createCodeViews()

    afterAll(dom.cleanup)

    // Without this placeholder, jest throws an error saying there are no tests.
    it('placeholder', () => {})

    for (const tokenize of [false, true]) {
        for (const codeView of testcases) {
            it((tokenize ? 'tokenizes' : 'does not tokenize') + ` the DOM when tokenize: ${String(tokenize)}`, () => {
                of(codeView.codeView)
                    .pipe(findPositionsFromEvents({ domFunctions: codeView, tokenize }))
                    .subscribe()

                const htmlBefore = codeView.getCodeElementFromLineNumber(codeView.codeView, 5)!.outerHTML
                codeView.getCodeElementFromLineNumber(codeView.codeView, 5)!.dispatchEvent(
                    createMouseEvent('mouseover', {
                        x: 0,
                        y: 0,
                    })
                )
                const htmlAfter = codeView.getCodeElementFromLineNumber(codeView.codeView, 5)!.outerHTML

                if (tokenize) {
                    expect(htmlBefore).not.toEqual(htmlAfter)
                } else {
                    expect(htmlBefore).toEqual(htmlAfter)
                }
            })
        }
    }
})

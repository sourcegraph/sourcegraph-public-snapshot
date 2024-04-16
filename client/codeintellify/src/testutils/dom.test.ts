import { afterAll, describe, expect, it } from 'vitest'

import { DOM } from './dom'

describe('can create dom elements from generated code tables', () => {
    const dom = new DOM()
    afterAll(dom.cleanup)

    it('can create the code view test cases and their helper function work', () => {
        for (const codeViewProps of dom.createCodeViews()) {
            const { codeView, getCodeElementFromTarget, getCodeElementFromLineNumber, getLineNumberFromCodeElement } =
                codeViewProps

            for (let index = 1; index < 10; index++) {
                const cellFromLine = getCodeElementFromLineNumber(codeView, index)
                expect(cellFromLine).not.toEqual(null)
                const cellFromTarget = getCodeElementFromTarget(cellFromLine!)
                expect(cellFromTarget).toEqual(cellFromLine)
                const line = getLineNumberFromCodeElement(cellFromTarget!)
                expect(line).toEqual(index)
            }
        }
    })
})

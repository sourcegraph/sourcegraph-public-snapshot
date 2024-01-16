import { describe, expect, it, afterEach } from 'vitest'

import { isCodyOnlyLicense, isCodeSearchOnlyLicense, isCodeSearchPlusCodyLicense } from './license'

describe('licensing utils', () => {
    const origContext = window.context
    afterEach(() => {
        window.context = origContext
    })

    it('Cody only license', () => {
        window.context = {
            licenseInfo: {
                features: {
                    cody: true,
                    codeSearch: false,
                },
            },
        } as any

        expect(isCodyOnlyLicense()).toEqual(true)
        expect(isCodeSearchOnlyLicense()).toEqual(false)
        expect(isCodeSearchPlusCodyLicense()).toEqual(false)
    })

    it('Code Search only license', () => {
        window.context = {
            licenseInfo: {
                features: {
                    cody: false,
                    codeSearch: true,
                },
            },
        } as any

        expect(isCodyOnlyLicense()).toEqual(false)
        expect(isCodeSearchOnlyLicense()).toEqual(true)
        expect(isCodeSearchPlusCodyLicense()).toEqual(false)
    })

    it('Code Search plus Cody license', () => {
        window.context = {
            licenseInfo: {
                features: {
                    cody: true,
                    codeSearch: true,
                },
            },
        } as any

        expect(isCodyOnlyLicense()).toEqual(false)
        expect(isCodeSearchOnlyLicense()).toEqual(false)
        expect(isCodeSearchPlusCodyLicense()).toEqual(true)
    })
})

import { describe, expect, it } from '@jest/globals'

import { parseProductReference } from './SiteAdminFeatureFlagsPage'

describe('parseProductReference', () => {
    it('parses main branch tags correctly', () => {
        const parsed = parseProductReference('214157_2023-04-19_5.0-89aa613e7e1e')
        expect(parsed).toEqual('89aa613e7e1e')
    })
    it('leaves tags untouched', () => {
        const parsed = parseProductReference('v5.0.0')
        expect(parsed).toEqual('v5.0.0')
    })
    it('unknown versions fall back to main', () => {
        for (const reference of [
            '0.0.0', // dev version
            '214157_2023-04-19', // unknown format
            '214157_2023-04-19_foobar', // different last segment format
        ]) {
            const parsed = parseProductReference(reference)
            expect(parsed).toEqual('main')
        }
    })
})

import { describe, expect, it } from '@jest/globals'

import { createLinkUrl } from './createLinkUrl'

describe('createLinkUrl', () => {
    it('generates full URL', () => {
        expect(
            createLinkUrl({
                pathname: 'https://sourcegraph.com/search',
                search: 'q=hello',
                hash: 'h1',
            })
        ).toBe('https://sourcegraph.com/search?q=hello#h1')
    })

    it('generates relative URL', () => {
        expect(createLinkUrl({ search: 's=123', hash: 'test-hash' })).toBe('?s=123#test-hash')
        expect(createLinkUrl({ pathname: '', search: 's=123', hash: 'test-hash' })).toBe('?s=123#test-hash')
    })

    it('supports variations search and hash', () => {
        expect(createLinkUrl({ search: 'q=param' })).toBe('?q=param')
        expect(createLinkUrl({ search: '?q=param' })).toBe('?q=param')
        expect(createLinkUrl({ hash: 'section' })).toBe('#section')
        expect(createLinkUrl({ hash: '#section' })).toBe('#section')
    })
})

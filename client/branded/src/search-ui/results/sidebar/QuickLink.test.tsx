import { describe, expect, it } from '@jest/globals'

import type { QuickLink } from '@sourcegraph/shared/src/schema/settings.schema'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { getQuickLinks } from './QuickLink'

describe('QuickLink', () => {
    it('should be empty if quicklinks not present', () => {
        const links = getQuickLinks({ subjects: [], final: {} })
        expect(links.length).toBe(0)
    })

    it('should have correct links when quicklinks present', () => {
        const quicklinks: QuickLink[] = [
            { name: 'Home', url: '/' },
            {
                name: 'This is a quicklink with a very long name lorem ipsum dolor sit amet',
                url: 'http://example.com',
                description: 'Example QuickLink',
            },
        ]
        const links = getQuickLinks({ subjects: [], final: { quicklinks } })
        expect(links.length).toBe(2)
        expect(renderWithBrandedContext(<>{links}</>).asFragment()).toMatchSnapshot()
    })
})

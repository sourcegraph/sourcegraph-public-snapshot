import React from 'react'

import { renderWithRouter } from '@sourcegraph/shared/src/testing/render-with-router'

import { ToggleRenderedFileMode } from './ToggleRenderedFileMode'

describe('ToggleRenderedFileMode', () => {
    const route = '/github.com/sourcegraph/sourcegraph/-/blob/README.md'

    describe('in rendered view', () => {
        it('renders link correctly', () => {
            const renderResult = renderWithRouter(<ToggleRenderedFileMode mode="rendered" actionType="nav" />, {
                route,
            })

            const toggle = renderResult.getByText('Raw')
            expect(toggle.closest('a')).toHaveAttribute('href', `${route}?view=code`)
        })
    })

    describe('in code view', () => {
        it('renders link correctly', () => {
            const renderResult = renderWithRouter(<ToggleRenderedFileMode mode="code" actionType="nav" />, {
                route: `${route}?view=code`,
            })

            const toggle = renderResult.getByText('Formatted')
            expect(toggle.closest('a')).toHaveAttribute('href', route)
        })

        it('still renders link correctly when a line has been selected', () => {
            const renderResult = renderWithRouter(<ToggleRenderedFileMode mode="code" actionType="nav" />, {
                route: `${route}?L10&view=code`,
            })

            const toggle = renderResult.getByText('Formatted')
            expect(toggle.closest('a')).toHaveAttribute('href', route)
        })
    })
})

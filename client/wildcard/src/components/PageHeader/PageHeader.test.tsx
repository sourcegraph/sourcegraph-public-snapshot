import { beforeEach, describe, expect, it } from '@jest/globals'
import type { RenderResult } from '@testing-library/react'
import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'

import { renderWithBrandedContext } from '../../testing'

import { PageHeader } from './PageHeader'

describe('PageHeader', () => {
    let queries: RenderResult
    const breadcrumbs = [
        {
            to: '/link-0',
            icon: PuzzleOutlineIcon,
        },
        {
            to: '/link-1',
            text: 'Link 1',
        },
        {
            to: '/link-2',
            text: 'Link 2',
        },
        {
            text: 'Current Page',
        },
    ]

    beforeEach(() => {
        queries = renderWithBrandedContext(<PageHeader path={breadcrumbs} />)
    })

    it('renders correctly', () => {
        expect(queries.baseElement).toMatchSnapshot()
    })

    it('renders links correctly', () => {
        expect(queries.getByText('Link 1').closest('a')?.pathname).toBe('/link-1')
        expect(queries.getByText('Link 2').closest('a')?.pathname).toBe('/link-2')
    })

    it('renders current page correctly', () => {
        expect(queries.getByText(/Current Page/).tagName).not.toBe('A')
    })
})

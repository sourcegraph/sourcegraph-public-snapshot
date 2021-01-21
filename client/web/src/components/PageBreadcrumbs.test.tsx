import React from 'react'
import { render, RenderResult } from '@testing-library/react'
import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'
import { PageBreadcrumbs } from './PageBreadcrumbs'

describe('PageBreadcrumbs', () => {
    let queries: RenderResult
    const breadcrumbs = [
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
        queries = render(<PageBreadcrumbs icon={PuzzleOutlineIcon} path={breadcrumbs} />)
    })

    it('renders links correctly', () => {
        expect((queries.getByText('Link 1') as HTMLAnchorElement).pathname).toBe('/link-1')
        expect((queries.getByText('Link 2') as HTMLAnchorElement).pathname).toBe('/link-2')
    })

    it('renders current page correctly', () => {
        expect(queries.getByText(/Current Page/).tagName).not.toBe('A')
    })
})

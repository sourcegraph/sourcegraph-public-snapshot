import React from 'react'

import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { useSeriesToggle } from './use-series-toggle'

const UseSeriesToggleExample: React.FunctionComponent = () => {
    const { toggle, selectedSeriesIds, isSeriesHovered, isSeriesSelected, setHoveredId } = useSeriesToggle()

    return (
        <div>
            <div onMouseEnter={() => setHoveredId('foo')}>
                {isSeriesSelected('foo') && <span>Foo is selected</span>}
                {isSeriesHovered('foo') && <span>Foo is hovered</span>}
                <button onClick={() => toggle('foo')}>Foo</button>
            </div>
            <div onMouseEnter={() => setHoveredId('bar')}>
                {isSeriesSelected('bar') && <span>Bar is selected</span>}
                {isSeriesHovered('bar') && <span>Bar is hovered</span>}
                <button onClick={() => toggle('bar')}>Bar</button>
            </div>

            <div>Selected series: {selectedSeriesIds.join(',')}</div>
        </div>
    )
}

describe('useSeriesToggle', () => {
    it('renders all series when none are selected', () => {
        render(<UseSeriesToggleExample />)

        expect(screen.getByText('Foo is selected')).toBeInTheDocument()
        expect(screen.getByText('Bar is selected')).toBeInTheDocument()
    })

    it('renders only Foo series when Foo is selected', () => {
        render(<UseSeriesToggleExample />)
        userEvent.click(screen.getByRole('button', { name: 'Foo' }))

        expect(screen.getByText('Foo is selected')).toBeInTheDocument()
        expect(screen.queryByText('Bar is selected')).not.toBeInTheDocument()
        expect(screen.getByText('Selected series: foo')).toBeInTheDocument()
    })

    it('renders only Foo & Bar series when Foo & Bar are selected', () => {
        render(<UseSeriesToggleExample />)
        userEvent.click(screen.getByRole('button', { name: 'Foo' }))
        userEvent.click(screen.getByRole('button', { name: 'Bar' }))

        expect(screen.getByText('Foo is selected')).toBeInTheDocument()
        expect(screen.getByText('Bar is selected')).toBeInTheDocument()
        expect(screen.getByText('Selected series: foo,bar')).toBeInTheDocument()
    })

    it('renders "Foo is hovered"', () => {
        render(<UseSeriesToggleExample />)
        userEvent.hover(screen.getByRole('button', { name: 'Foo' }))

        expect(screen.getByText('Foo is hovered')).toBeInTheDocument()
    })

    it('renders "Bar is hovered"', () => {
        render(<UseSeriesToggleExample />)
        userEvent.hover(screen.getByRole('button', { name: 'Foo' }))
        userEvent.hover(screen.getByRole('button', { name: 'Bar' }))

        expect(screen.queryByText('Foo is hovered')).not.toBeInTheDocument()
        expect(screen.getByText('Bar is hovered')).toBeInTheDocument()
    })
})

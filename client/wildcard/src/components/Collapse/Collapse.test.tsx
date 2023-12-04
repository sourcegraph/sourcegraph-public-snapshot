import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import sinon from 'sinon'
import { describe, expect, it } from 'vitest'

import { Collapse, CollapseHeader, CollapsePanel } from '.'

describe('Collapse', () => {
    it('should render opened when `openByDefault` is true', () => {
        const { asFragment } = render(
            <Collapse openByDefault={true}>
                <CollapseHeader>I am Collapse Header</CollapseHeader>
                <CollapsePanel>I am Collapse Panel</CollapsePanel>
            </Collapse>
        )

        expect(screen.getByText(/I am Collapse Panel/)).toHaveClass('show')

        expect(asFragment()).toMatchSnapshot()
    })

    it('should toggle when clicking on header', () => {
        const { asFragment } = render(
            <Collapse>
                <CollapseHeader>I am Collapse Header</CollapseHeader>
                <CollapsePanel>I am Collapse Panel</CollapsePanel>
            </Collapse>
        )

        expect(screen.getByText(/I am Collapse Panel/)).not.toHaveClass('show')

        userEvent.click(screen.getByText(/I am Collapse Header/))

        expect(screen.getByText(/I am Collapse Panel/)).toHaveClass('show')

        expect(asFragment()).toMatchSnapshot()

        userEvent.click(screen.getByText(/I am Collapse Header/))
        expect(screen.getByText(/I am Collapse Panel/)).not.toHaveClass('show')

        expect(asFragment()).toMatchSnapshot()
    })

    it('should dispatch `onOpenChange` on controlled mode', () => {
        const handleOpenChange = sinon.spy()

        render(
            <Collapse isOpen={false} onOpenChange={handleOpenChange}>
                <CollapseHeader>I am Collapse Header</CollapseHeader>
                <CollapsePanel>I am Collapse Panel</CollapsePanel>
            </Collapse>
        )

        userEvent.click(screen.getByText(/I am Collapse Header/))

        sinon.assert.calledOnceWithExactly(handleOpenChange, true)
    })
})

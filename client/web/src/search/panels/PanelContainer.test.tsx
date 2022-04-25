import { render } from '@testing-library/react'

import { PanelContainer } from './PanelContainer'

describe('PanelContainer', () => {
    const defaultProps = {
        title: 'Test Panel',
        state: 'loading',
        loadingContent: <div>Loading</div>,
        populatedContent: <div>Content</div>,
        emptyContent: <div>Empty</div>,
    }

    test('loading state', () => {
        expect(render(<PanelContainer {...defaultProps} state="loading" />).asFragment()).toMatchSnapshot()
    })

    test('empty state', () => {
        expect(render(<PanelContainer {...defaultProps} state="empty" />).asFragment).toMatchSnapshot()
    })

    test('content state', () => {
        expect(render(<PanelContainer {...defaultProps} state="populated" />).asFragment).toMatchSnapshot()
    })

    test('with action buttons', () => {
        const actionButtons = <button type="button">Button</button>
        expect(
            render(<PanelContainer {...defaultProps} state="populated" actionButtons={actionButtons} />).asFragment()
        ).toMatchSnapshot()
    })
})

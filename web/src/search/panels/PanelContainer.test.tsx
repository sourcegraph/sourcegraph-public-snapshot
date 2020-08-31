import renderer from 'react-test-renderer'
import React from 'react'
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
        expect(renderer.create(<PanelContainer {...defaultProps} state="loading" />)).toMatchSnapshot()
    })

    test('empty state', () => {
        expect(renderer.create(<PanelContainer {...defaultProps} state="empty" />)).toMatchSnapshot()
    })

    test('content state', () => {
        expect(renderer.create(<PanelContainer {...defaultProps} state="populated" />)).toMatchSnapshot()
    })

    test('with action buttons', () => {
        const actionButtons = <button type="button">Button</button>
        expect(
            renderer.create(<PanelContainer {...defaultProps} state="content" actionButtons={actionButtons} />)
        ).toMatchSnapshot()
    })
})

import React from 'react'
import renderer from 'react-test-renderer'
import { DraftBadge } from './DraftBadge'

describe('DraftBadge', () => {
    test('renders', () => {
        const result = renderer.create(<DraftBadge />)
        expect(result.toJSON()).toMatchSnapshot()
    })
})

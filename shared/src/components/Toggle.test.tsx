import React from 'react'
import renderer from 'react-test-renderer'
import { Toggle } from './Toggle'

describe('Toggle', () => {
    test('value is false', () => {
        const component = renderer.create(<Toggle value={false} />)
        const tree = component.toJSON()
        expect(tree).toMatchSnapshot()
    })

    test('value is true', () => {
        const component = renderer.create(<Toggle value={true} />)
        const tree = component.toJSON()
        expect(tree).toMatchSnapshot()
    })

    test('disabled', () => {
        const component = renderer.create(<Toggle disabled={true} />)
        const tree = component.toJSON()
        expect(tree).toMatchSnapshot()
    })
})

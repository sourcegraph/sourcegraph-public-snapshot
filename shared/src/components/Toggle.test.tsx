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
        let tree = component.toJSON()
        expect(tree).toMatchSnapshot()

        // Clicking while disabled is a noop.
        tree!.props.onClick({ preventDefault: () => void 0, currentTarget: { blur: () => void 0 } })
        tree = component.toJSON()
        expect(tree).toMatchSnapshot()
    })

    test('className', () => expect(renderer.create(<Toggle className="c" />).toJSON()).toMatchSnapshot())
})

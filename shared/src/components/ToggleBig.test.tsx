import React from 'react'
import renderer from 'react-test-renderer'
import { ToggleBig } from './ToggleBig'

describe('Toggle', () => {
    test('value is false', () => {
        const component = renderer.create(<ToggleBig value={false} />)
        const tree = component.toJSON()
        expect(tree).toMatchSnapshot()
    })

    test('value is true', () => {
        const component = renderer.create(<ToggleBig value={true} />)
        const tree = component.toJSON()
        expect(tree).toMatchSnapshot()
    })

    test('disabled', () => {
        const component = renderer.create(<ToggleBig disabled={true} />)
        let tree = component.toJSON()
        expect(tree).toMatchSnapshot()

        // Clicking while disabled is a noop.
        tree!.props.onClick({ preventDefault: () => undefined, currentTarget: { blur: () => undefined } })
        tree = component.toJSON()
        expect(tree).toMatchSnapshot()
    })

    test('className', () => expect(renderer.create(<ToggleBig className="c" />).toJSON()).toMatchSnapshot())
})

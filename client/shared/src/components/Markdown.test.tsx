import { createMemoryHistory } from 'history'
import React from 'react'
import renderer from 'react-test-renderer'

import { Markdown } from './Markdown'

describe('Markdown', () => {
    it('renders', () => {
        const history = createMemoryHistory()
        const component = renderer.create(<Markdown history={history} dangerousInnerHTML="hello" />)
        expect(component.toJSON()).toMatchSnapshot()
    })
})

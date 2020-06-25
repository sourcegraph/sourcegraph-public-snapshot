import React from 'react'
import { Markdown } from './Markdown'
import { createMemoryHistory } from 'history'
import { mount } from 'enzyme'

describe('Markdown', () => {
    it('renders', () => {
        const history = createMemoryHistory()
        expect(mount(<Markdown history={history} dangerousInnerHTML="hello" />)).toMatchSnapshot()
    })
})

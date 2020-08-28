import renderer from 'react-test-renderer'
import React from 'react'
import { EnterpriseHomePanels } from './EnterpriseHomePanels'

describe('EnterpriseHomePanels', () => {
    test('render', () => {
        expect(renderer.create(<EnterpriseHomePanels />)).toMatchSnapshot()
    })
})

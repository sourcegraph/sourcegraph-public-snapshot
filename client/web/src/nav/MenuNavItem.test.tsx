import { mount } from 'enzyme'
import React from 'react'

import { MenuNavItem } from './MenuNavItem'

describe('MenuNavItem', () => {
    test('add menu children', () => {
        expect(
            mount(
                <MenuNavItem>
                    <div>item</div>
                    <div>item</div>
                </MenuNavItem>
            )
        ).toMatchSnapshot()
    })
})

import React from 'react'
import { MenuNavItem } from './MenuNavItem'
import { mount } from 'enzyme'

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

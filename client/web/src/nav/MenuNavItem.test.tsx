import { render } from '@testing-library/react'
import React from 'react'

import { MenuNavItem } from './MenuNavItem'

describe('MenuNavItem', () => {
    test('add menu children', () => {
        expect(
            render(
                <MenuNavItem>
                    <div>item</div>
                    <div>item</div>
                </MenuNavItem>
            ).asFragment()
        ).toMatchSnapshot()
    })
})

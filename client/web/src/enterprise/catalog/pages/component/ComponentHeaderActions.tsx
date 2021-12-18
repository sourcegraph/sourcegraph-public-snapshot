import React from 'react'

import { Menu } from '@sourcegraph/wildcard'
import { ComponentTagsFields } from '../../../../graphql-operations'

interface Props {
    component: ComponentTagsFields
}

export const ComponentHeaderActions: React.FunctionComponent<Props> = () => (
    <Menu>
        <MenuButton variant="primary" outline={true}>
            Actions <span aria-hidden={true}>▾</span>
        </MenuButton>
        <MenuPopover>
            <MenuItems>
                <MenuHeader>This is a menu</MenuHeader>
                <MenuItem onSelect={() => alert('Clicked!')}>Click me</MenuItem>
                <MenuItem onSelect={() => alert('Clicked!')}>Alternative action</MenuItem>
                <MenuDivider />
                <MenuLink as="a" href="https://www.example.com">
                    Go somewhere
                </MenuLink>
            </MenuItems>
        </MenuPopover>
    </Menu>
)

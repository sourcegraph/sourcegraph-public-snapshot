import { Meta, Story } from '@storybook/react'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Menu, MenuButton, MenuDivider, MenuHeader, MenuItem, MenuLink, MenuPopover, MenuItems } from '.'

const config: Meta = {
    title: 'wildcard/Menu',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: Menu,
    },
}

export default config

export const MenuExample: Story = () => (
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

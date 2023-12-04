import type { Meta, StoryFn } from '@storybook/react'
import { noop } from 'lodash'

import { BrandedStory } from '../../stories/BrandedStory'
import { Link } from '../Link'

import { Menu, MenuButton, MenuDivider, MenuHeader, MenuItem, MenuLink, MenuList } from '.'

const config: Meta = {
    title: 'wildcard/Menu',

    decorators: [story => <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>],

    parameters: {
        component: Menu,
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
    },
}

export default config

export const MenuExample: StoryFn = () => (
    <Menu>
        <MenuButton variant="primary" outline={true}>
            Actions <span aria-hidden={true}>▾</span>
        </MenuButton>

        <MenuList>
            <MenuHeader>This is a menu</MenuHeader>
            <MenuItem onSelect={() => alert('Clicked!')}>Click me</MenuItem>
            <MenuItem onSelect={() => alert('Clicked!')}>Alternative action</MenuItem>
            <MenuItem onSelect={noop} disabled={true}>
                I'm disabled
            </MenuItem>
            <MenuDivider />
            <MenuLink as={Link} to="https://www.example.com">
                Go somewhere
            </MenuLink>
            <MenuLink disabled={true} as={Link} to="https://www.example.com">
                Disabled link
            </MenuLink>
        </MenuList>
    </Menu>
)

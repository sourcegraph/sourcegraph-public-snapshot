import { Meta, Story } from '@storybook/react'
import { noop } from 'lodash'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Link } from '../Link'

import { Menu, MenuButton, MenuDivider, MenuHeader, MenuItem, MenuLink, MenuList } from '.'

const config: Meta = {
    title: 'wildcard/Menu',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: Menu,
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
    },
}

export default config

export const MenuExample: Story = () => (
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

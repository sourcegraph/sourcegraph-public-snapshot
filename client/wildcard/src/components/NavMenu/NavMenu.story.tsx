import React from 'react'

import { Meta, Story } from '@storybook/react'
import { noop } from 'lodash'
import AntennaIcon from 'mdi-react/AntennaIcon'
import BarChartIcon from 'mdi-react/BarChartIcon'
import FileTreeOutlineIcon from 'mdi-react/FileTreeOutlineIcon'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuIcon from 'mdi-react/MenuIcon'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Badge } from '../Badge'
import { Button } from '../Button'
import { Select } from '../Form'
import { Icon } from '../Icon'

import { NavMenu, NavMenuSectionProps } from './NavMenu'
import { avatarUrl } from './utils'

import styles from './NavMenu.module.scss'

const config: Meta = {
    title: 'wildcard/NavMenu',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: NavMenu,
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
    },
}

export default config

const themeItemContent = (
    <div className="d-flex align-items-center">
        <div className="mr-2">Theme</div>
        <Select
            aria-label=""
            isCustomStyle={true}
            selectSize="sm"
            selectClassName="test-theme-toggle"
            onChange={noop}
            value="light"
            className="mb-0 flex-1"
        >
            <option value="light">Light</option>
            <option value="dark">Dark</option>
            <option value="system">System</option>
        </Select>
    </div>
)

const navItems: NavMenuSectionProps[] = [
    {
        headerContent: (
            <>
                {' '}
                Signed in as <strong>@ralph</strong>
            </>
        ),
    },
    {
        navItems: [
            {
                itemContent: 'Settings',
                to: '/users/ralph/settings/organizations',
            },
            {
                itemContent: 'Your repositories',
                to: '/users/ralph/settings/repositories',
            },
            {
                itemContent: 'Saved searches',
                to: '/users/ralph/searches',
            },
            {
                itemContent: (
                    <>
                        Your organizations <Badge variant="info">NEW</Badge>
                    </>
                ),
                to: '/users/ralph/settings/organizations',
            },
        ],
    },
    {
        navItems: [{ itemContent: themeItemContent, itemAs: 'div' }],
    },
    {
        headerContent: 'Your organizations',
        navItems: [
            {
                itemContent: 'Sourcegraph',
                to: '/users/ralph/settings/organizations',
                key: 'sourcegraph-123456',
            },
            {
                itemContent: 'Gitstart',
                to: '/users/ralph/settings/repositories',
                key: 'gitstart-123456',
            },
        ],
    },
    {
        navItems: [
            {
                itemContent: 'Site admin',
                onSelect: noop,
            },
            {
                itemContent: 'Help',
                onSelect: noop,
                suffixIcon: OpenInNewIcon,
            },
            {
                itemContent: 'Keyboard shortcuts',
                onSelect: noop,
            },
            {
                itemContent: 'Keyboard shortcuts',
                onSelect: noop,
            },
        ],
    },
    {
        navItems: [
            {
                itemContent: 'About Sourcegraph',
                suffixIcon: OpenInNewIcon,
                to: 'https://about.sourcegraph.com',
            },
            {
                itemContent: 'Browser Extension',
                suffixIcon: OpenInNewIcon,
                to: 'https://docs.sourcegraph.com/integration/browser_extension',
            },
        ],
    },
]

export const UserNav: Story = () => (
    <NavMenu
        navTrigger={{
            variant: 'icon',
            content: isOpen => (
                <>
                    <Icon as="img" className={styles.avatar} src={avatarUrl} />
                    <Icon as={isOpen ? MenuUpIcon : MenuDownIcon} />
                </>
            ),
        }}
        sections={navItems}
    />
)

const singleSectionNavItems: NavMenuSectionProps[] = [
    {
        navItems: [
            { itemContent: 'Batch Changes', prefixIcon: FileTreeOutlineIcon },
            {
                itemContent: (
                    <Button variant="link" className="w-100 text-left">
                        <Icon as={BarChartIcon} /> Insight
                    </Button>
                ),
            },
            {
                itemContent: (
                    <Button variant="link" className="w-100 text-left">
                        <Icon as={AntennaIcon} /> Monitoring
                    </Button>
                ),
            },
        ],
    },
]

export const SingleSectionNavMenuExample: Story = () => (
    <NavMenu
        navTrigger={{
            variant: 'secondary',
            outline: true,
            content: isOpen => (
                <>
                    <Icon as={MenuIcon} />
                    <Icon as={isOpen ? MenuUpIcon : MenuDownIcon} />
                </>
            ),
        }}
        sections={singleSectionNavItems}
    />
)

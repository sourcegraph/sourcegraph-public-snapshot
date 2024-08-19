import { mdiPoll, mdiAntenna, mdiMenu, mdiMenuUp, mdiMenuDown } from '@mdi/js'
import type { Meta, StoryFn } from '@storybook/react'
import { noop } from 'lodash'
import FileTreeOutlineIcon from 'mdi-react/FileTreeOutlineIcon'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'

import { BrandedStory } from '../../stories/BrandedStory'
import { Badge } from '../Badge'
import { Button } from '../Button'
import { Select } from '../Form'
import { Icon } from '../Icon'

import { NavMenu, type NavMenuSectionProps } from './NavMenu'

import styles from './NavMenu.module.scss'

const config: Meta = {
    title: 'wildcard/NavMenu',

    decorators: [story => <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>],

    parameters: {
        component: NavMenu,
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
            data-testid="theme-toggle"
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
                content: 'Settings',
                to: '/users/ralph/settings/organizations',
                key: 'Settings',
            },
            {
                content: 'Your repositories',
                to: '/users/ralph/settings/repositories',
                key: 'repositories',
            },
            {
                content: 'Saved searches',
                to: '/users/ralph/searches',
                key: 'searches',
            },
            {
                content: (
                    <>
                        Your organizations <Badge variant="info">NEW</Badge>
                    </>
                ),
                to: '/users/ralph/settings/organizations',
                key: 'organizations',
            },
        ],
    },
    {
        navItems: [{ content: themeItemContent, itemAs: 'div', key: 6 }],
    },
    {
        headerContent: 'Your organizations',
        navItems: [
            {
                content: 'Sourcegraph',
                to: '/users/ralph/settings/organizations',
                key: 'sourcegraph-123456',
            },
            {
                content: 'Gitstart',
                to: '/users/ralph/settings/repositories',
                key: 'gitstart-123456',
            },
        ],
    },
    {
        navItems: [
            {
                content: 'Site admin',
                onSelect: noop,
                key: 'admin',
            },
            {
                content: 'Help',
                onSelect: noop,
                suffixIcon: OpenInNewIcon,
                key: 'Help',
            },
            {
                content: 'Keyboard shortcuts',
                onSelect: noop,
                key: 'shortcuts',
            },
        ],
    },
    {
        navItems: [
            {
                content: 'About Sourcegraph',
                suffixIcon: OpenInNewIcon,
                to: 'https://sourcegraph.com',
                key: 'Sourcegraph',
            },
            {
                content: 'Browser Extension',
                suffixIcon: OpenInNewIcon,
                to: 'https://sourcegraph.com/docs/integration/browser_extension',
                key: 'Extension',
            },
        ],
    },
]

export const UserNav: StoryFn = () => (
    <NavMenu
        navTrigger={{
            variant: 'icon',
            triggerContent: {
                trigger: isOpen => (
                    <>
                        <Icon aria-hidden={true} as="img" className={styles.avatar} src={avatarUrl} />
                        <Icon aria-hidden={true} svgPath={isOpen ? mdiMenuUp : mdiMenuDown} />
                    </>
                ),
            },
        }}
        sections={navItems}
    />
)

const singleSectionNavItems: NavMenuSectionProps[] = [
    {
        navItems: [
            { content: 'Batch Changes', prefixIcon: FileTreeOutlineIcon, key: 'Batch Changes' },
            {
                content: (
                    <Button variant="link" className="w-100 text-left">
                        <Icon aria-hidden={true} svgPath={mdiPoll} /> Insight
                    </Button>
                ),
                key: 'Insight',
            },
            {
                content: (
                    <Button variant="link" className="w-100 text-left">
                        <Icon aria-hidden={true} svgPath={mdiAntenna} /> Monitoring
                    </Button>
                ),
                key: 'Monitoring',
            },
        ],
    },
]

export const SingleSectionNavMenuExample: StoryFn = () => (
    <NavMenu
        navTrigger={{
            variant: 'secondary',
            outline: true,
            triggerContent: {
                trigger: isOpen => (
                    <>
                        <Icon aria-hidden={true} svgPath={mdiMenu} />
                        <Icon aria-hidden={true} svgPath={isOpen ? mdiMenuUp : mdiMenuDown} />
                    </>
                ),
            },
        }}
        sections={singleSectionNavItems}
    />
)

const avatarUrl =
    'data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iaXNvLTg4NTktMSI/Pg0KPCEtLSBHZW5lcmF0b3I6IEFkb2JlIElsbHVzdHJhdG9yIDE5LjAuMCwgU1ZHIEV4cG9ydCBQbHVnLUluIC4gU1ZHIFZlcnNpb246IDYuMDAgQnVpbGQgMCkgIC0tPg0KPHN2ZyB2ZXJzaW9uPSIxLjEiIGlkPSJMYXllcl8xIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB4PSIwcHgiIHk9IjBweCINCgkgdmlld0JveD0iMCAwIDE0NSAxNDUiIHN0eWxlPSJlbmFibGUtYmFja2dyb3VuZDpuZXcgMCAwIDE0NSAxNDU7IiB4bWw6c3BhY2U9InByZXNlcnZlIj4NCjxnIGlkPSJtZW5fNSI+DQoJPHJlY3Qgc3R5bGU9ImZpbGw6IzVBQkNCQjsiIHdpZHRoPSIxNDUiIGhlaWdodD0iMTQ1Ii8+DQoJPGc+DQoJCTxnPg0KCQkJPGc+DQoJCQkJPGc+DQoJCQkJCTxwYXRoIHN0eWxlPSJmaWxsOiNGMUM5QTU7IiBkPSJNMTA5LjM3NCwxMTUuMzk0Yy00Ljk2NC05LjM5Ni0zNi44NzUtMTUuMjkyLTM2Ljg3NS0xNS4yOTJzLTMxLjkxLDUuODk2LTM2Ljg3NCwxNS4yOTINCgkJCQkJCUMzMS45NTcsMTI4LjQzMywyOC44ODgsMTQ1LDI4Ljg4OCwxNDVoNDMuNjExaDQzLjYxMkMxMTYuMTEyLDE0NSwxMTQuMDQsMTI3LjIzNiwxMDkuMzc0LDExNS4zOTR6Ii8+DQoJCQkJCTxwYXRoIHN0eWxlPSJmaWxsOiNFNEI2OTI7IiBkPSJNNzIuNDk5LDEwMC4xMDJjMCwwLDMxLjkxMSw1Ljg5NiwzNi44NzUsMTUuMjkyYzQuNjY1LDExLjg0Miw2LjczNywyOS42MDYsNi43MzcsMjkuNjA2SDcyLjQ5OQ0KCQkJCQkJVjEwMC4xMDJ6Ii8+DQoJCQkJCTxyZWN0IHg9IjYzLjgxMyIgeT0iODEiIHN0eWxlPSJmaWxsOiNGMUM5QTU7IiB3aWR0aD0iMTcuMzc0IiBoZWlnaHQ9IjI5LjA3NyIvPg0KCQkJCQk8cmVjdCB4PSI3Mi40OTkiIHk9IjgxIiBzdHlsZT0iZmlsbDojRTRCNjkyOyIgd2lkdGg9IjguNjg4IiBoZWlnaHQ9IjI5LjA3NyIvPg0KCQkJCQk8cGF0aCBzdHlsZT0ib3BhY2l0eTowLjE7ZmlsbDojRERBQzhDO2VuYWJsZS1iYWNrZ3JvdW5kOm5ldyAgICA7IiBkPSJNNjMuODEzLDk0LjQ3NGMxLjU2Myw0LjQ4NSw3Ljg2OCw3LjA1NywxMi40OTksNy4wNTcNCgkJCQkJCWMxLjY3NiwwLDMuMzA2LTAuMjgxLDQuODc1LTAuNzk1VjgxSDYzLjgxM1Y5NC40NzR6Ii8+DQoJCQkJCTxwYXRoIHN0eWxlPSJmaWxsOiNGMUM5QTU7IiBkPSJNOTQuODM3LDYyLjY1M2MwLTE4LjE2Mi0xMC4wMDEtMjguNDg5LTIyLjMzOC0yOC40ODljLTEyLjMzNiwwLTIyLjMzNywxMC4zMjctMjIuMzM3LDI4LjQ4OQ0KCQkJCQkJYzAsMjQuNDI4LDEwLjAwMSwzMi44ODYsMjIuMzM3LDMyLjg4NkM4NC44MzcsOTUuNTM5LDk0LjgzNyw4Ni4wNjMsOTQuODM3LDYyLjY1M3oiLz4NCgkJCQkJPHBhdGggc3R5bGU9ImZpbGw6I0YxQzlBNTsiIGQ9Ik05NC44MzcsNjIuNjUzYzAtMTguMTYyLTEwLjAwMS0yOC40ODktMjIuMzM4LTI4LjQ4OWMtMTIuMzM2LDAtMjIuMzM3LDEwLjMyNy0yMi4zMzcsMjguNDg5DQoJCQkJCQljMCwyNC40MjgsMTAuMDAxLDMyLjg4NiwyMi4zMzcsMzIuODg2Qzg0LjgzNyw5NS41MzksOTQuODM3LDg2LjA2Myw5NC44MzcsNjIuNjUzeiIvPg0KCQkJCQk8cGF0aCBzdHlsZT0iZmlsbDojRjFDOUE1OyIgZD0iTTQ1LjE2MSw2Ny4wMzFjLTAuNjg0LTQuOTU3LDIuMDQ2LTkuMzE4LDYuMDkyLTkuNzRjNC4wNTMtMC40MjIsNy44ODgsMy4yNTksOC41NjcsOC4yMTYNCgkJCQkJCWMwLjY4Myw0Ljk1My0yLjA1Myw5LjMxNS02LjEsOS43MzlDNDkuNjcxLDc1LjY2NSw0NS44NCw3MS45ODgsNDUuMTYxLDY3LjAzMXoiLz4NCgkJCQkJPHBhdGggc3R5bGU9ImZpbGw6I0U0QjY5MjsiIGQ9Ik05NC44MzcsNjIuNjUzYzAtMTguMTYyLTEwLjAwMS0yOC40ODktMjIuMzM4LTI4LjQ4OXY2MS4zNzUNCgkJCQkJCUM4NC44MzcsOTUuNTM5LDk0LjgzNyw4Ni4wNjMsOTQuODM3LDYyLjY1M3oiLz4NCgkJCQkJPHBhdGggc3R5bGU9ImZpbGw6IzEwMkY0MTsiIGQ9Ik0xMDkuMzc0LDExNS4zOTRjLTMuMTgxLTYuMDIxLTE3LjQxOC0xMC42MDEtMjcuMjQyLTEzLjExNw0KCQkJCQkJYy0wLjM4Miw0Ljk5LTQuNTQ1LDguOTIzLTkuNjMzLDguOTIzYy01LjA4OCwwLTkuMjUtMy45MzMtOS42MzItOC45MjNjLTkuODI0LDIuNTE2LTI0LjA2MSw3LjA5Ni0yNy4yNDIsMTMuMTE3DQoJCQkJCQlDMzEuOTU3LDEyOC40MzMsMjguODg4LDE0NSwyOC44ODgsMTQ1aDQzLjYxMWg0My42MTJDMTE2LjExMiwxNDUsMTE0LjA0LDEyNy4yMzYsMTA5LjM3NCwxMTUuMzk0eiIvPg0KCQkJCTwvZz4NCgkJCTwvZz4NCgkJPC9nPg0KCQk8cGF0aCBzdHlsZT0iZmlsbDojMjMxRjIwOyIgZD0iTTUzLjk0MSw4NC4yN2M0Ljg1OSw4LjI1Miw5LjY5OCw5LjUyOCw5LjY5OCw5LjUyOGwxLjQ2Niw0Ljc1NUg2M2gtMS42NWwtMC45MTUtMi4wNTFINTguNDINCgkJCWwtMC43MzMtMi41MTdsLTEuOTk2LDAuODM5bC0wLjQ3NS0yLjc5N2gtMS44MzNsLTAuNjQxLTIuMzMxbC0xLjM3NCwwLjY1M2wtMS4zNzQtMi4xNDVsLTEuOTIzLTEuMDI1bDAuOTE2LTMuMTdoLTIuMjg5DQoJCQlsMC4wOTItMy4zNTdsLTEuNzQxLDAuNTZsLTAuNjQxLTIuOTgzbC0yLjkzMS0xLjExOWwyLjU2NC0yLjYxMWwtMS4yODItMC40NjZsMS4zNzMtMi40MjNsLTEuMDA3LTAuNDY2bC0wLjE4NC0yLjg5MWwtMi45My0wLjU1OQ0KCQkJbDEuNjQ5LTIuNjExbC0yLjAxNi0zLjE3bDIuMzgyLTEuMzA2bC0xLjI4Mi0zLjE2OWwwLjczMi0yLjk4NGwtMi43NDctMi4yMzhsMy44NDYtMC4zNzJ2LTIuNzk3bDEuMS0yLjQyNGwtMS4xLTIuMDUxDQoJCQlsMi4wMTYtMS42NzhsLTIuMDE2LTEuNjc5bDIuMTk4LTIuMjM4di0xLjMwNWwyLjAxNi0xLjY3OGwtMC45MTYtMi42MTFsMy42NjIsMC43NDZsMC4xODQtNC40NzVsMi43NDgsMC45MzJsMC45MTUtMS44NjQNCgkJCWwzLjExNCwwLjM3M2MwLDAtMC43MzItMS40OTIsMC41NDktMS40OTJjMS4yODIsMCw0LjM5NywxLjMwNiw0LjM5NywxLjMwNmwxLjY0OC0yLjIzOGwzLjY2MywyLjYxMWwxLjI4My0yLjYxMWwyLjM4MSwyLjc5Nw0KCQkJbDIuMjM0LTIuNzk3YzAsMCwwLDE1LjY2NCwwLDI1LjM2Yy0yLjg5NS0wLjI0OC00LjQzMy0wLjI0OC0xMC43ODItMi43MzVjLTIuMTk3LDEuNDkyLTkuMjgsMTEuNjg2LTkuMjgsMTIuNjgNCgkJCXMtMS4yMjIsMTEuMTg5LTEuMjIyLDExLjE4OWwtMS4yNzQtNi4wOEM1MC4wOTMsNjcuMDgxLDUxLjM2Myw3OS44OTEsNTMuOTQxLDg0LjI3eiIvPg0KCQk8cGF0aCBzdHlsZT0iZmlsbDojMjMxRjIwOyIgZD0iTTkwLjYyOSw4NC4yN2MtNC42NjksOC4yNTItOS4zMTgsOS41MjgtOS4zMTgsOS41MjhsLTEuNDA4LDQuNzU1aDIuMDIyaDEuNTg1bDAuODgtMi4wNTFoMS45MzcNCgkJCWwwLjcwMy0yLjUxN2wxLjkxOSwwLjgzOWwwLjQ1Ny0yLjc5N2gxLjc2MWwwLjYxNS0yLjMzMWwxLjMyLDAuNjUzbDEuMzItMi4xNDVsMS44NDgtMS4wMjVsLTAuODgtMy4xN2gyLjE5OWwtMC4wODgtMy4zNTcNCgkJCWwxLjY3MywwLjU2bDAuNjE1LTIuOTgzbDIuODE2LTEuMTE5bC0yLjQ2NS0yLjYxMWwxLjIzMy0wLjQ2NmwtMS4zMi0yLjQyM2wwLjk2OC0wLjQ2NmwwLjE3Ny0yLjg5MWwyLjgxNC0wLjU1OWwtMS41ODQtMi42MTENCgkJCWwxLjkzNy0zLjE3bC0yLjI4OC0xLjMwNmwxLjIzMS0zLjE2OWwtMC43MDMtMi45ODRsMi42NC0yLjIzOGwtMy42OTUtMC4zNzJ2LTIuNzk3bC0xLjA1Ny0yLjQyNGwxLjA1Ny0yLjA1MWwtMS45MzctMS42NzgNCgkJCWwxLjkzNy0xLjY3OWwtMi4xMTItMi4yMzh2LTEuMzA1bC0xLjkzNi0xLjY3OGwwLjg4MS0yLjYxMWwtMy41MiwwLjc0NmwtMC4xNzctNC40NzVsLTIuNjQsMC45MzJsLTAuODgtMS44NjRsLTIuOTkyLDAuMzczDQoJCQljMCwwLDAuNzA0LTEuNDkyLTAuNTI3LTEuNDkyYy0xLjIzMiwwLTQuMjI1LDEuMzA2LTQuMjI1LDEuMzA2bC0xLjU4NC0yLjIzOGwtMy41MiwyLjYxMWwtMS4yMzItMi42MTFsLTIuMjg4LDIuNzk3bC0yLjE0Ni0yLjc5Nw0KCQkJYzAsMCwwLDE1LjY2NCwwLDI1LjM2YzIuNzgtMC4yNDgsNC4yNTktMC4yNDgsMTAuMzU5LTIuNzM1YzIuMTEyLDEuNDkyLDguOTE3LDExLjY4Niw4LjkxNywxMi42OHMxLjE3NCwxMS4xODksMS4xNzQsMTEuMTg5DQoJCQlsMS4yMjUtNi4wOEM5NC4zMjcsNjcuMDgxLDkzLjEwNyw3OS44OTEsOTAuNjI5LDg0LjI3eiIvPg0KCQk8cGF0aCBzdHlsZT0iZmlsbDojRjFDOUE1OyIgZD0iTTQ1LjE2MSw2Ny4wMzFjLTAuNjg0LTQuOTU3LDIuMDQ2LTkuMzE4LDYuMDkyLTkuNzRjNC4wNTMtMC40MjIsNy44ODgsMy4yNTksOC41NjcsOC4yMTYNCgkJCWMwLjY4Myw0Ljk1My0yLjA1Myw5LjMxNS02LjEsOS43MzlDNDkuNjcxLDc1LjY2NSw0NS44NCw3MS45ODgsNDUuMTYxLDY3LjAzMXoiLz4NCgkJPHBhdGggc3R5bGU9ImZpbGw6I0U0QjY5MjsiIGQ9Ik05MS40MzgsNzUuMjQ2Yy00LjA1LTAuNDI0LTYuNzgzLTQuNzg3LTYuMDk4LTkuNzM5YzAuNjc3LTQuOTU3LDQuNTEzLTguNjM4LDguNTYzLTguMjE2DQoJCQljNC4wNDcsMC40MjIsNi43NzcsNC43ODMsNi4wOTQsOS43NEM5OS4zMTgsNzEuOTg4LDk1LjQ4Nyw3NS42NjUsOTEuNDM4LDc1LjI0NnoiLz4NCgk8L2c+DQo8L2c+DQo8Zz4NCjwvZz4NCjxnPg0KPC9nPg0KPGc+DQo8L2c+DQo8Zz4NCjwvZz4NCjxnPg0KPC9nPg0KPGc+DQo8L2c+DQo8Zz4NCjwvZz4NCjxnPg0KPC9nPg0KPGc+DQo8L2c+DQo8Zz4NCjwvZz4NCjxnPg0KPC9nPg0KPGc+DQo8L2c+DQo8Zz4NCjwvZz4NCjxnPg0KPC9nPg0KPGc+DQo8L2c+DQo8L3N2Zz4NCg=='

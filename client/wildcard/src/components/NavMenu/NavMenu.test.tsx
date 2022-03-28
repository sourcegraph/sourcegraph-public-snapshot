import React from 'react'

import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import sinon from 'sinon'

import { AnchorLink } from '../Link'

import { NavMenuSectionProps } from './NavMenu'

import { NavMenu } from '.'

describe('<NavMenu />', () => {
    it('Should render Menu Items Correctly', () => {
        const onSelect = sinon.spy(() => undefined)
        const menuNavItems: NavMenuSectionProps[] = [
            {
                headerContent: <> Signed in as @ralph</>,
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
                ],
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
                        itemContent: 'Keyboard shortcuts',
                        onSelect,
                    },
                    {
                        itemContent: 'Help',
                        disabled: true,
                        onSelect,
                    },
                ],
            },
            {
                navItems: [
                    {
                        itemContent: 'About Sourcegraph',
                        itemAs: AnchorLink,
                        to: 'https://about.sourcegraph.com',
                    },
                    {
                        itemContent: 'Browser Extension',
                        itemAs: AnchorLink,
                        to: 'https://docs.sourcegraph.com/integration/browser_extension',
                    },
                ],
            },
        ]

        render(<NavMenu navTrigger={{ content: 'menu trigger' }} sections={menuNavItems} />)
        const button = screen.getByRole('button', { name: 'menu trigger' })
        expect(button).toBeVisible()
        userEvent.click(button)

        expect(document.body).toMatchSnapshot()
        for (const navItem of menuNavItems) {
            const { headerContent, navItems = [] } = navItem

            if (headerContent && typeof headerContent === 'string') {
                expect(screen.getByText(headerContent)).toBeVisible()
            }

            for (const { itemContent } of navItems) {
                expect(screen.getByText(itemContent as string)).toBeVisible()
            }
        }
    })
})

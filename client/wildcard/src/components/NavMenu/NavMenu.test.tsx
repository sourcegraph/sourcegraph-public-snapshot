import { describe, expect, it } from '@jest/globals'
import userEvent from '@testing-library/user-event'
import sinon from 'sinon'

import { renderWithBrandedContext } from '../../testing'
import { AnchorLink } from '../Link'

import { NavMenu } from '.'
import type { NavMenuSectionProps } from './NavMenu'

describe('<NavMenu />', () => {
    it('Should render Menu Items Correctly', () => {
        const onSelect = sinon.spy(() => undefined)
        const menuNavItems: NavMenuSectionProps[] = [
            {
                headerContent: <> Signed in as @ralph</>,
                navItems: [
                    {
                        content: 'Settings',
                        to: '/users/ralph/settings/organizations',
                        key: 1,
                    },
                    {
                        content: 'Your repositories',
                        to: '/users/ralph/settings/repositories',
                        key: 2,
                    },
                    {
                        content: 'Saved searches',
                        to: '/users/ralph/searches',
                        key: 3,
                    },
                ],
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
                        content: 'Keyboard shortcuts',
                        onSelect,
                        key: 4,
                    },
                    {
                        content: 'Help',
                        disabled: true,
                        onSelect,
                        key: 5,
                    },
                ],
            },
            {
                navItems: [
                    {
                        content: 'About Sourcegraph',
                        itemAs: AnchorLink,
                        to: 'https://about.sourcegraph.com',
                        key: 6,
                    },
                    {
                        content: 'Browser Extension',
                        itemAs: AnchorLink,
                        to: 'https://docs.sourcegraph.com/integration/browser_extension',
                        key: 7,
                    },
                ],
            },
        ]

        const result = renderWithBrandedContext(
            <NavMenu navTrigger={{ triggerContent: { text: 'menu trigger' } }} sections={menuNavItems} />
        )
        const button = result.getByRole('button', { name: 'menu trigger' })
        expect(button).toBeVisible()
        userEvent.click(button)

        expect(document.body).toMatchSnapshot()
        for (const navItem of menuNavItems) {
            const { headerContent, navItems = [] } = navItem

            if (headerContent && typeof headerContent === 'string') {
                expect(result.getByText(headerContent)).toBeInTheDocument()
            }

            for (const { content } of navItems) {
                expect(result.getByText(content as string)).toBeInTheDocument()
            }
        }
    })
})

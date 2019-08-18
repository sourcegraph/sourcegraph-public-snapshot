import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import * as H from 'history'
import React, { useCallback } from 'react'
import { MemoryRouter } from 'react-router'
import * as GQL from '../../../shared/src/graphql/schema'
import { ThemePreference } from '../theme'
import { UserNavItem } from './UserNavItem'

import './UserNavItem.scss'

const onThemePreferenceChange = action('onThemePreferenceChange')

const { add } = storiesOf('UserNavItem', module).addDecorator(story => (
    <div className="theme-light" style={{ display: 'inline-block', position: 'relative', margin: '2rem' }}>
        {story()}
    </div>
))

const OpenUserNavItem: React.FunctionComponent<UserNavItem['props']> = props => {
    const openDropdown = useCallback((e: UserNavItem | null) => {
        if (e) {
            e.setState({ isOpen: true })
        }
    }, [])
    return <UserNavItem {...props} ref={openDropdown} />
}

add('Site admin', () => (
    <MemoryRouter>
        <OpenUserNavItem
            authenticatedUser={{
                username: 'alice',
                avatarURL: null,
                session: { __typename: 'Session', canSignOut: true },
                settingsURL: '#',
                siteAdmin: true,
                organizations: {
                    __typename: 'OrgConnection',
                    totalCount: 3,
                    nodes: [
                        { id: '0', settingsURL: '#', displayName: 'Acme Corp' },
                        { id: '1', settingsURL: '#', displayName: 'Beta Inc' },
                    ] as GQL.IOrg[],
                },
            }}
            isLightTheme={true}
            themePreference={ThemePreference.Light}
            location={H.createMemoryHistory().location}
            onThemePreferenceChange={onThemePreferenceChange}
            showDiscussions={true}
            showDotComMarketing={true}
        />
    </MemoryRouter>
))

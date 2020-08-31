import { action } from '@storybook/addon-actions'
import { boolean } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import React, { useCallback } from 'react'
import * as GQL from '../../../shared/src/graphql/schema'
import { ThemePreference } from '../theme'
import { UserNavItem } from './UserNavItem'
import { WebStory } from '../components/WebStory'

const onThemePreferenceChange = action('onThemePreferenceChange')

const { add } = storiesOf('web/UserNavItem', module)

const OpenUserNavItem: React.FunctionComponent<UserNavItem['props']> = props => {
    const openDropdown = useCallback((userNavItem: UserNavItem | null) => {
        if (userNavItem) {
            userNavItem.setState({ isOpen: true })
        }
    }, [])
    return <UserNavItem {...props} ref={openDropdown} />
}

add(
    'Site admin',
    () => (
        <WebStory>
            {webProps => (
                <OpenUserNavItem
                    {...webProps}
                    authenticatedUser={{
                        username: 'alice',
                        avatarURL: null,
                        session: { canSignOut: true },
                        settingsURL: '#',
                        siteAdmin: true,
                        organizations: {
                            nodes: [
                                { id: '0', settingsURL: '#', displayName: 'Acme Corp' },
                                { id: '1', settingsURL: '#', displayName: 'Beta Inc' },
                            ] as GQL.IOrg[],
                        },
                    }}
                    themePreference={webProps.isLightTheme ? ThemePreference.Light : ThemePreference.Dark}
                    onThemePreferenceChange={onThemePreferenceChange}
                    showDotComMarketing={boolean('showDotComMarketing', true)}
                />
            )}
        </WebStory>
    ),
    {
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/HWLuLefEdev5KYtoEGHjFj/Sourcegraph-Components-Contractor?node-id=1346%3A0',
        },
    }
)

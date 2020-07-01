import * as H from 'history'
import { storiesOf } from '@storybook/react'
import React from 'react'
import webMainStyles from '../SourcegraphWebApp.scss'
import { subtypeOf } from '../../../shared/src/util/types'
import { SearchPatternType } from '../../../shared/src/graphql/schema'
import { action } from '@storybook/addon-actions'
import { RepogroupPage } from './RepogroupPage'
import { python2To3Metadata } from './Python2To3'

const { add } = storiesOf('web/VersionContextDropdown', module).addDecorator(story => (
    <>
        <style>{webMainStyles}</style>
        <div className="theme-light">{story()}</div>
    </>
))

const setVersionContext = action('setVersionContext')
const history = H.createMemoryHistory({ keyLength: 0 })
const commonProps = subtypeOf<Partial<VersionContextDropdownProps>>()({
    alwaysShow: true,
    history,
    // Make sure the dropdown is not rendered outside the theme-light container
    portal: false,
    caseSensitive: false,
    patternType: SearchPatternType.literal,
    availableVersionContexts: [
        { name: 'test 1', description: 'test 1', revisions: [{ rev: 'test', repo: 'github.com/test/test' }] },
        { name: 'test 2', description: 'test 2', revisions: [{ rev: 'test', repo: 'github.com/test/test' }] },
        { name: 'test 3', description: 'test 3', revisions: [{ rev: 'test', repo: 'github.com/test/test' }] },
    ],
    navbarSearchQuery: 'test',
    setVersionContext,
})

add('Default', () => (
    <RepogroupPage
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
        showDotComMarketing={true}
        repogroupMetadata={python2To3Metadata}
    />
))

import * as H from 'history'
import { storiesOf } from '@storybook/react'
import { RepogroupPage, RepogroupPageProps } from './RepogroupPage'
import React from 'react'
import { python2To3Metadata } from './Python2To3'
import * as GQL from '../../../shared/src/graphql/schema'
import { NEVER } from 'rxjs'
import { NOOP_SETTINGS_CASCADE } from '../../../shared/src/util/searchTestHelpers'
import sinon from 'sinon'
import { ThemePreference } from '../theme'
import { NOOP_TELEMETRY_SERVICE } from '../../../shared/src/telemetry/telemetryService'
import { ActionItemComponentProps } from '../../../shared/src/actions/ActionItem'
import { Services } from '../../../shared/src/api/client/services'
import { MemoryRouter } from 'react-router'
import webStyles from '../SourcegraphWebApp.scss'
import { AuthenticatedUser } from '../auth'
import { SearchPatternType } from '../graphql-operations'

const { add } = storiesOf('web/RepogroupPage', module)
    .addParameters({
        percy: { widths: [993] },
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/Xc4M24VTQq8itU0Lgb1Wwm/RFC-159-Visual-Design?node-id=66%3A611',
        },
        chromatic: { viewports: [769, 993, 1200] },
    })
    .addDecorator(story => (
        <>
            <style>{webStyles}</style>
            <div className="theme-light">{story()}</div>
        </>
    ))

const history = H.createMemoryHistory()

const EXTENSIONS_CONTROLLER: ActionItemComponentProps['extensionsController'] = {
    executeCommand: () => new Promise(resolve => setTimeout(resolve, 750)),
}

const PLATFORM_CONTEXT: ActionItemComponentProps['platformContext'] = {
    forceUpdateTooltip: () => undefined,
    settings: NEVER,
}

const authUser: AuthenticatedUser = {
    __typename: 'User',
    id: '0',
    email: 'alice@sourcegraph.com',
    username: 'alice',
    avatarURL: null,
    session: { canSignOut: true },
    displayName: null,
    url: '',
    settingsURL: '#',
    siteAdmin: true,
    organizations: {
        nodes: [
            { id: '0', settingsURL: '#', displayName: 'Acme Corp' },
            { id: '1', settingsURL: '#', displayName: 'Beta Inc' },
        ] as GQL.IOrg[],
    },
    tags: [],
    viewerCanAdminister: true,
    databaseID: 0,
}

const commonProps: RepogroupPageProps = {
    settingsCascade: {
        ...NOOP_SETTINGS_CASCADE,
        subjects: [],
        final: {
            'search.repositoryGroups': {
                python: [
                    'github.com/python/test',
                    'github.com/python/test2',
                    'github.com/python/test3',
                    'github.com/python/test4',
                ],
            },
        },
    },
    isLightTheme: true,
    themePreference: ThemePreference.Light,
    onThemePreferenceChange: sinon.spy(() => {}),
    patternType: SearchPatternType.literal,
    setPatternType: sinon.spy(() => {}),
    caseSensitive: false,
    copyQueryButton: false,
    extensionsController: { ...EXTENSIONS_CONTROLLER, services: {} as Services },
    platformContext: PLATFORM_CONTEXT,
    filtersInQuery: {},
    history,
    interactiveSearchMode: false,
    keyboardShortcuts: [],
    onFiltersInQueryChange: sinon.spy(() => {}),
    setCaseSensitivity: sinon.spy(() => {}),
    splitSearchModes: true,
    telemetryService: NOOP_TELEMETRY_SERVICE,
    toggleSearchMode: sinon.spy(() => {}),
    versionContext: undefined,
    activation: undefined,
    location: history.location,
    isSourcegraphDotCom: true,
    setVersionContext: sinon.spy(() => {}),
    availableVersionContexts: [],
    authRequired: false,
    showCampaigns: false,
    authenticatedUser: authUser,
    repogroupMetadata: python2To3Metadata,
    autoFocus: false,
    globbing: false,
    showOnboardingTour: false,
}

add('Repogroup page with smart search field', () => (
    <MemoryRouter>
        <RepogroupPage {...commonProps} />
    </MemoryRouter>
))

add('Repogroup page without smart search field', () => (
    <MemoryRouter>
        <RepogroupPage {...commonProps} />
    </MemoryRouter>
))

import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import { subDays } from 'date-fns'
import React from 'react'
import { NEVER, Observable, of } from 'rxjs'

import { ActionItemComponentProps } from '@sourcegraph/shared/src/actions/ActionItem'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { IRepository, ISearchContext, ISearchContextRepositoryRevisions } from '@sourcegraph/shared/src/graphql/schema'
import {
    mockFetchAutoDefinedSearchContexts,
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { NOOP_SETTINGS_CASCADE } from '@sourcegraph/shared/src/util/searchTestHelpers'
import { subtypeOf } from '@sourcegraph/shared/src/util/types'
import { WebStory } from '@sourcegraph/web/src/components/WebStory'

import { AuthenticatedUser } from '../auth'
import { SearchPatternType } from '../graphql-operations'
import { ThemePreference } from '../theme'

import { cncf } from './cncf'
import { CommunitySearchContextPage, CommunitySearchContextPageProps } from './CommunitySearchContextPage'
import { temporal } from './Temporal'

const { add } = storiesOf('web/CommunitySearchContextPage', module).addParameters({
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/Xc4M24VTQq8itU0Lgb1Wwm/RFC-159-Visual-Design?node-id=66%3A611',
    },
    chromatic: { viewports: [769, 1200] },
})

const EXTENSIONS_CONTROLLER: ActionItemComponentProps['extensionsController'] = {
    executeCommand: () => new Promise(resolve => setTimeout(resolve, 750)),
}

const PLATFORM_CONTEXT: CommunitySearchContextPageProps['platformContext'] = {
    forceUpdateTooltip: () => undefined,
    settings: NEVER,
    sourcegraphURL: '',
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

const repositories: ISearchContextRepositoryRevisions[] = [
    {
        __typename: 'SearchContextRepositoryRevisions',
        repository: {
            __typename: 'Repository',
            name: 'github.com/example/example2',
        } as IRepository,
        revisions: ['main'],
    },
    {
        __typename: 'SearchContextRepositoryRevisions',
        repository: {
            __typename: 'Repository',
            name: 'github.com/example/example1',
        } as IRepository,
        revisions: ['main'],
    },
]

const fetchCommunitySearchContext = (): Observable<ISearchContext> =>
    of({
        __typename: 'SearchContext',
        id: '1',
        spec: 'public-ctx',
        name: 'public-ctx',
        namespace: null,
        public: true,
        autoDefined: false,
        description: 'Repositories on Sourcegraph',
        repositories,
        updatedAt: subDays(new Date(), 1).toISOString(),
        viewerCanManage: true,
    })

const commonProps = () =>
    subtypeOf<Partial<CommunitySearchContextPageProps>>()({
        settingsCascade: NOOP_SETTINGS_CASCADE,
        onThemePreferenceChange: action('onThemePreferenceChange'),
        parsedSearchQuery: 'r:golang/oauth2 test f:travis',
        patternType: SearchPatternType.literal,
        setPatternType: action('setPatternType'),
        caseSensitive: false,
        extensionsController: { ...EXTENSIONS_CONTROLLER },
        platformContext: PLATFORM_CONTEXT,
        keyboardShortcuts: [],
        setCaseSensitivity: action('setCaseSensitivity'),
        activation: undefined,
        isSourcegraphDotCom: true,
        showSearchContext: true,
        searchContextsEnabled: true,
        showSearchContextManagement: false,
        selectedSearchContextSpec: '',
        setSelectedSearchContextSpec: () => {},
        defaultSearchContextSpec: '',
        authRequired: false,
        batchChangesEnabled: false,
        authenticatedUser: authUser,
        communitySearchContextMetadata: temporal,
        globbing: false,
        showOnboardingTour: false,
        fetchAutoDefinedSearchContexts: mockFetchAutoDefinedSearchContexts(),
        fetchSearchContexts: mockFetchSearchContexts,
        hasUserAddedRepositories: false,
        hasUserAddedExternalServices: false,
        getUserSearchContextNamespaces: mockGetUserSearchContextNamespaces,
        fetchSearchContextBySpec: fetchCommunitySearchContext,
    })

add('Temporal', () => (
    <WebStory>
        {webProps => (
            <CommunitySearchContextPage
                {...webProps}
                {...commonProps()}
                themePreference={webProps.isLightTheme ? ThemePreference.Light : ThemePreference.Dark}
            />
        )}
    </WebStory>
))

add('CNCF', () => (
    <WebStory>
        {webProps => (
            <CommunitySearchContextPage
                {...webProps}
                {...commonProps()}
                communitySearchContextMetadata={cncf}
                themePreference={webProps.isLightTheme ? ThemePreference.Light : ThemePreference.Dark}
            />
        )}
    </WebStory>
))

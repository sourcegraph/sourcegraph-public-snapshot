import { action } from '@storybook/addon-actions'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import { subDays } from 'date-fns'
import { EMPTY, NEVER, Observable, of } from 'rxjs'

import { subtypeOf } from '@sourcegraph/common'
import { SearchContextFields } from '@sourcegraph/search'
import { ActionItemComponentProps } from '@sourcegraph/shared/src/actions/ActionItem'
import {
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { NOOP_SETTINGS_CASCADE } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { AuthenticatedUser } from '../auth'
import { WebStory } from '../components/WebStory'
import { SearchPatternType } from '../graphql-operations'
import { useExperimentalFeatures } from '../stores'
import { ThemePreference } from '../theme'

import { cncf } from './cncf'
import { CommunitySearchContextPage, CommunitySearchContextPageProps } from './CommunitySearchContextPage'
import { temporal } from './Temporal'

const decorator: DecoratorFn = Story => {
    useExperimentalFeatures.setState({ showSearchContext: true })
    return <Story />
}

const config: Meta = {
    title: 'web/CommunitySearchContextPage',
    parameters: {
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/Xc4M24VTQq8itU0Lgb1Wwm/RFC-159-Visual-Design?node-id=66%3A611',
        },
        chromatic: { viewports: [769, 1200] },
    },
    decorators: [decorator],
}

export default config

const EXTENSIONS_CONTROLLER: ActionItemComponentProps['extensionsController'] = {
    executeCommand: () => new Promise(resolve => setTimeout(resolve, 750)),
}

const PLATFORM_CONTEXT: CommunitySearchContextPageProps['platformContext'] = {
    settings: NEVER,
    sourcegraphURL: '',
    requestGraphQL: () => EMPTY,
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
        ] as AuthenticatedUser['organizations']['nodes'],
    },
    tags: [],
    viewerCanAdminister: true,
    databaseID: 0,
    tosAccepted: true,
    searchable: true,
    emails: [],
}

const repositories: SearchContextFields['repositories'] = [
    {
        __typename: 'SearchContextRepositoryRevisions',
        repository: {
            __typename: 'Repository',
            name: 'github.com/example/example2',
        },
        revisions: ['main'],
    },
    {
        __typename: 'SearchContextRepositoryRevisions',
        repository: {
            __typename: 'Repository',
            name: 'github.com/example/example1',
        },
        revisions: ['main'],
    },
]

const fetchCommunitySearchContext = (): Observable<SearchContextFields> =>
    of({
        __typename: 'SearchContext',
        id: '1',
        spec: 'public-ctx',
        name: 'public-ctx',
        namespace: null,
        public: true,
        autoDefined: false,
        description: 'Repositories on Sourcegraph',
        query: '',
        repositories,
        updatedAt: subDays(new Date(), 1).toISOString(),
        viewerCanManage: true,
        viewerHasAsDefault: false,
        viewerHasStarred: false,
    })

const commonProps = () =>
    subtypeOf<Partial<CommunitySearchContextPageProps>>()({
        settingsCascade: NOOP_SETTINGS_CASCADE,
        onThemePreferenceChange: action('onThemePreferenceChange'),
        parsedSearchQuery: 'r:golang/oauth2 test f:travis',
        patternType: SearchPatternType.standard,
        setPatternType: action('setPatternType'),
        caseSensitive: false,
        extensionsController: { ...EXTENSIONS_CONTROLLER },
        platformContext: PLATFORM_CONTEXT,
        setCaseSensitivity: action('setCaseSensitivity'),
        activation: undefined,
        isSourcegraphDotCom: true,
        searchContextsEnabled: true,
        selectedSearchContextSpec: '',
        setSelectedSearchContextSpec: () => {},
        authRequired: false,
        batchChangesEnabled: false,
        authenticatedUser: authUser,
        communitySearchContextMetadata: temporal,
        globbing: false,
        fetchSearchContexts: mockFetchSearchContexts,
        getUserSearchContextNamespaces: mockGetUserSearchContextNamespaces,
        fetchSearchContextBySpec: fetchCommunitySearchContext,
    })

export const Temporal: Story = () => (
    <WebStory>
        {webProps => (
            <CommunitySearchContextPage
                {...webProps}
                {...commonProps()}
                themePreference={webProps.isLightTheme ? ThemePreference.Light : ThemePreference.Dark}
            />
        )}
    </WebStory>
)

export const CNCFStory: Story = () => (
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
)

CNCFStory.storyName = 'CNCF'

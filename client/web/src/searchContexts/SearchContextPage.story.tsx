import { storiesOf } from '@storybook/react'
import { subDays } from 'date-fns'
import React from 'react'
import { NEVER, Observable, of, throwError } from 'rxjs'

import {
    IRepository,
    ISearchContext,
    ISearchContextRepositoryRevisions,
    SearchPatternType,
} from '@sourcegraph/shared/src/graphql/schema'
import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../components/WebStory'
import { ThemePreference } from '../theme'

import { SearchContextPage } from './SearchContextPage'
import {
    mockFetchAutoDefinedSearchContexts,
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from './testHelpers'

const { add } = storiesOf('web/searchContexts/SearchContextPage', module)
    .addParameters({
        chromatic: { viewports: [1200] },
    })
    .addDecorator(story => <div className="p-3 container">{story()}</div>)

const repositories: ISearchContextRepositoryRevisions[] = [
    {
        __typename: 'SearchContextRepositoryRevisions',
        repository: {
            __typename: 'Repository',
            name: 'github.com/example/example',
        } as IRepository,
        revisions: ['REVISION1', 'REVISION2'],
    },
    {
        __typename: 'SearchContextRepositoryRevisions',
        repository: {
            __typename: 'Repository',
            name: 'github.com/example/really-really-really-really-really-really-long-name',
        } as IRepository,
        revisions: ['REVISION3', 'LONG-LONG-LONG-LONG-LONG-LONG-LONG-LONG-REVISION'],
    },
]

const mockContext: ISearchContext = {
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
}

const props = {
    isMacPlatform: true,
    globbing: true,
    streamSearch: () => NEVER,
    fetchHighlightedFileLineRanges: () => NEVER,
    settingsCascade: EMPTY_SETTINGS_CASCADE,
    isSourcegraphDotCom: false,
    telemetryService: NOOP_TELEMETRY_SERVICE,
    authenticatedUser: null,
    setVersionContext: () => Promise.resolve(undefined),
    availableVersionContexts: [],
    parsedSearchQuery: 'r:golang/oauth2 test f:travis',
    patternType: SearchPatternType.literal,
    setPatternType: () => undefined,
    caseSensitive: false,
    setCaseSensitivity: () => undefined,
    platformContext: {} as any,
    keyboardShortcuts: [],
    versionContext: undefined,
    showSearchContext: false,
    showSearchContextManagement: false,
    selectedSearchContextSpec: '',
    setSelectedSearchContextSpec: () => {},
    defaultSearchContextSpec: '',
    showRepogroupHomepage: false,
    showEnterpriseHomePanels: false,
    showOnboardingTour: false,
    showQueryBuilder: false,
    fetchAutoDefinedSearchContexts: mockFetchAutoDefinedSearchContexts(),
    fetchSearchContexts: mockFetchSearchContexts,
    hasUserAddedRepositories: false,
    hasUserAddedExternalServices: false,
    getUserSearchContextNamespaces: mockGetUserSearchContextNamespaces,
    featureFlags: new Map(),
    themePreference: ThemePreference.Light,
    onThemePreferenceChange: () => undefined,
}

const fetchPublicContext = (): Observable<ISearchContext> => of(mockContext)

const fetchPrivateContext = (): Observable<ISearchContext> =>
    of({
        ...mockContext,
        spec: 'private-ctx',
        name: 'private-ctx',
        namespace: null,
        public: false,
    })

const fetchAutoDefinedContext = (): Observable<ISearchContext> =>
    of({
        ...mockContext,
        autoDefined: true,
    })

add(
    'public context',
    () => (
        <WebStory>
            {webProps => <SearchContextPage {...webProps} {...props} fetchSearchContextBySpec={fetchPublicContext} />}
        </WebStory>
    ),
    {}
)

add(
    'empty description',
    () => (
        <WebStory>
            {webProps => (
                <SearchContextPage
                    {...webProps}
                    {...props}
                    fetchSearchContextBySpec={() =>
                        of({
                            ...mockContext,
                            description: '',
                        })
                    }
                />
            )}
        </WebStory>
    ),
    {}
)

add(
    'public context with search notebook description',
    () => (
        <WebStory>
            {webProps => (
                <SearchContextPage
                    {...webProps}
                    {...props}
                    fetchSearchContextBySpec={() =>
                        of({
                            ...mockContext,
                            description: `## Search and review commits

Sourcegraph allows you to search code and code history in a single interface.  To find all TODOs in commit messages, add \`type:commit\` to your query.

\`\`\`sourcegraph
repo:^github\\.com/sourcegraph/sourcegraph$ type:commit TODO
\`\`\`

## Search code in commits

You can also search the code in commits by setting the type to \`diff\`.

\`\`\`sourcegraph
repo:^github\\.com/sourcegraph/sourcegraph$ type:diff TODO
\`\`\`

## Find code removed in a timeframe

By combining \`before:\` and \`after:\` filters, you can search for a range of time the code may have existed.

\`\`\`sourcegraph
repo:^github\\.com/sourcegraph/sourcegraph$ type:diff before:yesterday after:"last week" TODO
\`\`\`
                            `,
                        })
                    }
                />
            )}
        </WebStory>
    ),
    {}
)

add(
    'search box with selected context',
    () => (
        <WebStory>
            {webProps => (
                <SearchContextPage
                    {...webProps}
                    {...props}
                    showSearchContext={true}
                    selectedSearchContextSpec={mockContext.spec}
                    fetchSearchContextBySpec={fetchPublicContext}
                />
            )}
        </WebStory>
    ),
    {}
)

add(
    'autodefined context',
    () => (
        <WebStory>
            {webProps => (
                <SearchContextPage {...webProps} {...props} fetchSearchContextBySpec={fetchAutoDefinedContext} />
            )}
        </WebStory>
    ),
    {}
)

add(
    'private context',
    () => (
        <WebStory>
            {webProps => <SearchContextPage {...webProps} {...props} fetchSearchContextBySpec={fetchPrivateContext} />}
        </WebStory>
    ),
    {}
)

add(
    'loading',
    () => (
        <WebStory>
            {webProps => <SearchContextPage {...webProps} {...props} fetchSearchContextBySpec={() => NEVER} />}
        </WebStory>
    ),
    {}
)

add(
    'error',
    () => (
        <WebStory>
            {webProps => (
                <SearchContextPage
                    {...webProps}
                    {...props}
                    fetchSearchContextBySpec={() => throwError(new Error('Failed to fetch search context'))}
                />
            )}
        </WebStory>
    ),
    {}
)

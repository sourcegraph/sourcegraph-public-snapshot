import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { subDays } from 'date-fns'
import { NEVER, type Observable, of } from 'rxjs'
import sinon from 'sinon'

import type { SearchContextFields } from '@sourcegraph/shared/src/graphql-operations'
import { NOOP_PLATFORM_CONTEXT } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import type { AuthenticatedUser } from '../../auth'
import { WebStory } from '../../components/WebStory'
import type { OrgAreaOrganizationFields, RepositoryFields } from '../../graphql-operations'

import { SearchContextForm } from './SearchContextForm'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/enterprise/searchContexts/SearchContextForm',
    decorators: [decorator],
    parameters: {
        chromatic: { viewports: [1200], disableSnapshot: false },
    },
}

export default config

const onSubmit = (): Observable<SearchContextFields> =>
    of({
        __typename: 'SearchContext',
        id: '1',
        spec: 'public-ctx',
        name: 'public-ctx',
        namespace: null,
        public: true,
        autoDefined: false,
        description: 'Repositories on Sourcegraph',
        repositories: [],
        query: '',
        updatedAt: subDays(new Date(), 1).toISOString(),
        viewerCanManage: true,
        viewerHasAsDefault: false,
        viewerHasStarred: false,
    })

const searchContextToEdit: SearchContextFields = {
    __typename: 'SearchContext',
    id: '1',
    spec: 'public-ctx',
    name: 'public-ctx',
    namespace: null,
    public: true,
    autoDefined: false,
    description: 'Repositories on Sourcegraph',
    query: '',
    repositories: [
        {
            __typename: 'SearchContextRepositoryRevisions',
            revisions: ['HEAD'],
            repository: { name: 'github.com/example/example' } as RepositoryFields,
        },
    ],
    updatedAt: subDays(new Date(), 1).toISOString(),
    viewerCanManage: true,
    viewerHasAsDefault: false,
    viewerHasStarred: false,
}

const authUser: AuthenticatedUser = {
    __typename: 'User',
    id: '0',
    username: 'alice',
    avatarURL: null,
    session: { canSignOut: true },
    displayName: null,
    url: '',
    settingsURL: '#',
    siteAdmin: true,
    organizations: {
        nodes: [
            { id: '0', settingsURL: '#', name: 'ACME', displayName: 'Acme Corp' },
            { id: '1', settingsURL: '#', name: 'BETA', displayName: 'Beta Inc' },
        ] as OrgAreaOrganizationFields[],
    },
    viewerCanAdminister: true,
    hasVerifiedEmail: true,
    completedPostSignup: true,
    databaseID: 0,
    tosAccepted: true,
    emails: [{ email: 'alice@sourcegraph.com', isPrimary: true, verified: true }],
    latestSettings: null,
    permissions: { nodes: [] },
}

const deleteSearchContext = sinon.fake(() => NEVER)

export const EmptyCreate: StoryFn = () => (
    <WebStory>
        {webProps => (
            <SearchContextForm
                {...webProps}
                authenticatedUser={authUser}
                onSubmit={onSubmit}
                deleteSearchContext={deleteSearchContext}
                isSourcegraphDotCom={false}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )}
    </WebStory>
)

EmptyCreate.storyName = 'empty create'

export const EditExisting: StoryFn = () => (
    <WebStory>
        {webProps => (
            <SearchContextForm
                {...webProps}
                searchContext={searchContextToEdit}
                authenticatedUser={authUser}
                onSubmit={onSubmit}
                deleteSearchContext={deleteSearchContext}
                isSourcegraphDotCom={false}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )}
    </WebStory>
)

EditExisting.storyName = 'edit existing'

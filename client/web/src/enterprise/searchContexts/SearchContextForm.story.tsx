import { DecoratorFn, Meta, Story } from '@storybook/react'
import { subDays } from 'date-fns'
import { NEVER, Observable, of } from 'rxjs'
import sinon from 'sinon'

import { SearchContextFields } from '@sourcegraph/search'
import { NOOP_PLATFORM_CONTEXT } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { AuthenticatedUser } from '../../auth'
import { WebStory } from '../../components/WebStory'
import { OrgAreaOrganizationFields, RepositoryFields } from '../../graphql-operations'

import { SearchContextForm } from './SearchContextForm'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

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
            { id: '0', settingsURL: '#', name: 'ACME', displayName: 'Acme Corp' },
            { id: '1', settingsURL: '#', name: 'BETA', displayName: 'Beta Inc' },
        ] as OrgAreaOrganizationFields[],
    },
    tags: [],
    viewerCanAdminister: true,
    databaseID: 0,
    tosAccepted: true,
    searchable: true,
    emails: [],
}

const deleteSearchContext = sinon.fake(() => NEVER)

export const EmptyCreate: Story = () => (
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

export const EditExisting: Story = () => (
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

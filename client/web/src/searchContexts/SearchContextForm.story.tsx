import { IOrg, ISearchContext } from '@sourcegraph/shared/src/graphql/schema'
import { storiesOf } from '@storybook/react'
import { subDays } from 'date-fns'
import React from 'react'
import { Observable, of } from 'rxjs'
import { AuthenticatedUser } from '../auth'

import { WebStory } from '../components/WebStory'

import { SearchContextForm } from './SearchContextForm'

const { add } = storiesOf('web/searchContexts/SearchContextForm', module)
    .addParameters({
        chromatic: { viewports: [1200] },
    })
    .addDecorator(story => <div className="p-3 container web-content">{story()}</div>)

const onSubmit = (): Observable<ISearchContext> =>
    of({
        __typename: 'SearchContext',
        id: '1',
        spec: 'public-ctx',
        public: true,
        autoDefined: false,
        description: 'Repositories on Sourcegraph',
        repositories: [],
        updatedAt: subDays(new Date(), 1).toISOString(),
    })

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
        ] as IOrg[],
    },
    tags: [],
    viewerCanAdminister: true,
    databaseID: 0,
}

add(
    'empty create',
    () => (
        <WebStory>
            {webProps => <SearchContextForm {...webProps} authenticatedUser={authUser} onSubmit={onSubmit} />}
        </WebStory>
    ),
    {}
)

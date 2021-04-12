import * as H from 'history'
import React from 'react'
import { createRenderer } from 'react-test-renderer/shallow'
import { of } from 'rxjs'

import { IRepository, IGitRef } from '@sourcegraph/shared/src/graphql/schema'

import { RepositoryReleasesTagsPage } from './RepositoryReleasesTagsPage'

describe('RepositoryReleasesTagsPage', () => {
    const history = H.createMemoryHistory()
    test('renders', () =>
        expect(
            createRenderer().render(
                <RepositoryReleasesTagsPage
                    history={history}
                    location={history.location}
                    repo={{ id: '123' } as IRepository}
                    queryGitReferences={() =>
                        of({
                            totalCount: 0,
                            nodes: [] as IGitRef[],
                            __typename: 'GitRefConnection',
                            pageInfo: { __typename: 'PageInfo', endCursor: '', hasNextPage: false },
                        })
                    }
                />
            )
        ).toMatchSnapshot())
})

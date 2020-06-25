import React from 'react'
import * as H from 'history'
import { RepositoryReleasesTagsPage } from './RepositoryReleasesTagsPage'
import { of } from 'rxjs'
import { shallow } from 'enzyme'

describe('RepositoryReleasesTagsPage', () => {
    const history = H.createMemoryHistory()
    test('renders', () =>
        expect(
            shallow(
                <RepositoryReleasesTagsPage
                    history={history}
                    location={history.location}
                    repo={{ id: '123' }}
                    queryGitReferences={() =>
                        of({
                            totalCount: 0,
                            nodes: [],
                            __typename: 'GitRefConnection',
                            pageInfo: { __typename: 'PageInfo', endCursor: '', hasNextPage: false },
                        })
                    }
                />
            )
        ).toMatchSnapshot())
})

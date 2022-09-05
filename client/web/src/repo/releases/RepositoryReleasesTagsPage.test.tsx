import { render } from '@testing-library/react'
import * as H from 'history'
import { of } from 'rxjs'

import { IRepository, IGitRef } from '@sourcegraph/shared/src/schema'

import { RepositoryReleasesTagsPage } from './RepositoryReleasesTagsPage'

describe('RepositoryReleasesTagsPage', () => {
    const history = H.createMemoryHistory()
    test('renders', () =>
        expect(
            render(
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
            ).asFragment()
        ).toMatchSnapshot())
})

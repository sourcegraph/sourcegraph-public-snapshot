import { render } from '@testing-library/react'
import * as H from 'history'
import { of } from 'rxjs'

import { RepositoryFields } from '../../graphql-operations'

import { RepositoryReleasesTagsPage } from './RepositoryReleasesTagsPage'

describe('RepositoryReleasesTagsPage', () => {
    const history = H.createMemoryHistory()
    test('renders', () =>
        expect(
            render(
                <RepositoryReleasesTagsPage
                    history={history}
                    location={history.location}
                    repo={{ id: '123' } as RepositoryFields}
                    queryGitReferences={() =>
                        of({
                            totalCount: 0,
                            nodes: [],
                            __typename: 'GitRefConnection',
                            pageInfo: { __typename: 'PageInfo', endCursor: '', hasNextPage: false },
                        })
                    }
                />
            ).asFragment()
        ).toMatchSnapshot())
})

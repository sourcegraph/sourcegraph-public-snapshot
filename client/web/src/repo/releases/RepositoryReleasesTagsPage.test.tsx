import { render } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom-v5-compat'
import { of } from 'rxjs'

import { RepositoryFields } from '../../graphql-operations'

import { RepositoryReleasesTagsPage } from './RepositoryReleasesTagsPage'

describe('RepositoryReleasesTagsPage', () => {
    test('renders', () =>
        expect(
            render(
                <MemoryRouter>
                    <RepositoryReleasesTagsPage
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
                </MemoryRouter>
            ).asFragment()
        ).toMatchSnapshot())
})

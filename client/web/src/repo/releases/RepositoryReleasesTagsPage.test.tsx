import { describe, expect, test } from '@jest/globals'
import { render } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { of } from 'rxjs'

import type { RepositoryFields } from '../../graphql-operations'

import { RepositoryReleasesTagsPage } from './RepositoryReleasesTagsPage'

describe('RepositoryReleasesTagsPage', () => {
    test('renders', () =>
        expect(
            render(
                <MemoryRouter>
                    <RepositoryReleasesTagsPage
                        repo={{ id: '123' } as RepositoryFields}
                        isPackage={false}
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

    test('renders for packages', () =>
        expect(
            render(
                <MemoryRouter>
                    <RepositoryReleasesTagsPage
                        repo={{ id: '123' } as RepositoryFields}
                        isPackage={true}
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

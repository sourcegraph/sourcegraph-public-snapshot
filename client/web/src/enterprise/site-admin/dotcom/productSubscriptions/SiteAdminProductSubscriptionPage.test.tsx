import { act } from '@testing-library/react'
import { of } from 'rxjs'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { DotComProductSubscriptionResult, ProductLicensesResult } from '../../../../graphql-operations'

import { SiteAdminProductSubscriptionPage } from './SiteAdminProductSubscriptionPage'
import { mockLicenseContext } from './testUtils'

jest.mock('mdi-react/ArrowLeftIcon', () => 'ArrowLeftIcon')

jest.mock('mdi-react/AddIcon', () => 'AddIcon')

describe('SiteAdminProductSubscriptionPage', () => {
    const origContext = window.context
    beforeEach(() => {
        window.context = mockLicenseContext
    })
    afterEach(() => {
        window.context = origContext
    })
    test('renders', () => {
        const component = renderWithBrandedContext(
            <MockedTestProvider mocks={[]}>
                <SiteAdminProductSubscriptionPage
                    _queryProductSubscription={() =>
                        of<DotComProductSubscriptionResult['dotcom']['productSubscription']>({
                            __typename: 'ProductSubscription',
                            createdAt: '2020-01-01',
                            url: '/s',
                            account: null,
                            id: 'l1',
                            isArchived: false,
                            name: 'sn1',
                            productLicenses: {
                                __typename: 'ProductLicenseConnection',
                                nodes: [
                                    {
                                        createdAt: '2020-01-01',
                                        id: 'l1',
                                        licenseKey: 'lk1',
                                        info: {
                                            __typename: 'ProductLicenseInfo',
                                            expiresAt: '2021-01-01',
                                            tags: ['a'],
                                            userCount: 123,
                                        },
                                    },
                                ],
                                totalCount: 1,
                                pageInfo: { hasNextPage: false },
                            },
                            activeLicense: null,
                            llmProxyAccess: {
                                __typename: 'LLMProxyAccess',
                                enabled: false,
                                rateLimit: null,
                            },
                        })
                    }
                    _queryProductLicenses={() =>
                        of<ProductLicensesResult['dotcom']['productSubscription']['productLicenses']>({
                            __typename: 'ProductLicenseConnection',
                            nodes: [
                                {
                                    createdAt: '2020-01-01',
                                    id: 'l1',
                                    licenseKey: 'lk1',
                                    info: {
                                        __typename: 'ProductLicenseInfo',
                                        expiresAt: '2021-01-01',
                                        productNameWithBrand: 'NB',
                                        tags: ['a'],
                                        userCount: 123,
                                    },
                                    subscription: {
                                        id: 'l1',
                                        name: 'sn1',
                                        urlForSiteAdmin: null,
                                        account: null,
                                        activeLicense: { id: 'l1' },
                                    },
                                },
                            ],
                            totalCount: 1,
                            pageInfo: { hasNextPage: false },
                        })
                    }
                />
            </MockedTestProvider>,
            { route: '/p' }
        )
        act(() => undefined)
        expect(component.asFragment()).toMatchSnapshot()
    })
})

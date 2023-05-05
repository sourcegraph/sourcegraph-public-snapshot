import { MockedResponse } from '@apollo/client/testing'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import {
    DotComProductSubscriptionResult,
    DotComProductSubscriptionVariables,
    ProductLicensesResult,
    ProductLicensesVariables,
} from '../../../../graphql-operations'

import { PRODUCT_LICENSES, DOTCOM_PRODUCT_SUBSCRIPTION } from './backend'
import { SiteAdminProductSubscriptionPage } from './SiteAdminProductSubscriptionPage'
import { mockLicenseContext } from './testUtils'

jest.mock('mdi-react/ArrowLeftIcon', () => 'ArrowLeftIcon')

jest.mock('mdi-react/AddIcon', () => 'AddIcon')

const subscriptionMock: MockedResponse<DotComProductSubscriptionResult, DotComProductSubscriptionVariables> = {
    request: {
        query: getDocumentNode(DOTCOM_PRODUCT_SUBSCRIPTION),
        variables: { uuid: '' },
    },
    result: {
        data: {
            __typename: 'Query',
            dotcom: {
                __typename: 'DotcomQuery',
                productSubscription: {
                    __typename: 'ProductSubscription',
                    createdAt: '2023-05-05T13:10:30.080Z',
                    url: '/s',
                    account: null,
                    id: 'l1',
                    isArchived: false,
                    name: 'sn1',
                    productLicenses: {
                        __typename: 'ProductLicenseConnection',
                        nodes: [
                            {
                                __typename: 'ProductLicense',
                                createdAt: '2023-05-05T13:10:30.080Z',
                                id: 'l1',
                                licenseKey: 'lk1',
                                info: {
                                    __typename: 'ProductLicenseInfo',
                                    expiresAt: '2024-05-05T13:10:30.080Z',
                                    tags: ['a'],
                                    userCount: 123,
                                },
                            },
                        ],
                        totalCount: 1,
                        pageInfo: { __typename: 'PageInfo', hasNextPage: false },
                    },
                    activeLicense: null,
                    sourcegraphAccessToken: '123',
                    llmProxyAccess: {
                        __typename: 'LLMProxyAccess',
                        enabled: false,
                        rateLimit: null,
                    },
                },
            },
        },
    },
}

const licensesMock: MockedResponse<ProductLicensesResult> = {
    request: {
        query: getDocumentNode(PRODUCT_LICENSES),
        variables: {
            first: 20,
            subscriptionUUID: '',
            after: null,
        },
    },
    result: {
        data: {
            __typename: 'Query',
            dotcom: {
                __typename: 'DotcomQuery',
                productSubscription: {
                    __typename: 'ProductSubscription',
                    productLicenses: {
                        __typename: 'ProductLicenseConnection',
                        nodes: [
                            {
                                __typename: 'ProductLicense',
                                createdAt: '2023-05-05T13:10:30.080Z',
                                id: 'l1',
                                licenseKey: 'lk1',
                                info: {
                                    __typename: 'ProductLicenseInfo',
                                    expiresAt: '2024-05-05T13:10:30.080Z',
                                    productNameWithBrand: 'NB',
                                    tags: ['a'],
                                    userCount: 123,
                                },
                                subscription: {
                                    __typename: 'ProductSubscription',
                                    id: 'l1',
                                    name: 'sn1',
                                    urlForSiteAdmin: null,
                                    account: null,
                                    activeLicense: { __typename: 'ProductLicense', id: 'l1' },
                                },
                            },
                        ],
                        totalCount: 1,
                        pageInfo: { __typename: 'PageInfo', hasNextPage: false },
                    },
                },
            },
        },
    },
}

describe('SiteAdminProductSubscriptionPage', () => {
    const origContext = window.context
    beforeEach(() => {
        window.context = mockLicenseContext
    })
    afterEach(() => {
        window.context = origContext
    })
    test('renders', async () => {
        const component = renderWithBrandedContext(
            <MockedTestProvider mocks={[subscriptionMock, licensesMock, subscriptionMock, licensesMock]}>
                <SiteAdminProductSubscriptionPage />
            </MockedTestProvider>,
            { route: '/p' }
        )
        await waitForNextApolloResponse()
        await waitForNextApolloResponse()
        expect(component.asFragment()).toMatchSnapshot()
    })
})

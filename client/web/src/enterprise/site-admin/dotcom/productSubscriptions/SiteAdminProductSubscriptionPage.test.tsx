import { act } from '@testing-library/react'
import * as H from 'history'
import { of } from 'rxjs'

import * as GQL from '@sourcegraph/shared/src/schema'
import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'

import { SiteAdminProductSubscriptionPage } from './SiteAdminProductSubscriptionPage'

jest.mock('mdi-react/ArrowLeftIcon', () => 'ArrowLeftIcon')

jest.mock('mdi-react/AddIcon', () => 'AddIcon')

const history = H.createMemoryHistory()
const location = H.createLocation('/')

describe('SiteAdminProductSubscriptionPage', () => {
    test('renders', () => {
        const component = renderWithBrandedContext(
            <SiteAdminProductSubscriptionPage
                match={{ isExact: true, params: { subscriptionUUID: 's' }, path: '/p', url: '/p' }}
                history={history}
                location={location}
                _queryProductSubscription={() =>
                    of<GQL.IProductSubscription>({
                        __typename: 'ProductSubscription',
                        events: [] as GQL.IProductSubscription['events'],
                        createdAt: '2020-01-01',
                        url: '/s',
                        urlForSiteAdminBilling: null,
                    } as GQL.IProductSubscription)
                }
                _queryProductLicenses={() =>
                    of<GQL.IProductLicenseConnection>({
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
                                subscription: { activeLicense: { id: 'l1' } } as GQL.IProductSubscription,
                            },
                        ] as GQL.IProductLicense[],
                        totalCount: 1,
                        pageInfo: { hasNextPage: false } as GQL.IPageInfo,
                    })
                }
            />,
            { history }
        )
        act(() => undefined)
        expect(component.asFragment()).toMatchSnapshot()
    })
})

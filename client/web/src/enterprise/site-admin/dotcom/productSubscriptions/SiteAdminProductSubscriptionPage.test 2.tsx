import * as H from 'history'
import React from 'react'
import { MemoryRouter } from 'react-router'
import renderer, { act } from 'react-test-renderer'
import { of } from 'rxjs'

import * as GQL from '@sourcegraph/shared/src/graphql/schema'

import { SiteAdminProductSubscriptionPage } from './SiteAdminProductSubscriptionPage'

jest.mock('mdi-react/ArrowLeftIcon', () => 'ArrowLeftIcon')

jest.mock('mdi-react/AddIcon', () => 'AddIcon')

jest.mock('./SiteAdminProductLicenseNode', () => ({ SiteAdminProductLicenseNode: 'SiteAdminProductLicenseNode' }))

jest.mock('../../../dotcom/productSubscriptions/AccountName', () => ({ AccountName: 'AccountName' }))

const history = H.createMemoryHistory()
const location = H.createLocation('/')

describe('SiteAdminProductSubscriptionPage', () => {
    test('renders', () => {
        const component = renderer.create(
            <MemoryRouter>
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
                />
            </MemoryRouter>
        )
        act(() => undefined)
        expect(component.toJSON()).toMatchSnapshot()
    })
})

import { render } from '@testing-library/react'
import React from 'react'
import { MemoryRouter } from 'react-router'

import * as GQL from '@sourcegraph/shared/src/schema'

import { SiteAdminProductLicenseNode } from './SiteAdminProductLicenseNode'

jest.mock('../../../dotcom/productSubscriptions/AccountName', () => ({ AccountName: 'AccountName' }))

describe('SiteAdminProductLicenseNode', () => {
    test('active', () => {
        expect(
            render(
                <MemoryRouter>
                    <SiteAdminProductLicenseNode
                        node={
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
                                    name: 's',
                                    activeLicense: { id: 'l1' },
                                    urlForSiteAdmin: '/s',
                                } as GQL.IProductSubscription,
                            } as GQL.IProductLicense
                        }
                        showSubscription={true}
                    />
                </MemoryRouter>
            ).asFragment()
        ).toMatchSnapshot()
    })

    test('inactive', () => {
        expect(
            render(
                <MemoryRouter>
                    <SiteAdminProductLicenseNode
                        node={
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
                                    name: 's',
                                    activeLicense: { id: 'l0' },
                                    urlForSiteAdmin: '/s',
                                } as GQL.IProductSubscription,
                            } as GQL.IProductLicense
                        }
                        showSubscription={true}
                    />
                </MemoryRouter>
            ).asFragment()
        ).toMatchSnapshot()
    })
})

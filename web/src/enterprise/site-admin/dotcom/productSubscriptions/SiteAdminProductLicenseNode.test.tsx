import React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { SiteAdminProductLicenseNode } from './SiteAdminProductLicenseNode'
import { MemoryRouter } from 'react-router'
import { mount } from 'enzyme'

jest.mock('../../../dotcom/productSubscriptions/AccountName', () => ({ AccountName: 'AccountName' }))

describe('SiteAdminProductLicenseNode', () => {
    test('active', () => {
        expect(
            mount(
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
            ).children()
        ).toMatchSnapshot()
    })

    test('inactive', () => {
        expect(
            mount(
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
            ).children()
        ).toMatchSnapshot()
    })
})

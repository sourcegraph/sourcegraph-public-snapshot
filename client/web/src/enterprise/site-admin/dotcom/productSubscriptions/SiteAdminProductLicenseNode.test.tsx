import * as GQL from '@sourcegraph/shared/src/schema'
import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'

import { SiteAdminProductLicenseNode } from './SiteAdminProductLicenseNode'

jest.mock('../../../dotcom/productSubscriptions/AccountName', () => ({ AccountName: 'AccountName' }))

describe('SiteAdminProductLicenseNode', () => {
    test('active', () => {
        expect(
            renderWithBrandedContext(
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
            ).asFragment()
        ).toMatchSnapshot()
    })

    test('inactive', () => {
        expect(
            renderWithBrandedContext(
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
            ).asFragment()
        ).toMatchSnapshot()
    })
})

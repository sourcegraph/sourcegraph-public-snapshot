import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { SiteAdminProductLicenseNode } from './SiteAdminProductLicenseNode'
import { mockLicenseContext } from './testUtils'

jest.mock('../../../dotcom/productSubscriptions/AccountName', () => ({ AccountName: 'AccountName' }))

describe('SiteAdminProductLicenseNode', () => {
    const origContext = window.context
    beforeEach(() => {
        window.context = mockLicenseContext
    })
    afterEach(() => {
        window.context = origContext
    })
    test('active', () => {
        expect(
            renderWithBrandedContext(
                <MockedTestProvider mocks={[]}>
                    <SiteAdminProductLicenseNode
                        node={{
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
                                id: 'id1',
                                account: null,
                                name: 's',
                                activeLicense: { id: 'l1' },
                                urlForSiteAdmin: '/s',
                            },
                        }}
                        showSubscription={true}
                    />
                </MockedTestProvider>
            ).asFragment()
        ).toMatchSnapshot()
    })

    test('inactive', () => {
        expect(
            renderWithBrandedContext(
                <SiteAdminProductLicenseNode
                    node={{
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
                            id: 'id1',
                            account: null,
                            name: 's',
                            activeLicense: { id: 'l0' },
                            urlForSiteAdmin: '/s',
                        },
                    }}
                    showSubscription={true}
                />
            ).asFragment()
        ).toMatchSnapshot()
    })
})

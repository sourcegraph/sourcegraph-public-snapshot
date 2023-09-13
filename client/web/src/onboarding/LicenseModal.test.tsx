import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { LicenseKeyModal } from './LicenseModal'

describe('LicenseKeyModal', () => {
    it('should render', () => {
        expect(
            renderWithBrandedContext(
                <LicenseKeyModal
                    licenseKey={{ key: 'test', tags: ['test'], userCount: 10, expiresAt: '2030-01-01' }}
                    onHandleLicenseCheck={() => {}}
                    config="{}"
                    id={1}
                />
            ).asFragment()
        ).toMatchSnapshot()
    })
})

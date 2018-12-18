import { noop } from 'lodash'
import * as React from 'react'
import { cleanup, fireEvent, render } from 'react-testing-library'

import { WithRouterProvider } from '../../../shared/src/testing/router'
import { SiteAdminAddExternalServicePage } from './SiteAdminAddExternalServicePage'

describe('SiteAdminAddExternalServicePage', () => {
    afterEach(cleanup)

    it('updates the kind selector after change', () => {
        const { container } = render(
            <WithRouterProvider>
                {({ history, location }) => (
                    <SiteAdminAddExternalServicePage
                        history={history}
                        location={location}
                        isLightTheme={true}
                        eventLogger={{
                            log: noop,
                            logViewEvent: noop,
                        }}
                    />
                )}
            </WithRouterProvider>
        )

        const kindInput = container.querySelector<HTMLSelectElement>('#external-service-page-form-kind')!
        fireEvent.change(kindInput, { target: { value: 'GITOLITE' } })

        expect(kindInput.value).toBe('GITOLITE')
    })
})

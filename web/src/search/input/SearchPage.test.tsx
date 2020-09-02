import React from 'react'
import { cleanup, render } from '@testing-library/react'
import { Controller } from '../../../../shared/src/extensions/controller'
import { createMemoryHistory } from 'history'
import { NOOP_TELEMETRY_SERVICE } from '../../../../shared/src/telemetry/telemetryService'
import { SearchPage, SearchPageProps } from './SearchPage'
import { Services } from '../../../../shared/src/api/client/services'

jest.mock('./SearchPageInput', () => ({
    SearchPageInput: () => null,
}))

describe('SearchPage', () => {
    afterAll(cleanup)

    let container: HTMLElement

    const history = createMemoryHistory()
    const defaultProps = {
        isSourcegraphDotCom: false,
        settingsCascade: {
            final: null,
            subjects: null,
        },
        location: history.location,
        extensionsController: {
            services: {} as Services,
        } as Pick<Controller, 'executeCommand' | 'services'>,
        telemetryService: NOOP_TELEMETRY_SERVICE,
    } as SearchPageProps

    it('should not show enterprise home panels if on Sourcegraph.com', () => {
        container = render(<SearchPage {...defaultProps} isSourcegraphDotCom={true} />).container
        const enterpriseHomePanels = container.querySelector('.enterprise-home-panels')
        expect(enterpriseHomePanels).toBeFalsy()
    })

    it('should not show enterprise home panels if showEnterpriseHomePanels disabled', () => {
        container = render(<SearchPage {...defaultProps} />).container
        const enterpriseHomePanels = container.querySelector('.enterprise-home-panels')
        expect(enterpriseHomePanels).toBeFalsy()
    })

    it('should show enterprise home panels if showEnterpriseHomePanels enabled and not on Sourcegraph.com', () => {
        container = render(<SearchPage {...defaultProps} showEnterpriseHomePanels={true} />).container
        const enterpriseHomePanels = container.querySelector('.enterprise-home-panels')
        expect(enterpriseHomePanels).toBeTruthy()
    })
})

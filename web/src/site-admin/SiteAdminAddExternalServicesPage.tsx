import * as H from 'history'
import React from 'react'
import { PageTitle } from '../components/PageTitle'
import { ThemeProps } from '../../../shared/src/theme'
import { ExternalServiceCard } from '../components/ExternalServiceCard'
import { codeHostExternalServices, nonCodeHostExternalServices, allExternalServices } from './externalServices'
import { SiteAdminAddExternalServicePage } from './SiteAdminAddExternalServicePage'

interface Props extends ThemeProps {
    history: H.History
    eventLogger: {
        logViewEvent: (event: 'AddExternalService') => void
        log: (event: 'AddExternalServiceFailed' | 'AddExternalServiceSucceeded', eventProperties?: any) => void
    }
}

/**
 * Page for choosing a service kind and variant to add, among the available options.
 */
export class SiteAdminAddExternalServicesPage extends React.Component<Props> {
    /**
     * Returns the id of the external service from the URL parameters.
     */
    private getExternalServiceID(): string | null {
        return new URLSearchParams(this.props.history.location.search).get('id') ?? null
    }

    public render(): JSX.Element | null {
        const id = this.getExternalServiceID()
        if (id) {
            const externalService = allExternalServices[id]
            if (externalService) {
                return <SiteAdminAddExternalServicePage {...this.props} externalService={externalService} />
            }
        }
        return (
            <div className="add-external-services-page mt-3">
                <PageTitle title="Add repositories" />
                <div className="d-flex justify-content-between align-items-center mt-3 mb-3">
                    <h2 className="mb-0">Add repositories</h2>
                </div>
                <p className="mt-2">Add repositories from one of these code hosts.</p>
                {Object.entries(codeHostExternalServices).map(([id, externalService]) => (
                    <div className="add-external-services-page__card" key={id}>
                        <ExternalServiceCard to={getAddURL(id)} {...externalService} />
                    </div>
                ))}
                <br />
                <h2>Other connections</h2>
                <p className="mt-2">Add connections to non-code-host services.</p>
                {Object.entries(nonCodeHostExternalServices).map(([id, externalService]) => (
                    <div className="add-external-services-page__card" key={id}>
                        <ExternalServiceCard to={getAddURL(id)} {...externalService} />
                    </div>
                ))}
            </div>
        )
    }
}

function getAddURL(id: string): string {
    const parameters = new URLSearchParams()
    parameters.append('id', id)
    return `?${parameters.toString()}`
}

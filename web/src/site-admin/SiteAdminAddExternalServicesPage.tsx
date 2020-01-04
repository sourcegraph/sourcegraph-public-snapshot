import * as H from 'history'
import React from 'react'
import { PageTitle } from '../components/PageTitle'
import { ThemeProps } from '../../../shared/src/theme'
import { ExternalServiceCard } from '../components/ExternalServiceCard'
import { onboardingExternalServices, nonCodeHostExternalServices } from './externalServices'
import { SiteAdminAddExternalServicePage } from './SiteAdminAddExternalServicePage'
import { map } from 'lodash'

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
     * Gets the external service kind and add-service kind from the URL paramsters
     */
    private getExternalServiceID(): string | null {
        return new URLSearchParams(this.props.history.location.search).get('id') ?? null
    }

    public render(): JSX.Element | null {
        const id = this.getExternalServiceID()
        if (id) {
            const externalService = onboardingExternalServices[id]
            if (externalService) {
                return <SiteAdminAddExternalServicePage {...this.props} externalService={externalService} />
            }
        }
        return (
            <div className="add-external-services-page mt-3">
                <PageTitle title="Add repositories" />
                <h1>Add repositories</h1>
                <p>Where would you like to add repositories from?</p>
                {map(onboardingExternalServices, (externalService, id) => (
                    <div className="add-external-services-page__card" key={id}>
                        <ExternalServiceCard to={getAddURL(id)} {...externalService} />
                    </div>
                ))}
                <br />
                <h2>Other connections</h2>
                {map(nonCodeHostExternalServices, (externalService, id) => (
                    <div className="add-external-services-page__card" key={id}>
                        <ExternalServiceCard to={getAddURL(id)} {...externalService} />
                    </div>
                ))}
            </div>
        )
    }
}

function getAddURL(id: string): string {
    const params = new URLSearchParams()
    params.append('id', id)
    return `?${params.toString()}`
}

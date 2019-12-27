import * as H from 'history'
import React from 'react'
import { PageTitle } from '../components/PageTitle'
import { ThemeProps } from '../../../shared/src/theme'
import { ExternalServiceCard } from '../components/ExternalServiceCard'
import { externalServices } from './externalServices'
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
            const externalService = externalServices[id]
            if (externalService) {
                return <SiteAdminAddExternalServicePage {...this.props} externalService={externalService} />
            }
        }
        return (
            <div className="add-external-services-page mt-3">
                <PageTitle title="Choose an external service type to add" />
                <h1>Add external service</h1>
                <p>Choose an external service to add to Sourcegraph.</p>
                {map(externalServices, (externalService, id) => (
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

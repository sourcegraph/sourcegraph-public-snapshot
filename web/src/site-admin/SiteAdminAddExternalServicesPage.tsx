import * as H from 'history'
import React from 'react'
import * as GQL from '../../../shared/src/graphql/schema'
import { PageTitle } from '../components/PageTitle'
import { ThemeProps } from '../../../shared/src/theme'
import { ExternalServiceCard } from '../components/ExternalServiceCard'
import {
    AddExternalServiceMetadata,
    ALL_EXTERNAL_SERVICE_ADD_VARIANTS,
    ExternalServiceVariant,
    getExternalService,
    isExternalServiceVariant,
} from './externalServices'
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
     * Gets the external service kind and add-service kind from the URL paramsters
     */
    private getExternalServiceKind(): {
        kind: GQL.ExternalServiceKind | null
        variant: ExternalServiceVariant | undefined
    } {
        const params = new URLSearchParams(this.props.history.location.search)
        let kind = params.get('kind') || undefined
        if (kind) {
            kind = kind.toUpperCase()
        }
        const isKnownKind = (kind: string): kind is GQL.ExternalServiceKind =>
            !!getExternalService(kind as GQL.ExternalServiceKind)

        const q = params.get('variant')
        const variant = q && isExternalServiceVariant(q) ? q : undefined
        return { kind: kind && isKnownKind(kind) ? kind : null, variant }
    }

    private static getAddURL(serviceToAdd: AddExternalServiceMetadata): string {
        const params = new URLSearchParams()
        params.append('kind', serviceToAdd.kind.toLowerCase())
        if (serviceToAdd.variant) {
            params.append('variant', serviceToAdd.variant)
        }
        return `?${params.toString()}`
    }

    public render(): JSX.Element | null {
        const { kind, variant } = this.getExternalServiceKind()
        if (kind) {
            return <SiteAdminAddExternalServicePage {...this.props} kind={kind} variant={variant} />
        }
        return (
            <div className="add-external-services-page mt-3">
                <PageTitle title="Choose an external service type to add" />
                <h1>Add external service</h1>
                <p>Choose an external service to add to Sourcegraph.</p>
                {ALL_EXTERNAL_SERVICE_ADD_VARIANTS.map((service, i) => (
                    <div className="add-external-services-page__card" key={i}>
                        <ExternalServiceCard to={SiteAdminAddExternalServicesPage.getAddURL(service)} {...service} />
                    </div>
                ))}
            </div>
        )
    }
}

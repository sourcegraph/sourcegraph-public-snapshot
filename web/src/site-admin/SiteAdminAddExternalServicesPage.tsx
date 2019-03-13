import * as H from 'history'
import { map as lodashMap } from 'lodash'
import React from 'react'
import { LinkOrButton } from '../../../shared/src/components/LinkOrButton'
import * as GQL from '../../../shared/src/graphql/schema'
import { PageTitle } from '../components/PageTitle'
import { ThemeProps } from '../theme'
import { ExternalServiceCard } from './ExternalServiceCard'
import {
    AddExternalServiceMetadata,
    ALL_ADD_EXTERNAL_SERVICES,
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
    private getExternalServiceKind(): [GQL.ExternalServiceKind | null, ExternalServiceVariant | null] {
        const params = new URLSearchParams(this.props.history.location.search)
        let kind = params.get('kind') || undefined
        if (kind) {
            kind = kind.toUpperCase()
        }
        const isKnownKind = (kind: string): kind is GQL.ExternalServiceKind =>
            !!getExternalService(kind as GQL.ExternalServiceKind)

        const q = params.get('variant')
        const variant = q && isExternalServiceVariant(q) ? q : null

        return [kind && isKnownKind(kind) ? kind : null, variant]
    }

    private static getAddURL(addService: AddExternalServiceMetadata): string {
        const components: { [key: string]: string } = {
            kind: encodeURIComponent(addService.serviceKind.toLowerCase()),
        }
        if (addService.variant) {
            components.variant = encodeURIComponent(addService.variant)
        }
        return '?' + lodashMap(components, (v, k) => `${encodeURIComponent(k)}=${encodeURIComponent(v)}`).join('&')
    }

    public render(): JSX.Element | null {
        const [kind, variant] = this.getExternalServiceKind()
        if (kind) {
            return <SiteAdminAddExternalServicePage {...this.props} kind={kind} variant={variant || undefined} />
        } else {
            return (
                <div className="add-external-services-page">
                    <PageTitle title="Choose an external service type to add" />
                    <h1>Add external service</h1>
                    <p>Choose an external service to add to Sourcegraph.</p>
                    {ALL_ADD_EXTERNAL_SERVICES.map((addService, i) => (
                        <LinkOrButton
                            className="add-external-services-page__active-card"
                            key={i}
                            to={SiteAdminAddExternalServicesPage.getAddURL(addService)}
                        >
                            <ExternalServiceCard {...addService} />
                        </LinkOrButton>
                    ))}
                </div>
            )
        }
    }
}

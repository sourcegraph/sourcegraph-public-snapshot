import * as H from 'history'
import React from 'react'
import * as GQL from '../../../shared/src/graphql/schema'
import { PageTitle } from '../components/PageTitle'
import { ThemeProps } from '../../../shared/src/theme'
import { CodeHostCard } from '../components/CodeHostCard'
import {
    AddCodeHostMetadata,
    ALL_EXTERNAL_SERVICE_ADD_VARIANTS,
    CodeHostVariant,
    getCodeHost,
    isCodeHostVariant,
} from './externalServices'
import { SiteAdminAddCodeHostPage } from './SiteAdminAddCodeHostPage'

interface Props extends ThemeProps {
    history: H.History
    eventLogger: {
        logViewEvent: (event: 'AddCodeHost') => void
        log: (event: 'AddCodeHostFailed' | 'AddCodeHostSucceeded', eventProperties?: any) => void
    }
}

/**
 * Page for choosing a service kind and variant to add, among the available options.
 */
export class SiteAdminAddCodeHostsPage extends React.Component<Props> {
    /**
     * Gets the external service kind and add-service kind from the URL paramsters
     */
    private getCodeHostKind(): {
        kind: GQL.CodeHostKind | null
        variant: CodeHostVariant | undefined
    } {
        const params = new URLSearchParams(this.props.history.location.search)
        let kind = params.get('kind') || undefined
        if (kind) {
            kind = kind.toUpperCase()
        }
        const isKnownKind = (kind: string): kind is GQL.CodeHostKind =>
            !!getCodeHost(kind as GQL.CodeHostKind)

        const q = params.get('variant')
        const variant = q && isCodeHostVariant(q) ? q : undefined
        return { kind: kind && isKnownKind(kind) ? kind : null, variant }
    }

    private static getAddURL(serviceToAdd: AddCodeHostMetadata): string {
        const params = new URLSearchParams()
        params.append('kind', serviceToAdd.kind.toLowerCase())
        if (serviceToAdd.variant) {
            params.append('variant', serviceToAdd.variant)
        }
        return `?${params.toString()}`
    }

    public render(): JSX.Element | null {
        const { kind, variant } = this.getCodeHostKind()
        if (kind) {
            return <SiteAdminAddCodeHostPage {...this.props} kind={kind} variant={variant} />
        }
        return (
            <div className="add-external-services-page mt-3">
                <PageTitle title="Choose an external service type to add" />
                <h1>Add external service</h1>
                <p>Choose an external service to add to Sourcegraph.</p>
                {ALL_EXTERNAL_SERVICE_ADD_VARIANTS.map((service, i) => (
                    <div className="add-external-services-page__card" key={i}>
                        <CodeHostCard to={SiteAdminAddCodeHostsPage.getAddURL(service)} {...service} />
                    </div>
                ))}
            </div>
        )
    }
}

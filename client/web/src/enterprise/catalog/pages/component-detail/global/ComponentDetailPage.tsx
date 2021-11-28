import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useEffect } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { useQuery } from '@sourcegraph/shared/src/graphql/apollo'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { AuthenticatedUser } from '../../../../../auth'
import { HeroPage } from '../../../../../components/HeroPage'
import { PageTitle } from '../../../../../components/PageTitle'
import { CatalogComponentByIDResult, CatalogComponentByIDVariables } from '../../../../../graphql-operations'
import { CatalogComponentFiltersProps } from '../../../core/component-filters'
import { ComponentList } from '../../overview/components/component-list/ComponentList'
import { Sidebar } from '../sidebar/Sidebar'

import { ComponentDetailContent } from './ComponentDetailContent'
import { CATALOG_COMPONENT_BY_ID } from './gql'

export interface Props extends CatalogComponentFiltersProps, TelemetryProps {
    /** The GraphQL ID of the CatalogComponent. */
    catalogComponentID: Scalars['ID']

    authenticatedUser: AuthenticatedUser
}

/**
 * The catalog component detail page.
 */
export const ComponentDetailPage: React.FunctionComponent<Props> = ({
    catalogComponentID,
    filters,
    onFiltersChange,
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('CatalogComponentDetail')
    }, [telemetryService])

    const { data, error, loading } = useQuery<CatalogComponentByIDResult, CatalogComponentByIDVariables>(
        CATALOG_COMPONENT_BY_ID,
        {
            variables: { id: catalogComponentID },

            // Cache this data but always re-request it in the background when we revisit
            // this page to pick up newer changes.
            fetchPolicy: 'cache-and-network',

            // For subsequent requests while this page is open, make additional network
            // requests; this is necessary for `refetch` to actually use the network. (see
            // https://github.com/apollographql/apollo-client/issues/5515)
            nextFetchPolicy: 'network-only',
        }
    )

    if (loading && !data) {
        return (
            <div className="text-center">
                <LoadingSpinner className="icon-inline mx-auto my-4" />
            </div>
        )
    }
    if (error && !data) {
        throw new Error(error.message)
    }
    if (!data || !data.node || data.node.__typename !== 'CatalogComponent') {
        return <HeroPage icon={AlertCircleIcon} title="Component not found in catalog" />
    }

    const catalogComponent = data.node

    return (
        <>
            <PageTitle title={catalogComponent.name} />
            <Sidebar>
                <ComponentList
                    selected={catalogComponent}
                    filters={filters}
                    onFiltersChange={onFiltersChange}
                    className="flex-1"
                    size="sm"
                />
            </Sidebar>
            <div className="p-2 overflow-auto">
                <ComponentDetailContent catalogComponent={catalogComponent} telemetryService={telemetryService} />
            </div>
        </>
    )
}

import * as H from 'history'
import AddIcon from 'mdi-react/AddIcon'
import React, { useEffect } from 'react'
import { Redirect } from 'react-router'

import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { createAggregateError } from '@sourcegraph/shared/src/util/errors'
import { useConnection } from '@sourcegraph/web/src/components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '@sourcegraph/web/src/components/FilteredConnection/ui'

import { AuthenticatedUser } from '../../auth'
import {
    ListExternalServiceFields,
    Scalars,
    ExternalServicesResult,
    ExternalServicesVariables,
} from '../../graphql-operations'
import { PageTitle } from '../PageTitle'

import { EXTERNAL_SERVICES } from './backend'
import { ExternalServiceNode } from './ExternalServiceNode'

interface Props extends ActivationProps, TelemetryProps {
    history: H.History
    routingPrefix: string
    afterDeleteRoute: string
    userID?: Scalars['ID']
    authenticatedUser: Pick<AuthenticatedUser, 'id'>
}

const BATCH_COUNT = 20

/**
 * A page displaying the external services on this site.
 */
export const ExternalServicesPage: React.FunctionComponent<Props> = ({
    afterDeleteRoute,
    history,
    routingPrefix,
    activation,
    userID,
    telemetryService,
    authenticatedUser,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('SiteAdminExternalServices')
    }, [telemetryService])

    const { connection, loading, error, fetchMore, hasNextPage } = useConnection<
        ExternalServicesResult,
        ExternalServicesVariables,
        ListExternalServiceFields
    >({
        query: EXTERNAL_SERVICES,
        variables: { first: BATCH_COUNT, after: null, namespace: userID ?? null },
        getConnection: ({ data, errors }) => {
            if (!data || !data.externalServices || errors) {
                throw createAggregateError(errors)
            }

            return data.externalServices
        },
        options: {
            useURL: true,
        },
    })

    useEffect(() => {
        if (activation && connection?.totalCount && connection.totalCount > 0) {
            // TODO: Check
            activation.update({ ConnectedCodeHost: true })
        }
        // Activation changes in here, so we cannot recreate the callback on change,
        // or queryConnection will constantly change, resulting in infinite refetch loops.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    const isManagingOtherUser = !!userID && userID !== authenticatedUser.id

    if (!isManagingOtherUser && connection?.totalCount === 0) {
        return <Redirect to={`${routingPrefix}/external-services/new`} />
    }

    const summary = connection && (
        <ConnectionSummary
            connection={connection}
            first={BATCH_COUNT}
            noun="code host"
            pluralNoun="code hosts"
            hasNextPage={hasNextPage}
            noSummaryIfAllNodesVisible={true}
        />
    )

    return (
        <div className="site-admin-external-services-page">
            <PageTitle title="Manage code hosts" />
            <div className="d-flex justify-content-between align-items-center mb-3">
                <h2 className="mb-0">Manage code hosts</h2>
                {!isManagingOtherUser && (
                    <Link
                        className="btn btn-primary test-goto-add-external-service-page"
                        to={`${routingPrefix}/external-services/new`}
                    >
                        <AddIcon className="icon-inline" /> Add code host
                    </Link>
                )}
            </div>
            <p className="mt-2">Manage code host connections to sync repositories.</p>
            <ConnectionContainer className="list-group list-group-flush mt-3">
                {error && <ConnectionError errors={[error.message]} />}
                {connection && (
                    <ConnectionList>
                        {connection.nodes.map(node => (
                            <ExternalServiceNode
                                key={node.id}
                                node={node}
                                history={history}
                                routingPrefix={routingPrefix}
                                afterDeleteRoute={afterDeleteRoute}
                            />
                        ))}
                    </ConnectionList>
                )}
                {loading && <ConnectionLoading />}
                {!loading && connection && (
                    <SummaryContainer>
                        {summary}
                        {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
                    </SummaryContainer>
                )}
            </ConnectionContainer>
        </div>
    )
}

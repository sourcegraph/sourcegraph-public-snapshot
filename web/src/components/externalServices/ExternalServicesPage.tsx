import AddIcon from 'mdi-react/AddIcon'
import React, { useEffect, useMemo, useCallback } from 'react'
import { Redirect } from 'react-router'
import { Link } from 'react-router-dom'
import { Subject } from 'rxjs'
import { map, tap } from 'rxjs/operators'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../FilteredConnection'
import { PageTitle } from '../PageTitle'
import * as H from 'history'
import { queryExternalServices } from './backend'
import { ExternalServiceNodeProps, ExternalServiceNode } from './ExternalServiceNode'
import { ListExternalServiceFields, Scalars } from '../../graphql-operations'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { TelemetryProps } from '../../../../shared/src/telemetry/telemetryService'

interface Props extends ActivationProps, TelemetryProps {
    history: H.History
    location: H.Location
    routingPrefix: string
    afterDeleteRoute: string
    userID?: Scalars['ID']
}

/**
 * A page displaying the external services on this site.
 */
export const ExternalServicesPage: React.FunctionComponent<Props> = ({
    afterDeleteRoute,
    history,
    location,
    routingPrefix,
    activation,
    userID,
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('SiteAdminExternalServices')
    }, [telemetryService])
    const updates = useMemo(() => new Subject<void>(), [])
    const onDidUpdateExternalServices = useCallback(() => updates.next(), [updates])

    const noExternalServices = useObservable(
        useMemo(
            () =>
                queryExternalServices({ first: 1, after: null, namespace: userID ?? null }).pipe(
                    map(externalServicesResult => externalServicesResult.totalCount === 0)
                ),
            [userID]
        )
    )

    const queryConnection = useCallback(
        (args: FilteredConnectionQueryArgs) =>
            queryExternalServices({
                first: args.first ?? null,
                after: args.after ?? null,
                namespace: userID ?? null,
            }).pipe(
                tap(externalServices => {
                    if (activation && externalServices.totalCount > 0) {
                        activation.update({ ConnectedCodeHost: true })
                    }
                })
            ),
        [userID, activation]
    )

    if (noExternalServices === true) {
        return <Redirect to={`${routingPrefix}/external-services/new`} />
    }
    return (
        <div className="site-admin-external-services-page">
            <PageTitle title="Manage repositories - Admin" />
            <div className="d-flex justify-content-between align-items-center mb-3">
                <h2 className="mb-0">Manage repositories</h2>
                <Link
                    className="btn btn-primary test-goto-add-external-service-page"
                    to={`${routingPrefix}/external-services/new`}
                >
                    <AddIcon className="icon-inline" /> Add repositories
                </Link>
            </div>
            <p className="mt-2">Manage code host connections to sync repositories.</p>
            <FilteredConnection<ListExternalServiceFields, Omit<ExternalServiceNodeProps, 'node'>>
                className="list-group list-group-flush mt-3"
                noun="external service"
                pluralNoun="external services"
                queryConnection={queryConnection}
                nodeComponent={ExternalServiceNode}
                nodeComponentProps={{
                    onDidUpdate: onDidUpdateExternalServices,
                    history,
                    routingPrefix,
                    afterDeleteRoute,
                }}
                hideSearch={true}
                noSummaryIfAllNodesVisible={true}
                cursorPaging={true}
                updates={updates}
                history={history}
                location={location}
            />
        </div>
    )
}

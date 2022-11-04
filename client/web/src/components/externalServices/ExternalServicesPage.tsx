import React, { useEffect, useMemo, useCallback, useState } from 'react'

import { mdiPlus } from '@mdi/js'
import * as H from 'history'
import { Redirect } from 'react-router'
import { Subject } from 'rxjs'
import { tap } from 'rxjs/operators'

import { isErrorLike, ErrorLike } from '@sourcegraph/common'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Link, Button, Icon, PageHeader, Container } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { ListExternalServiceFields, Scalars, ExternalServicesResult } from '../../graphql-operations'
import { FilteredConnection, FilteredConnectionQueryArguments } from '../FilteredConnection'
import { PageTitle } from '../PageTitle'

import { queryExternalServices as _queryExternalServices } from './backend'
import { ExternalServiceNodeProps, ExternalServiceNode } from './ExternalServiceNode'

interface Props extends ActivationProps, TelemetryProps {
    history: H.History
    location: H.Location
    routingPrefix: string
    afterDeleteRoute: string
    userID?: Scalars['ID']
    authenticatedUser: Pick<AuthenticatedUser, 'id'>

    /** For testing only. */
    queryExternalServices?: typeof _queryExternalServices
}

/**
 * A page displaying the external services on this site.
 */
export const ExternalServicesPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    afterDeleteRoute,
    history,
    location,
    routingPrefix,
    activation,
    userID,
    telemetryService,
    authenticatedUser,
    queryExternalServices = _queryExternalServices,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('SiteAdminExternalServices')
    }, [telemetryService])
    const updates = useMemo(() => new Subject<void>(), [])
    const onDidUpdateExternalServices = useCallback(() => updates.next(), [updates])

    const queryConnection = useCallback(
        (args: FilteredConnectionQueryArguments) =>
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
        // Activation changes in here, so we cannot recreate the callback on change,
        // or queryConnection will constantly change, resulting in infinite refetch loops.
        // eslint-disable-next-line react-hooks/exhaustive-deps
        [userID, queryExternalServices]
    )

    const [noExternalServices, setNoExternalServices] = useState<boolean>(false)
    const onUpdate = useCallback<
        (connection: ExternalServicesResult['externalServices'] | ErrorLike | undefined) => void
    >(connection => {
        if (connection && !isErrorLike(connection)) {
            setNoExternalServices(connection.totalCount === 0)
        }
    }, [])

    const isManagingOtherUser = !!userID && userID !== authenticatedUser.id

    if (!isManagingOtherUser && noExternalServices) {
        return <Redirect to={`${routingPrefix}/external-services/new`} />
    }
    return (
        <div className="site-admin-external-services-page">
            <PageTitle title="Manage code hosts" />
            <PageHeader
                path={[{ text: 'Manage code hosts' }]}
                description="Manage code host connections to sync repositories."
                headingElement="h2"
                actions={
                    <>
                        {!isManagingOtherUser && (
                            <Button
                                className="test-goto-add-external-service-page"
                                to={`${routingPrefix}/external-services/new`}
                                variant="primary"
                                as={Link}
                            >
                                <Icon aria-hidden={true} svgPath={mdiPlus} /> Add code host
                            </Button>
                        )}
                    </>
                }
                className="mb-3"
            />

            <Container className="mb-3">
                <FilteredConnection<
                    ListExternalServiceFields,
                    Omit<ExternalServiceNodeProps, 'node'>,
                    {},
                    ExternalServicesResult['externalServices']
                >
                    className="mb-0"
                    listClassName="list-group list-group-flush mb-0"
                    noun="code host"
                    pluralNoun="code hosts"
                    withCenteredSummary={true}
                    queryConnection={queryConnection}
                    nodeComponent={ExternalServiceNode}
                    nodeComponentProps={{
                        onDidUpdate: onDidUpdateExternalServices,
                        history,
                        routingPrefix,
                        afterDeleteRoute,
                    }}
                    hideSearch={true}
                    cursorPaging={true}
                    updates={updates}
                    history={history}
                    location={location}
                    onUpdate={onUpdate}
                />
            </Container>
        </div>
    )
}

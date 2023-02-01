import React, { useEffect } from 'react'

import { mdiPlus } from '@mdi/js'
import { Redirect } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Link, ButtonLink, Icon, PageHeader, Container } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { Scalars } from '../../graphql-operations'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../FilteredConnection/ui'
import { PageTitle } from '../PageTitle'

import { useExternalServicesConnection } from './backend'
import { ExternalServiceEditingDisabledAlert } from './ExternalServiceEditingDisabledAlert'
import { ExternalServiceEditingTemporaryAlert } from './ExternalServiceEditingTemporaryAlert'
import { ExternalServiceNode } from './ExternalServiceNode'

interface Props extends TelemetryProps {
    routingPrefix: string
    userID?: Scalars['ID']
    authenticatedUser: Pick<AuthenticatedUser, 'id'>

    externalServicesFromFile: boolean
    allowEditExternalServicesWithFile: boolean
}

/**
 * A page displaying the external services on this site.
 */
export const ExternalServicesPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    routingPrefix,
    userID,
    telemetryService,
    authenticatedUser,
    externalServicesFromFile,
    allowEditExternalServicesWithFile,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('SiteAdminExternalServices')
    }, [telemetryService])

    const { loading, hasNextPage, fetchMore, connection, error } = useExternalServicesConnection({
        first: null,
        after: null,
    })

    const editingDisabled = externalServicesFromFile && !allowEditExternalServicesWithFile
    const isManagingOtherUser = !!userID && userID !== authenticatedUser.id

    return !isManagingOtherUser && !loading && (connection?.nodes?.length ?? 0) === 0 ? (
        <Redirect to={`${routingPrefix}/external-services/new`} />
    ) : (
        <div className="site-admin-external-services-page">
            <PageTitle title="Manage code hosts" />
            <PageHeader
                path={[{ text: 'Manage code hosts' }]}
                description="Manage code host connections to sync repositories."
                headingElement="h2"
                actions={
                    <>
                        {!isManagingOtherUser && (
                            <ButtonLink
                                className="test-goto-add-external-service-page"
                                to={`${routingPrefix}/external-services/new`}
                                variant="primary"
                                as={Link}
                                disabled={editingDisabled}
                            >
                                <Icon aria-hidden={true} svgPath={mdiPlus} /> Add code host
                            </ButtonLink>
                        )}
                    </>
                }
                className="mb-3"
            />

            {editingDisabled && <ExternalServiceEditingDisabledAlert />}
            {externalServicesFromFile && allowEditExternalServicesWithFile && <ExternalServiceEditingTemporaryAlert />}

            <Container className="mb-3">
                <ConnectionContainer>
                    {error && <ConnectionError errors={[error.message]} />}
                    {loading && !connection && <ConnectionLoading />}
                    <ConnectionList as="ul" className="list-group" aria-label="CodeHosts">
                        {connection?.nodes?.map(node => (
                            <ExternalServiceNode
                                key={node.id}
                                node={node}
                                routingPrefix={routingPrefix}
                                editingDisabled={editingDisabled}
                            />
                        ))}
                    </ConnectionList>
                    {connection && (
                        <SummaryContainer className="mt-2" centered={true}>
                            <ConnectionSummary
                                noSummaryIfAllNodesVisible={false}
                                first={connection.totalCount ?? 0}
                                centered={true}
                                connection={connection}
                                noun="code host"
                                pluralNoun="code hosts"
                                hasNextPage={hasNextPage}
                            />
                            {hasNextPage && <ShowMoreButton centered={true} onClick={fetchMore} />}
                        </SummaryContainer>
                    )}
                </ConnectionContainer>
            </Container>
        </div>
    )
}

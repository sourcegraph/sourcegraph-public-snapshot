import { FC, useEffect } from 'react'

import { mdiPlus } from '@mdi/js'
import { Navigate } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Link, ButtonLink, Icon, PageHeader, Container } from '@sourcegraph/wildcard'

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
import { ExternalServiceEditingAppLimitAlert } from './ExternalServiceEditingAppLimitReachedAlert'
import { ExternalServiceEditingDisabledAlert } from './ExternalServiceEditingDisabledAlert'
import { ExternalServiceEditingTemporaryAlert } from './ExternalServiceEditingTemporaryAlert'
import { ExternalServiceNode } from './ExternalServiceNode.tsx'
import { isAppLocalFileService } from './isAppLocalFileService'

interface Props extends TelemetryProps {
    externalServicesFromFile: boolean
    allowEditExternalServicesWithFile: boolean
    isSourcegraphApp: boolean
}

/**
 * A page displaying the external services on this site.
 */
export const ExternalServicesPage: FC<Props> = ({
    telemetryService,
    externalServicesFromFile,
    allowEditExternalServicesWithFile,
    isSourcegraphApp,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('SiteAdminExternalServices')
    }, [telemetryService])

    const { loading, hasNextPage, fetchMore, connection, error } = useExternalServicesConnection({
        first: null,
        after: null,
    })

    const editingDisabled = externalServicesFromFile && !allowEditExternalServicesWithFile

    const externalServices = connection?.nodes ? connection?.nodes?.filter(node => !isAppLocalFileService(node)) : []
    const appLimitReached = isSourcegraphApp && externalServices.length >= 1

    return !loading && (connection?.nodes?.length ?? 0) === 0 ? (
        <Navigate to="/site-admin/external-services/new" replace={true} />
    ) : (
        <div className="site-admin-external-services-page">
            <PageTitle title="Manage code hosts" />
            <PageHeader
                path={[{ text: 'Manage code hosts' }]}
                description="Manage code host connections to sync repositories."
                headingElement="h2"
                actions={
                    <ButtonLink
                        className="test-goto-add-external-service-page"
                        to="/site-admin/external-services/new"
                        variant="primary"
                        as={Link}
                        disabled={editingDisabled || appLimitReached}
                    >
                        <Icon aria-hidden={true} svgPath={mdiPlus} /> Add code host
                    </ButtonLink>
                }
                className="mb-3"
            />

            {editingDisabled && <ExternalServiceEditingDisabledAlert />}
            {isSourcegraphApp && <ExternalServiceEditingAppLimitAlert />}
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
                                editingDisabled={editingDisabled}
                                isSourcegraphApp={isSourcegraphApp}
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

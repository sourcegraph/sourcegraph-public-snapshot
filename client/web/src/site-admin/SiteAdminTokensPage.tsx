import React, { useCallback, useMemo } from 'react'

import { mdiPlus } from '@mdi/js'
import { Subject } from 'rxjs'

import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, ButtonLink, Container, Icon, PageHeader } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../auth'
import { FilteredConnection } from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import type { AccessTokenFields } from '../graphql-operations'
import { AccessTokenNode, type AccessTokenNodeProps } from '../settings/tokens/AccessTokenNode'

import { queryAccessTokens } from './backend'

interface Props extends TelemetryProps {
    authenticatedUser: AuthenticatedUser
}

/**
 * Displays a list of all access tokens on the site.
 */
export const SiteAdminTokensPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    authenticatedUser,
    telemetryService,
}) => {
    useMemo(() => {
        telemetryService.logViewEvent('SiteAdminTokens')
    }, [telemetryService])
    const accessTokenUpdates = useMemo(() => new Subject<void>(), [])
    const onDidUpdateAccessToken = useCallback(() => accessTokenUpdates.next(), [accessTokenUpdates])
    const accessTokensEnabled = window.context.accessTokensAllow !== 'none'
    return (
        <div className="user-settings-tokens-page">
            <PageTitle title="Access tokens - Admin" />
            <PageHeader
                path={[{ text: 'Access tokens' }]}
                headingElement="h2"
                description={
                    <>
                        Tokens may be used to access the Sourcegraph API with the full privileges of the token's
                        creator.
                    </>
                }
                actions={
                    <>
                        {accessTokensEnabled && (
                            <ButtonLink
                                variant="primary"
                                className="ml-2"
                                to={`${authenticatedUser.settingsURL!}/tokens/new`}
                            >
                                <Icon aria-hidden={true} svgPath={mdiPlus} /> Generate access token
                            </ButtonLink>
                        )}
                        {!accessTokensEnabled && (
                            <Button
                                variant="primary"
                                title="Access token creation is disabled in site configuration"
                                className="ml-2"
                                disabled={true}
                            >
                                <Icon aria-hidden={true} svgPath={mdiPlus} /> Generate access token
                            </Button>
                        )}
                    </>
                }
                className="mb-3"
            />
            <Container className="mb-3">
                <FilteredConnection<AccessTokenFields, Omit<AccessTokenNodeProps, 'node'>>
                    className="list-group list-group-flush mb-0"
                    noun="access token"
                    pluralNoun="access tokens"
                    queryConnection={queryAccessTokens}
                    nodeComponent={AccessTokenNode}
                    nodeComponentProps={{
                        showSubject: true,
                        afterDelete: onDidUpdateAccessToken,
                    }}
                    updates={accessTokenUpdates}
                    hideSearch={true}
                />
            </Container>
        </div>
    )
}

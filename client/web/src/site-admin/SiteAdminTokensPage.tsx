import React, { useCallback, useMemo } from 'react'

import classNames from 'classnames'
import AddIcon from 'mdi-react/AddIcon'
import { RouteComponentProps } from 'react-router'
import { Observable, Subject } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Icon, Typography } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { requestGraphQL } from '../backend/graphql'
import { FilteredConnection } from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import {
    AccessTokenFields,
    SiteAdminAccessTokenConnectionFields,
    SiteAdminAccessTokensResult,
    SiteAdminAccessTokensVariables,
} from '../graphql-operations'
import { accessTokenFragment, AccessTokenNode, AccessTokenNodeProps } from '../settings/tokens/AccessTokenNode'

interface Props extends Pick<RouteComponentProps<{}>, 'history' | 'location'>, TelemetryProps {
    authenticatedUser: AuthenticatedUser
}

/**
 * Displays a list of all access tokens on the site.
 */
export const SiteAdminTokensPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    history,
    location,
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
            <div className="d-flex justify-content-between align-items-center mb-3">
                <Typography.H2 className="mb-0">Access tokens</Typography.H2>
                <Button
                    as={LinkOrSpan}
                    title={accessTokensEnabled ? '' : 'Access token creation is disabled in site configuration'}
                    className={classNames('ml-2', !accessTokensEnabled && 'disabled')}
                    to={accessTokensEnabled ? `${authenticatedUser.settingsURL!}/tokens/new` : null}
                >
                    <Icon as={AddIcon} /> Generate access token
                </Button>
            </div>
            <p>Tokens may be used to access the Sourcegraph API with the full privileges of the token's creator.</p>
            <FilteredConnection<AccessTokenFields, Omit<AccessTokenNodeProps, 'node'>>
                className="list-group list-group-flush mt-3"
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
                noSummaryIfAllNodesVisible={true}
                history={history}
                location={location}
            />
        </div>
    )
}

function queryAccessTokens(args: { first?: number }): Observable<SiteAdminAccessTokenConnectionFields> {
    return requestGraphQL<SiteAdminAccessTokensResult, SiteAdminAccessTokensVariables>(
        gql`
            query SiteAdminAccessTokens($first: Int) {
                site {
                    accessTokens(first: $first) {
                        ...SiteAdminAccessTokenConnectionFields
                    }
                }
            }
            fragment SiteAdminAccessTokenConnectionFields on AccessTokenConnection {
                nodes {
                    ...AccessTokenFields
                }
                totalCount
                pageInfo {
                    hasNextPage
                }
            }
            ${accessTokenFragment}
        `,
        { first: args.first ?? null }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.site.accessTokens)
    )
}

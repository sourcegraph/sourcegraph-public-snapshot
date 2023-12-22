import React, { useCallback, useEffect, useMemo } from 'react'

import { mdiPlus } from '@mdi/js'
import { type Observable, Subject } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, PageHeader, Button, Icon, Text, ButtonLink, Tooltip } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../../../backend/graphql'
import { FilteredConnection } from '../../../components/FilteredConnection'
import { PageTitle } from '../../../components/PageTitle'
import type {
    AccessTokenFields,
    AccessTokensConnectionFields,
    AccessTokensResult,
    AccessTokensVariables,
    CreateAccessTokenResult,
} from '../../../graphql-operations'
import {
    accessTokenFragment,
    AccessTokenNode,
    type AccessTokenNodeProps,
} from '../../../settings/tokens/AccessTokenNode'
import type { UserSettingsAreaRouteContext } from '../UserSettingsArea'

interface Props extends Pick<UserSettingsAreaRouteContext, 'authenticatedUser' | 'user'>, TelemetryProps {
    /**
     * The newly created token, if any. This component must call onDidPresentNewToken
     * when it is finished presenting the token secret to the user.
     */
    newToken?: CreateAccessTokenResult['createAccessToken']

    /**
     * Called when the newly created access token has been presented to the user and may be purged
     * from all state (and not displayed to the user anymore).
     */
    onDidPresentNewToken: () => void
}

/**
 * Displays access tokens whose subject is a specific user.
 */
export const UserSettingsTokensPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    telemetryService,
    authenticatedUser,
    user,
    newToken,
    onDidPresentNewToken,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('UserSettingsTokens')
    }, [telemetryService])

    useEffect(
        () => () => {
            // Clear the newly created access token value from our application state; we assume the user
            // has already stored it elsewhere.
            onDidPresentNewToken()
        },
        [onDidPresentNewToken]
    )

    const accessTokenUpdates = useMemo(() => new Subject<void>(), [])
    const onDeleteAccessToken = useCallback(() => {
        accessTokenUpdates.next()
    }, [accessTokenUpdates])

    const queryUserAccessTokens = useCallback(
        (args: { first?: number }) => queryAccessTokens({ first: args.first ?? null, user: user.id }),
        [user.id]
    )

    const siteAdminViewingOtherUser = authenticatedUser && authenticatedUser.id !== user.id
    const accessTokensEnabled =
        (authenticatedUser.siteAdmin && window.context.accessTokensAllow === 'site-admin-create') ||
        (!siteAdminViewingOtherUser && window.context.accessTokensAllow === 'all-users-create')

    return (
        <div className="user-settings-tokens-page">
            <PageTitle title="Access tokens" />
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Access tokens' }]}
                description="Access tokens may be used to access the Sourcegraph API."
                actions={
                    <>
                        {accessTokensEnabled && (
                            <ButtonLink variant="primary" className="ml-2" to="new">
                                <Icon aria-hidden={true} svgPath={mdiPlus} /> Generate new token
                            </ButtonLink>
                        )}
                        {!accessTokensEnabled && (
                            <Tooltip
                                content={
                                    siteAdminViewingOtherUser
                                        ? 'Access token creation for other users is disabled in site configuration'
                                        : 'Access token creation is disabled in site configuration'
                                }
                            >
                                <Button variant="primary" className="ml-2" disabled={true}>
                                    <Icon aria-hidden={true} svgPath={mdiPlus} /> Generate new token
                                </Button>
                            </Tooltip>
                        )}
                    </>
                }
                className="mb-3"
            />
            <Container>
                <FilteredConnection<AccessTokenFields, Omit<AccessTokenNodeProps, 'node'>>
                    listClassName="list-group list-group-flush"
                    noun="access token"
                    pluralNoun="access tokens"
                    queryConnection={queryUserAccessTokens}
                    nodeComponent={AccessTokenNode}
                    nodeComponentProps={{
                        afterDelete: onDeleteAccessToken,
                        showSubject: false,
                        newToken,
                    }}
                    updates={accessTokenUpdates}
                    hideSearch={true}
                    noSummaryIfAllNodesVisible={true}
                    emptyElement={
                        <Text alignment="center" className="text-muted w-100 mb-0">
                            You don't have any access tokens.
                        </Text>
                    }
                />
            </Container>
        </div>
    )
}

const queryAccessTokens = (variables: AccessTokensVariables): Observable<AccessTokensConnectionFields> =>
    requestGraphQL<AccessTokensResult, AccessTokensVariables>(
        gql`
            query AccessTokens($user: ID!, $first: Int) {
                node(id: $user) {
                    __typename
                    ... on User {
                        accessTokens(first: $first) {
                            ...AccessTokensConnectionFields
                        }
                    }
                }
            }
            fragment AccessTokensConnectionFields on AccessTokenConnection {
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
        variables
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.node) {
                throw new Error('User not found')
            }
            if (data.node.__typename !== 'User') {
                throw new Error(`Node is a ${data.node.__typename}, not a User`)
            }
            return data.node.accessTokens
        })
    )

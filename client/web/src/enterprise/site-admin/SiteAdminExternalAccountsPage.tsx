import * as React from 'react'

import { type Observable, Subject, Subscription } from 'rxjs'
import { map } from 'rxjs/operators'

import { createAggregateError } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import type { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Button, Link, H2, Text } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../../backend/graphql'
import { FilteredConnection } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import type {
    ExternalAccountFields,
    ExternalAccountsConnectionFields,
    ExternalAccountsResult,
    ExternalAccountsVariables,
} from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import {
    ExternalAccountNode,
    type ExternalAccountNodeProps,
    externalAccountsConnectionFragment,
} from '../user/settings/ExternalAccountNode'

interface Props extends TelemetryV2Props {}

interface FilterParameters {
    user?: Scalars['ID']
    serviceType?: string
    serviceID?: string
    clientID?: string
}

/**
 * Displays the external accounts (from authentication providers) associated with the user's account.
 */
export class SiteAdminExternalAccountsPage extends React.Component<Props> {
    private subscriptions = new Subscription()
    private externalAccountUpdates = new Subject<void>()

    public componentDidMount(): void {
        this.props.telemetryRecorder.recordEvent('siteAdminExternalAccountsPage', 'viewed')
        eventLogger.logViewEvent('SiteAdminExternalAccounts')
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const nodeProps: Omit<ExternalAccountNodeProps, 'node'> = {
            onDidUpdate: this.onDidUpdateExternalAccount,
            showUser: true,
        }

        return (
            <div className="user-settings-external-accounts-page">
                <PageTitle title="External accounts" />
                <div className="d-flex justify-content-between align-items-center mb-3">
                    <H2 className="mb-0">External user accounts</H2>
                    <Button to="/site-admin/auth/providers" variant="secondary" as={Link}>
                        View auth providers
                    </Button>
                </div>
                <Text>
                    An external account (on an <Link to="/site-admin/auth/providers">authentication provider</Link>) is
                    linked to a Sourcegraph user when it's used to sign into Sourcegraph.
                </Text>
                <FilteredConnection<ExternalAccountFields, Omit<ExternalAccountNodeProps, 'node'>>
                    className="list-group list-group-flush mt-3"
                    noun="external user account"
                    pluralNoun="external user accounts"
                    queryConnection={this.queryExternalAccounts}
                    nodeComponent={ExternalAccountNode}
                    nodeComponentProps={nodeProps}
                    updates={this.externalAccountUpdates}
                    hideSearch={true}
                />
            </div>
        )
    }

    private queryExternalAccounts = (
        args: {
            first?: number
        } & FilterParameters
    ): Observable<ExternalAccountsConnectionFields> =>
        requestGraphQL<ExternalAccountsResult, ExternalAccountsVariables>(
            gql`
                query ExternalAccounts(
                    $first: Int
                    $user: ID
                    $serviceType: String
                    $serviceID: String
                    $clientID: String
                ) {
                    site {
                        externalAccounts(
                            first: $first
                            user: $user
                            serviceType: $serviceType
                            serviceID: $serviceID
                            clientID: $clientID
                        ) {
                            ...ExternalAccountsConnectionFields
                        }
                    }
                }

                ${externalAccountsConnectionFragment}
            `,
            {
                clientID: args.clientID ?? null,
                first: args.first ?? null,
                serviceID: args.serviceID ?? null,
                serviceType: args.serviceType ?? null,
                user: args.user ?? null,
            }
        ).pipe(
            map(({ data, errors }) => {
                if (!data?.site?.externalAccounts) {
                    throw createAggregateError(errors)
                }
                return data.site.externalAccounts
            })
        )

    private onDidUpdateExternalAccount = (): void => this.externalAccountUpdates.next()
}

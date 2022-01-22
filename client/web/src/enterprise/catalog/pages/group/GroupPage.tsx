import React, { useEffect } from 'react'

import { useQuery } from '@sourcegraph/http-client'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../../components/PageTitle'
import { GroupByNameResult, GroupByNameVariables } from '../../../../graphql-operations'

import { GROUP_BY_NAME } from './gql'
import { GroupDetailContent } from './GroupContent'

export interface Props extends TelemetryProps, ExtensionsControllerProps, ThemeProps, SettingsCascadeProps {
    /** The name of the group. */
    groupName: string
}

/**
 * The group detail page.
 */
export const GroupDetailPage: React.FunctionComponent<Props> = ({ groupName, telemetryService, ...props }) => {
    useEffect(() => {
        telemetryService.logViewEvent('GroupDetail')
    }, [telemetryService])

    const { data, error, loading } = useQuery<GroupByNameResult, GroupByNameVariables>(GROUP_BY_NAME, {
        variables: { name: groupName },

        // Cache this data but always re-request it in the background when we revisit
        // this page to pick up newer changes.
        fetchPolicy: 'cache-and-network',

        // For subsequent requests while this page is open, make additional network
        // requests; this is necessary for `refetch` to actually use the network. (see
        // https://github.com/apollographql/apollo-client/issues/5515)
        nextFetchPolicy: 'network-only',
    })

    return (
        <>
            <PageTitle
                title={
                    error
                        ? 'Error loading group'
                        : loading && !data
                        ? 'Loading group...'
                        : !data || !data.group
                        ? 'Group not found'
                        : data.group.name
                }
            />
            {loading && !data ? (
                <LoadingSpinner className="m-3 icon-inline" />
            ) : error && !data ? (
                <div className="m-3 alert alert-danger">Error: {error.message}</div>
            ) : !data || !data.group ? (
                <div className="m-3 alert alert-danger">Group not found</div>
            ) : (
                <GroupDetailContent {...props} group={data.group} telemetryService={telemetryService} />
            )}
        </>
    )
}

import React, { useEffect, useMemo } from 'react'

import { isDefined } from '@sourcegraph/common'
import { gql, useQuery } from '@sourcegraph/http-client'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { CatalogIcon } from '../../../../catalog'
import { PageTitle } from '../../../../components/PageTitle'
import { GroupPageResult, GroupPageVariables } from '../../../../graphql-operations'
import { CatalogPage } from '../../components/catalog-area-header/CatalogPage'
import { CatalogGroupIcon } from '../../components/CatalogGroupIcon'
import { GROUP_LINK_FRAGMENT } from '../../components/group-link/GroupLink'

import { GroupMembersTab, GROUP_MEMBERS_TAB_FRAGMENT } from './tabs/members/GroupMembersTab'
import { GroupOverviewTab, GROUP_OVERVIEW_TAB_FRAGMENT } from './tabs/overview/GroupOverviewTab'

export interface Props extends TelemetryProps, ExtensionsControllerProps, ThemeProps, SettingsCascadeProps {
    /** The name of the group. */
    groupName: string
}

/**
 * The catalog group page.
 */
export const GroupPage: React.FunctionComponent<Props> = ({ groupName, telemetryService, ...props }) => {
    useEffect(() => {
        telemetryService.logViewEvent('GroupDetail')
    }, [telemetryService])

    const { data, error, loading } = useQuery<GroupPageResult, GroupPageVariables>(
        gql`
            query GroupPage($name: String!) {
                group(name: $name) {
                    __typename
                    id
                    name
                    url

                    ancestorGroups {
                        ...GroupLinkFields2
                    }

                    ...GroupOverviewTabFields
                    ...GroupMembersTabFields
                }
            }

            ${GROUP_MEMBERS_TAB_FRAGMENT}
            ${GROUP_OVERVIEW_TAB_FRAGMENT}
            ${GROUP_LINK_FRAGMENT}
        `,
        {
            variables: { name: groupName },

            // Cache this data but always re-request it in the background when we revisit
            // this page to pick up newer changes.
            fetchPolicy: 'cache-and-network',

            // For subsequent requests while this page is open, make additional network
            // requests; this is necessary for `refetch` to actually use the network. (see
            // https://github.com/apollographql/apollo-client/issues/5515)
            nextFetchPolicy: 'network-only',
        }
    )

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
                <GroupPageContent {...props} group={data.group} />
            )}
        </>
    )
}

const TAB_CONTENT_CLASS_NAME = 'flex-1 align-self-stretch overflow-auto'

const GroupPageContent: React.FunctionComponent<{
    group: NonNullable<GroupPageResult['group']>
}> = ({ group }) => {
    const tabs = useMemo<React.ComponentProps<typeof CatalogPage>['tabs']>(
        () =>
            [
                {
                    path: '',
                    exact: true,
                    text: 'Overview',
                    content: <GroupOverviewTab group={group} className={TAB_CONTENT_CLASS_NAME} />,
                },
                {
                    path: 'members',
                    exact: true,
                    text: 'Members',
                    content: <GroupMembersTab group={group} className={TAB_CONTENT_CLASS_NAME} />,
                },
                // TODO(sqs): show group code/changes/etc. tabs
            ].filter(isDefined),
        [group]
    )
    return (
        <CatalogPage
            path={[
                { icon: CatalogIcon, to: '/catalog' },
                ...group.ancestorGroups.map(group => ({ icon: CatalogGroupIcon, text: group.name, to: group.url })),
                {
                    icon: CatalogGroupIcon,
                    text: group.name,
                    to: group.url,
                },
            ].filter(isDefined)}
            tabs={tabs}
        />
    )
}

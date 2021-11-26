import React, { useMemo } from 'react'

import { isDefined } from '@sourcegraph/common'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { CatalogIcon } from '../../../../catalog'
import { GroupDetailFields } from '../../../../graphql-operations'
import { CatalogPage } from '../../components/catalog-area-header/CatalogPage'
import { CatalogGroupIcon } from '../../components/CatalogGroupIcon'

import { GroupMembersTab } from './GroupMembersTab'
import { GroupOverviewTab } from './OverviewTab'

interface Props extends TelemetryProps, ExtensionsControllerProps, ThemeProps, SettingsCascadeProps {
    group: GroupDetailFields
}

const TAB_CONTENT_CLASS_NAME = 'flex-1 align-self-stretch overflow-auto'

export const GroupDetailContent: React.FunctionComponent<Props> = ({ group }) => {
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

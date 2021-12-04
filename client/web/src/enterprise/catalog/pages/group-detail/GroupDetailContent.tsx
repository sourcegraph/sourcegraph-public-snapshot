import classNames from 'classnames'
import React, { useMemo } from 'react'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { isDefined } from '@sourcegraph/shared/src/util/types'

import { CatalogIcon } from '../../../../catalog'
import { GroupDetailFields } from '../../../../graphql-operations'
import { CatalogPage } from '../../components/catalog-area-header/CatalogPage'
import { CatalogGroupIcon } from '../../components/CatalogGroupIcon'

import styles from './GroupDetailContent.module.scss'
import { GroupMembersTab } from './GroupMembersTab'
import { GroupOverviewTab } from './GroupOverviewTab'

interface Props extends TelemetryProps, ExtensionsControllerProps, ThemeProps, SettingsCascadeProps {
    group: GroupDetailFields
}

export interface GroupDetailContentCardProps {
    className?: string
    headerClassName?: string
    titleClassName?: string
    bodyClassName?: string
}

const cardProps: GroupDetailContentCardProps = {
    headerClassName: classNames('card-header', styles.cardHeader),
    titleClassName: classNames('card-title', styles.cardTitle),
    bodyClassName: classNames('card-body', styles.cardBody),
}

export const GroupDetailContent: React.FunctionComponent<Props> = ({ group, ...props }) => {
    const tabs = useMemo<React.ComponentProps<typeof CatalogPage>['tabs']>(
        () =>
            [
                {
                    path: '',
                    exact: true,
                    text: 'Overview',
                    content: <GroupOverviewTab {...cardProps} group={group} />,
                },
                {
                    path: 'members',
                    exact: true,
                    text: 'Members',
                    content: <GroupMembersTab group={group} className="p-3" />,
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
                },
            ].filter(isDefined)}
            tabs={tabs}
        />
    )
}

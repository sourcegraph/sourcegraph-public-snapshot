import classNames from 'classnames'
import SettingsIcon from 'mdi-react/SettingsIcon'
import React, { useMemo } from 'react'
import { Link } from 'react-router-dom'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { isDefined } from '@sourcegraph/shared/src/util/types'
import { PageHeader } from '@sourcegraph/wildcard'

import { CatalogIcon } from '../../../../catalog'
import { GroupDetailFields } from '../../../../graphql-operations'
import { CatalogGroupIcon } from '../../components/CatalogGroupIcon'
import { TabRouter } from '../entity-detail/global/TabRouter'

import styles from './GroupDetailContent.module.scss'
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
    const tabs = useMemo<React.ComponentProps<typeof TabRouter>['tabs']>(
        () =>
            [
                {
                    path: '',
                    exact: true,
                    label: 'Overview',
                    element: <GroupOverviewTab {...cardProps} group={group} />,
                },
                // TODO(sqs): show group code/changes/etc. tabs
            ].filter(isDefined),
        [group]
    )
    return (
        <div>
            <PageHeader
                path={[
                    { icon: CatalogIcon, to: '/catalog' },
                    ...group.ancestorGroups.map(group => ({ icon: CatalogGroupIcon, text: group.name, to: group.url })),
                    {
                        icon: CatalogGroupIcon,
                        text: group.name,
                    },
                ].filter(isDefined)}
                actions={
                    // eslint-disable-next-line react/forbid-dom-props
                    <nav className="d-flex align-items-center">
                        <Link to="#" className="d-inline-block btn btn-secondary btn-sm p-2 mb-0">
                            <SettingsIcon className="icon-inline" />
                        </Link>
                    </nav>
                }
                className="mt-3 mb-1"
            />
            <div className="mb-4" />
            <TabRouter tabs={tabs} />
        </div>
    )
}

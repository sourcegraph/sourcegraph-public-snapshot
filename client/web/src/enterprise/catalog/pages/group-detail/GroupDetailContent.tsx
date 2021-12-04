import classNames from 'classnames'
import React, { useMemo } from 'react'
import { useRouteMatch } from 'react-router'
import { NavLink } from 'react-router-dom'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { isDefined } from '@sourcegraph/shared/src/util/types'

import { CatalogIcon } from '../../../../catalog'
import { GroupDetailFields } from '../../../../graphql-operations'
import { CatalogAreaHeader } from '../../components/catalog-area-header/CatalogAreaHeader'
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
    const match = useRouteMatch()
    return (
        <>
            <CatalogAreaHeader
                path={[
                    { icon: CatalogIcon, to: '/catalog' },
                    ...group.ancestorGroups.map(group => ({ icon: CatalogGroupIcon, text: group.name, to: group.url })),
                    {
                        icon: CatalogGroupIcon,
                        text: group.name,
                    },
                ].filter(isDefined)}
                nav={
                    <ul className="nav nav-tabs">
                        {tabs.map(tab => (
                            <li key={tab.path} className="nav-item">
                                <NavLink
                                    to={tabPath(match.url, tab)}
                                    exact={tab.exact}
                                    className="nav-link px-3"
                                    // TODO(sqs): hack so that active items when bolded don't shift the ones to the right over by a few px because bold text is wider
                                    style={{ minWidth: '6rem' }}
                                >
                                    {tab.label}
                                </NavLink>
                            </li>
                        ))}
                    </ul>
                }
            />
            <TabRouter tabs={tabs} />
        </>
    )
}

function tabPath(basePath: string, tab: Pick<Tab, 'path'>): string {
    return tab.path ? `${basePath}/${tab.path}` : basePath
}

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

import { GroupOverviewTab } from './GroupOverviewTab'
import styles from './GroupDetailContent.module.scss'

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
                    {
                        icon: CatalogGroupIcon,
                        text: group.name,
                    },
                ]}
                actions={
                    // eslint-disable-next-line react/forbid-dom-props
                    <nav className="d-flex align-items-center" style={{ marginBottom: '-5px' }}>
                        <div className="d-inline-block mr-4">
                            <span className="small font-weight-bold">Parent group</span>
                            <br />
                            ParentGroup
                        </div>
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

import React, { useMemo } from 'react'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { isDefined } from '@sourcegraph/shared/src/util/types'

import { CatalogIcon } from '../../../../catalog'
import { PackageDetailFields } from '../../../../graphql-operations'
import { CatalogPage } from '../../components/catalog-area-header/CatalogPage'
import { catalogEntityIconComponent } from '../../components/CatalogEntityIcon'
import { CatalogGroupIcon } from '../../components/CatalogGroupIcon'

import { PackageOverviewTab } from './PackageOverviewTab'

interface Props
    extends TelemetryProps,
        ExtensionsControllerProps,
        ThemeProps,
        SettingsCascadeProps,
        PlatformContextProps {
    entity: PackageDetailFields
}

const TAB_CONTENT_CLASS_NAME = 'flex-1 align-self-stretch overflow-auto'

export const PackageDetailContent: React.FunctionComponent<Props> = ({ entity, ...props }) => {
    const tabs = useMemo<React.ComponentProps<typeof CatalogPage>['tabs']>(
        () =>
            [
                {
                    path: '',
                    exact: true,
                    text: 'Overview',
                    content: <PackageOverviewTab {...props} entity={entity} className={TAB_CONTENT_CLASS_NAME} />,
                },
            ].filter(isDefined),
        [entity, props]
    )
    return (
        <CatalogPage
            path={[
                { icon: CatalogIcon, to: '/catalog' },
                ...(entity.owner
                    ? [...entity.owner.ancestorGroups, entity.owner].map(owner => ({
                          icon: CatalogGroupIcon,
                          text: owner.name,
                          to: owner.url,
                      }))
                    : []),
                {
                    icon: catalogEntityIconComponent(entity),
                    text: entity.name,
                    to: entity.url,
                },
            ].filter(isDefined)}
            tabs={tabs}
        />
    )
}

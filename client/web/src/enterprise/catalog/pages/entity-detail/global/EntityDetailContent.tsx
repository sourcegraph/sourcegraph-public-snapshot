import classNames from 'classnames'
import React, { useMemo } from 'react'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { isDefined } from '@sourcegraph/shared/src/util/types'

import { CatalogIcon } from '../../../../../catalog'
import { CatalogEntityDetailFields } from '../../../../../graphql-operations'
import { CatalogPage } from '../../../components/catalog-area-header/CatalogPage'
import { catalogEntityIconComponent } from '../../../components/CatalogEntityIcon'
import { CatalogGroupIcon } from '../../../components/CatalogGroupIcon'

import { ComponentAPI } from './ComponentApi'
import { ComponentDocumentation } from './ComponentDocumentation'
import { ComponentUsage } from './ComponentUsage'
import { EntityChangesTab } from './EntityChangesTab'
import { EntityCodeTab } from './EntityCodeTab'
import styles from './EntityDetailContent.module.scss'
import { EntityOverviewTab } from './EntityOverviewTab'

interface Props extends TelemetryProps, ExtensionsControllerProps, ThemeProps, SettingsCascadeProps {
    entity: CatalogEntityDetailFields
}

export interface EntityDetailContentCardProps {
    className?: string
    headerClassName?: string
    titleClassName?: string
    bodyClassName?: string
    bodyScrollableClassName?: string
}

const cardProps: EntityDetailContentCardProps = {
    headerClassName: classNames('card-header', styles.cardHeader),
    titleClassName: classNames('card-title', styles.cardTitle),
    bodyClassName: classNames('card-body', styles.cardBody),
    bodyScrollableClassName: styles.cardBodyScrollable,
}

export const EntityDetailContent: React.FunctionComponent<Props> = ({ entity, ...props }) => {
    const tabs = useMemo<React.ComponentProps<typeof CatalogPage>['tabs']>(
        () =>
            [
                {
                    path: '',
                    exact: true,
                    text: 'Overview',
                    content: <EntityOverviewTab {...cardProps} entity={entity} />,
                },
                entity.__typename === 'CatalogComponent'
                    ? {
                          path: 'code',
                          text: 'Code',
                          content: <EntityCodeTab {...props} {...cardProps} entity={entity} />,
                      }
                    : null,
                entity.__typename === 'CatalogComponent'
                    ? {
                          path: 'changes',
                          text: 'Changes',
                          content: <EntityChangesTab {...props} {...cardProps} entity={entity} />,
                      }
                    : null,
                false && entity.__typename === 'CatalogComponent'
                    ? {
                          path: 'docs',
                          text: 'Docs',
                          content: <ComponentDocumentation catalogComponent={entity} />,
                      }
                    : null,
                false && entity.__typename === 'CatalogComponent'
                    ? {
                          path: 'api',
                          text: 'API',
                          content: <ComponentAPI {...props} catalogComponent={entity} />,
                      }
                    : null,
                entity.__typename === 'CatalogComponent'
                    ? {
                          path: 'usage',
                          text: 'Usage',
                          content: <ComponentUsage {...props} {...cardProps} catalogComponent={entity} />,
                      }
                    : null,
            ].filter(isDefined),
        [entity, props]
    )
    return (
        <CatalogPage
            path={[
                { icon: CatalogIcon, to: '/catalog' },
                ...[...entity.owner.ancestorGroups, entity.owner].map(owner => ({
                    icon: CatalogGroupIcon,
                    text: owner.name,
                    to: owner.url,
                })),
                {
                    icon: catalogEntityIconComponent(entity),
                    text: entity.name,
                },
            ].filter(isDefined)}
            tabs={tabs}
        />
    )
}

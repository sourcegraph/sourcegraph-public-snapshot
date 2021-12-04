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
import { EntityOverviewTab } from './EntityOverviewTab'

interface Props extends TelemetryProps, ExtensionsControllerProps, ThemeProps, SettingsCascadeProps {
    entity: CatalogEntityDetailFields
}

const TAB_CONTENT_CLASS_NAME = 'flex-1 align-self-stretch'

export const EntityDetailContent: React.FunctionComponent<Props> = ({ entity, ...props }) => {
    const tabs = useMemo<React.ComponentProps<typeof CatalogPage>['tabs']>(
        () =>
            [
                {
                    path: '',
                    exact: true,
                    text: 'Overview',
                    content: <EntityOverviewTab entity={entity} className={TAB_CONTENT_CLASS_NAME} />,
                },
                entity.__typename === 'CatalogComponent'
                    ? {
                          path: 'code',
                          text: 'Code',
                          content: <EntityCodeTab {...props} entity={entity} className={TAB_CONTENT_CLASS_NAME} />,
                      }
                    : null,
                entity.__typename === 'CatalogComponent'
                    ? {
                          path: 'changes',
                          text: 'Changes',
                          content: <EntityChangesTab {...props} entity={entity} className={TAB_CONTENT_CLASS_NAME} />,
                      }
                    : null,
                false && entity.__typename === 'CatalogComponent'
                    ? {
                          path: 'docs',
                          text: 'Docs',
                          content: (
                              <ComponentDocumentation catalogComponent={entity} className={TAB_CONTENT_CLASS_NAME} />
                          ),
                      }
                    : null,
                false && entity.__typename === 'CatalogComponent'
                    ? {
                          path: 'api',
                          text: 'API',
                          content: (
                              <ComponentAPI {...props} catalogComponent={entity} className={TAB_CONTENT_CLASS_NAME} />
                          ),
                      }
                    : null,
                entity.__typename === 'CatalogComponent'
                    ? {
                          path: 'usage',
                          text: 'Usage',
                          content: (
                              <ComponentUsage {...props} catalogComponent={entity} className={TAB_CONTENT_CLASS_NAME} />
                          ),
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

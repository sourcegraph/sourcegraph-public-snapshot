import React, { useMemo } from 'react'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { isDefined } from '@sourcegraph/shared/src/util/types'

import { CatalogIcon } from '../../../../../catalog'
import { ComponentStateDetailFields } from '../../../../../graphql-operations'
import { CatalogPage } from '../../../components/catalog-area-header/CatalogPage'
import { componentIconComponent } from '../../../components/ComponentIcon'
import { CatalogGroupIcon } from '../../../components/CatalogGroupIcon'

import { ComponentAPI } from './ComponentApi'
import { ComponentDocumentation } from './ComponentDocumentation'
import { EntityChangesTab } from './EntityChangesTab'
import { EntityCodeTab } from './EntityCodeTab'
import { EntityOverviewTab } from './EntityOverviewTab'
import { EntityUsageTab } from './EntityUsageTab'
import { EntityWhoKnowsTab } from './EntityWhoKnowsTab'
import { EntityRelationsTab } from './EntityRelationsTab'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'

interface Props
    extends TelemetryProps,
        ExtensionsControllerProps,
        ThemeProps,
        SettingsCascadeProps,
        PlatformContextProps {
    entity: ComponentStateDetailFields
}

const TAB_CONTENT_CLASS_NAME = 'flex-1 align-self-stretch overflow-auto'

export const EntityDetailContent: React.FunctionComponent<Props> = ({ entity, ...props }) => {
    const tabs = useMemo<React.ComponentProps<typeof CatalogPage>['tabs']>(
        () =>
            [
                {
                    path: '',
                    exact: true,
                    text: 'Overview',
                    content: <EntityOverviewTab {...props} entity={entity} className={TAB_CONTENT_CLASS_NAME} />,
                },
                entity.__typename === 'Component'
                    ? {
                          path: 'code',
                          text: 'Code',
                          content: <EntityCodeTab {...props} entity={entity} className={TAB_CONTENT_CLASS_NAME} />,
                      }
                    : null,
                entity.__typename === 'Component'
                    ? {
                          path: 'relations',
                          text: 'Relations',
                          content: <EntityRelationsTab {...props} entity={entity} className={TAB_CONTENT_CLASS_NAME} />,
                      }
                    : null,

                entity.__typename === 'Component'
                    ? {
                          path: 'changes',
                          text: 'Changes',
                          content: <EntityChangesTab {...props} entity={entity} className={TAB_CONTENT_CLASS_NAME} />,
                      }
                    : null,
                false && entity.__typename === 'Component'
                    ? {
                          path: 'docs',
                          text: 'Docs',
                          content: (
                              <ComponentDocumentation component={entity} className={TAB_CONTENT_CLASS_NAME} />
                          ),
                      }
                    : null,
                false && entity.__typename === 'Component'
                    ? {
                          path: 'api',
                          text: 'API',
                          content: (
                              <ComponentAPI {...props} component={entity} className={TAB_CONTENT_CLASS_NAME} />
                          ),
                      }
                    : null,
                entity.__typename === 'Component'
                    ? {
                          path: 'usage',
                          text: 'Usage',
                          content: (
                              <EntityUsageTab {...props} component={entity} className={TAB_CONTENT_CLASS_NAME} />
                          ),
                      }
                    : null,
                entity.__typename === 'Component'
                    ? {
                          path: 'who-knows',
                          text: 'Who knows',
                          content: (
                              <EntityWhoKnowsTab
                                  {...props}
                                  component={entity}
                                  className={TAB_CONTENT_CLASS_NAME}
                              />
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
                    icon: componentIconComponent(entity),
                    text: entity.name,
                    to: entity.url,
                },
            ].filter(isDefined)}
            tabs={tabs}
        />
    )
}

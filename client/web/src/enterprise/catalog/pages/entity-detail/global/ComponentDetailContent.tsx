import React, { useMemo } from 'react'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { isDefined } from '@sourcegraph/shared/src/util/types'

import { CatalogIcon } from '../../../../../catalog'
import { ComponentStateDetailFields } from '../../../../../graphql-operations'
import { CatalogPage } from '../../../components/catalog-area-header/CatalogPage'
import { CatalogGroupIcon } from '../../../components/CatalogGroupIcon'
import { componentIconComponent } from '../../../components/ComponentIcon'

import { ChangesTab } from './ChangesTab'
import { CodeTab } from './CodeTab'
import { ComponentAPI } from './ComponentApi'
import { ComponentDocumentation } from './ComponentDocumentation'
import { OverviewTab } from './OverviewTab'
import { RelationsTab } from './RelationsTab'
import { UsageTab } from './UsageTab'
import { WhoKnowsTab } from './WhoKnowsTab'

interface Props
    extends TelemetryProps,
        ExtensionsControllerProps,
        ThemeProps,
        SettingsCascadeProps,
        PlatformContextProps {
    component: ComponentStateDetailFields
}

const TAB_CONTENT_CLASS_NAME = 'flex-1 align-self-stretch overflow-auto'

export const ComponentDetailContent: React.FunctionComponent<Props> = ({ component, ...props }) => {
    const tabs = useMemo<React.ComponentProps<typeof CatalogPage>['tabs']>(
        () =>
            [
                {
                    path: '',
                    exact: true,
                    text: 'Overview',
                    content: <OverviewTab {...props} entity={component} className={TAB_CONTENT_CLASS_NAME} />,
                },

                {
                    path: 'code',
                    text: 'Code',
                    content: <CodeTab {...props} entity={component} className={TAB_CONTENT_CLASS_NAME} />,
                },
                {
                    path: 'graph',
                    text: 'Graph',
                    content: <RelationsTab {...props} entity={component} className={TAB_CONTENT_CLASS_NAME} />,
                },
                {
                    path: 'changes',
                    text: 'Changes',
                    content: <ChangesTab {...props} entity={component} className={TAB_CONTENT_CLASS_NAME} />,
                },
                false
                    ? {
                          path: 'docs',
                          text: 'Docs',
                          content: <ComponentDocumentation component={component} className={TAB_CONTENT_CLASS_NAME} />,
                      }
                    : null,
                false
                    ? {
                          path: 'api',
                          text: 'API',
                          content: <ComponentAPI {...props} component={component} className={TAB_CONTENT_CLASS_NAME} />,
                      }
                    : null,
                {
                    path: 'usage',
                    text: 'Usage',
                    content: <UsageTab {...props} component={component} className={TAB_CONTENT_CLASS_NAME} />,
                },

                {
                    path: 'who-knows',
                    text: 'Who knows',
                    content: <WhoKnowsTab {...props} component={component} className={TAB_CONTENT_CLASS_NAME} />,
                },
            ].filter(isDefined),
        [component, props]
    )
    return (
        <CatalogPage
            path={[
                { icon: CatalogIcon, to: '/catalog' },
                ...[...component.owner.ancestorGroups, component.owner].map(owner => ({
                    icon: CatalogGroupIcon,
                    text: owner.name,
                    to: owner.url,
                })),
                {
                    icon: componentIconComponent(component),
                    text: component.name,
                    to: component.url,
                },
            ].filter(isDefined)}
            tabs={tabs}
        />
    )
}

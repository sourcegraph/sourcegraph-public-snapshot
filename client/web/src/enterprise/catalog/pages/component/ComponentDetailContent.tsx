import classNames from 'classnames'
import AccountIcon from 'mdi-react/AccountIcon'
import React, { useMemo } from 'react'

import { isDefined } from '@sourcegraph/common'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { CatalogIcon } from '../../../../catalog'
import { ComponentKind, ComponentOwnerFields, ComponentStateDetailFields } from '../../../../graphql-operations'
import { CatalogPage } from '../../components/catalog-area-header/CatalogPage'
import { CatalogGroupIcon } from '../../components/CatalogGroupIcon'
import { componentIconComponent } from '../../components/ComponentIcon'

import { CodeTab } from './code/CodeTab'
import styles from './ComponentDetailContent.module.scss'
import { ComponentHeaderActions } from './ComponentHeaderActions'
import { OverviewTab } from './overview/OverviewTab'
import { RelationsTab } from './RelationsTab'
import { UsageTab } from './UsageTab'

interface Props
    extends TelemetryProps,
        ExtensionsControllerProps,
        ThemeProps,
        SettingsCascadeProps,
        PlatformContextProps {
    component: ComponentStateDetailFields
}

export const TAB_CONTENT_CLASS_NAME = classNames('flex-1 align-self-stretch', styles.tabContent)

export const ComponentDetailContent: React.FunctionComponent<Props> = ({ component, ...props }) => {
    const tabs = useMemo<React.ComponentProps<typeof CatalogPage>['tabs']>(
        () =>
            [
                {
                    path: ['', 'who-knows'],
                    exact: true,
                    text: 'Overview',
                    content: <OverviewTab {...props} component={component} className={TAB_CONTENT_CLASS_NAME} />,
                },

                {
                    path: 'code',
                    text: 'Code',
                    content: <CodeTab {...props} component={component} className={TAB_CONTENT_CLASS_NAME} />,
                },
                {
                    path: 'graph',
                    text: 'Graph',
                    content: <RelationsTab {...props} component={component} className={TAB_CONTENT_CLASS_NAME} />,
                },
                component.usage
                    ? {
                          path: 'usage',
                          text: 'Usage',
                          content: (
                              <UsageTab {...props} sourceLocationSet={component} className={TAB_CONTENT_CLASS_NAME} />
                          ),
                      }
                    : null,
            ].filter(isDefined),
        [component, props]
    )
    return (
        <CatalogPage
            path={catalogPagePathForComponent(component)}
            tabs={tabs}
            actions={<ComponentHeaderActions component={component} />}
        />
    )
}

export const catalogPagePathForComponent = (
    component: { __typename: 'Component'; name: string; kind: ComponentKind; url: string } & ComponentOwnerFields
): React.ComponentProps<typeof CatalogPage>['path'] =>
    [
        { icon: CatalogIcon, to: '/catalog' },
        ...[...(component.owner?.__typename === 'Group' ? component.owner.ancestorGroups : []), component.owner]
            .filter(isDefined)
            .map(owner =>
                owner.__typename === 'Group'
                    ? {
                          icon: CatalogGroupIcon,
                          text: owner.name,
                          to: owner.url,
                      }
                    : owner.__typename === 'Person'
                    ? { icon: AccountIcon, text: owner.displayName, to: owner.user?.url }
                    : { text: 'Unknown' }
            ),
        {
            icon: componentIconComponent(component),
            text: component.name,
            to: component.url,
        },
    ].filter(isDefined)

import classNames from 'classnames'
import AccountIcon from 'mdi-react/AccountIcon'
import React from 'react'

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
import { catalogComponentIconComponent } from '../../components/ComponentIcon'

import styles from './ComponentDetailContent.module.scss'

interface Props
    extends TelemetryProps,
        ExtensionsControllerProps,
        ThemeProps,
        SettingsCascadeProps,
        PlatformContextProps {
    component: ComponentStateDetailFields
}

export const TAB_CONTENT_CLASS_NAME = classNames('flex-1 align-self-stretch', styles.tabContent)

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
            icon: catalogComponentIconComponent(component),
            text: component.name,
            to: component.url,
        },
    ].filter(isDefined)

import classNames from 'classnames'
import AccountIcon from 'mdi-react/AccountIcon'
import React from 'react'

import { isDefined } from '@sourcegraph/common'

import { CatalogIcon } from '../../../../catalog'
import { ComponentKind, ComponentOwnerLinkFields } from '../../../../graphql-operations'
import { CatalogPage } from '../../components/catalog-area-header/CatalogPage'
import { CatalogGroupIcon } from '../../components/CatalogGroupIcon'
import { catalogComponentIconComponent } from '../../components/ComponentIcon'

export const catalogPagePathForComponent = (
    component: { __typename: 'Component'; name: string; kind: ComponentKind; url: string } & ComponentOwnerLinkFields
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

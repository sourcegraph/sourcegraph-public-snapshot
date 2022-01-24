import { MenuPopover } from '@reach/menu-button'
import React, { useRef } from 'react'
import { Link } from 'react-router-dom'

import { gql } from '@sourcegraph/http-client'
import { MenuLink, Menu, MenuDivider, MenuHeader, MenuButton } from '@sourcegraph/wildcard'
import { MenuItems } from '@sourcegraph/wildcard/src/components/Menu/MenuItems'

import { ComponentTagFields } from '../../../../graphql-operations'
import { positionBottomRight } from '../../../insights/components/context-menu/utils'
import { CatalogComponentIcon } from '../ComponentIcon'

export const COMPONENT_TAG_FRAGMENT = gql`
    fragment ComponentTagFields on ComponentTag {
        name
        components {
            nodes {
                __typename
                id
                name
                kind
                url
            }
        }
    }
`

interface Props {
    tag: ComponentTagFields
    buttonClassName?: string
}

export const ComponentTag: React.FunctionComponent<Props> = ({ tag: { name, components }, buttonClassName }) => {
    const targetButtonReference = useRef<HTMLButtonElement>(null)
    return (
        <Menu>
            <MenuButton variant="link" className={buttonClassName} ref={targetButtonReference}>
                {name}
            </MenuButton>
            <MenuPopover position={positionBottomRight}>
                <MenuItems>
                    <MenuHeader>Tag: {name}</MenuHeader>
                    {components.nodes.slice(0, 15 /* TODO(sqs) */).map(component => (
                        <MenuLink
                            key={component.id}
                            as={Link}
                            to={component.url}
                            className="d-flex align-items-center overflow-hidden text-truncate"
                        >
                            <CatalogComponentIcon component={component} className="icon-inline mr-2" /> {component.name}
                        </MenuLink>
                    ))}
                    <MenuDivider />
                    <MenuLink as={Link} to={`/catalog?q=${encodeURIComponent(`tag:${name}`)}`}>
                        View as table...
                    </MenuLink>
                </MenuItems>
            </MenuPopover>
        </Menu>
    )
}

import { MenuItems } from '@reach/menu-button'
import React, { useRef } from 'react'
import { Link } from 'react-router-dom'

import { MenuLink, Menu, MenuDivider, MenuHeader, MenuButton } from '@sourcegraph/wildcard'

import { MenuList } from '@sourcegraph/wildcard/src/components/Menu'
import { ComponentTagsFields } from '../../../../graphql-operations'
import { positionBottomRight } from '../../../insights/components/context-menu/utils'
import { ComponentIcon } from '../../components/ComponentIcon'

import styles from './ComponentHeaderActions.module.scss'

interface Props {
    component: ComponentTagsFields
}

export const ComponentHeaderActions: React.FunctionComponent<Props> = ({ component: { tags } }) => (
    <nav className={styles.container}>
        {tags.map(tag => (
            <ComponentTag
                key={tag.name}
                name={tag.name}
                components={tag.components.nodes}
                buttonClassName="p-1 border small text-muted"
            />
        ))}
    </nav>
)

export const ComponentTag: React.FunctionComponent<{
    name: string
    components: ComponentTagsFields['tags'][0]['components']['nodes']
    buttonClassName?: string
}> = ({ name, components, buttonClassName }) => {
    const targetButtonReference = useRef<HTMLButtonElement>(null)
    return (
        <Menu>
            <MenuButton variant="link" className={buttonClassName} ref={targetButtonReference}>
                {name}
            </MenuButton>
            <MenuList position={positionBottomRight}>
                <MenuItems>
                    <MenuHeader>Tag: {name}</MenuHeader>
                    {components.slice(0, 15 /* TODO(sqs) */).map(component => (
                        <MenuLink
                            key={component.id}
                            as={Link}
                            to={component.url}
                            className="d-flex align-items-center overflow-hidden text-truncate"
                        >
                            <ComponentIcon component={component} className="icon-inline mr-2" /> {component.name}
                        </MenuLink>
                    ))}
                    <MenuDivider />
                    <MenuLink as={Link} to={`/catalog?q=${encodeURIComponent(`tag:${name}`)}`}>
                        View as table...
                    </MenuLink>
                </MenuItems>
            </MenuList>
        </Menu>
    )
}

import { MenuPopover } from '@reach/menu-button'
import classNames from 'classnames'
import CheckBoldIcon from 'mdi-react/CheckBoldIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import { Menu, MenuButton, MenuHeader, MenuLink, MenuDivider } from '@sourcegraph/wildcard'
import { MenuItems } from '@sourcegraph/wildcard/src/components/Menu/MenuItems'

import { ComponentTitleWithIconAndKind } from '../../../enterprise/catalog/contributions/tree/SourceSetTitle'
import { SourceSetAtTreeViewOptionsProps } from '../../../enterprise/catalog/pages/source-set-at-tree/useSourceSetAtTreeViewOptions'
import { SourceSetViewModeInfoResult } from '../../../graphql-operations'

import styles from './SourceSetViewModeAction.module.scss'

// TODO(sqs): LICENSE move to enterprise/

type ComponentFields = Extract<SourceSetViewModeInfoResult['node'], { __typename: 'Repository' }>['components'][number]

export const ComponentActionPopoverButton: React.FunctionComponent<
    {
        component: ComponentFields
        buttonClassName?: string
    } & Pick<SourceSetAtTreeViewOptionsProps, 'sourceSetAtTreeViewMode' | 'sourceSetAtTreeViewModeURL'>
> = ({ component, buttonClassName, sourceSetAtTreeViewMode, sourceSetAtTreeViewModeURL }) => (
    <Menu>
        <MenuButton
            variant="secondary"
            outline={true}
            className={classNames(
                'py-1 px-2',
                styles.btn,
                sourceSetAtTreeViewMode === 'auto' ? styles.btnViewModeComponent : styles.btnViewModeTree,
                buttonClassName
            )}
        >
            <ComponentTitleWithIconAndKind component={component} strong={sourceSetAtTreeViewMode === 'auto'} />
        </MenuButton>
        <SourceSetViewModeActionMenuItems
            sourceSetAtTreeViewMode={sourceSetAtTreeViewMode}
            sourceSetAtTreeViewModeURL={sourceSetAtTreeViewModeURL}
        />
    </Menu>
)

export const SourceSetViewModeActionMenuItems: React.FunctionComponent<
    Pick<SourceSetAtTreeViewOptionsProps, 'sourceSetAtTreeViewMode' | 'sourceSetAtTreeViewModeURL'>
> = ({ sourceSetAtTreeViewMode, sourceSetAtTreeViewModeURL }) => {
    const checkIcon = <CheckBoldIcon className="icon-inline" />
    const noCheckIcon = <CheckBoldIcon className="icon-inline invisible" />

    return (
        <MenuPopover>
            <MenuItems>
                <MenuHeader>View as...</MenuHeader>
                <MenuDivider />
                <MenuLink as={Link} to={sourceSetAtTreeViewModeURL.auto}>
                    {sourceSetAtTreeViewMode === 'auto' ? checkIcon : noCheckIcon} Component
                </MenuLink>
                <MenuLink as={Link} to={sourceSetAtTreeViewModeURL.tree}>
                    {sourceSetAtTreeViewMode === 'tree' ? checkIcon : noCheckIcon} Tree
                </MenuLink>
            </MenuItems>
        </MenuPopover>
    )
}

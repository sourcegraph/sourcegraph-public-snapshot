import { MenuList as ReachMenuList, MenuListProps as ReachMenuListProps } from '@reach/menu-button'
import classNames from 'classnames'
import React from 'react'

import { ForwardReferenceComponent } from '../../types'

export type MenuListProps = ReachMenuListProps

export const MenuList = React.forwardRef(({ children, className, ...props }, reference) => (
    <ReachMenuList ref={reference} {...props} className={classNames('dropdown-menu show', className)}>
        {children}
    </ReachMenuList>
)) as ForwardReferenceComponent<'div', MenuListProps>

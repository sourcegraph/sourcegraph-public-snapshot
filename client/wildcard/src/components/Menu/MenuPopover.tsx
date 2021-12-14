import { MenuPopover as ReachMenuPopover, MenuPopoverProps as ReachMenuPopoverProps } from '@reach/menu-button'
import React from 'react'

import { ForwardReferenceComponent } from '../../types'

export type MenuPopoverProps = ReachMenuPopoverProps

export const MenuPopover = React.forwardRef(({ children, className, ...props }, reference) => (
    <ReachMenuPopover ref={reference} {...props}>
        {children}
    </ReachMenuPopover>
)) as ForwardReferenceComponent<'div', MenuPopoverProps>

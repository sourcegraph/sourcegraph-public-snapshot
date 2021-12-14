import { MenuButton as ReachMenuButton } from '@reach/menu-button'
import React from 'react'

import { ForwardReferenceComponent } from '../../types'
import { Button, ButtonProps } from '../Button'

export type MenuButtonProps = Omit<ButtonProps, 'as'>

export const MenuButton = React.forwardRef(({ children, className, ...props }, reference) => (
    <ReachMenuButton ref={reference} as={Button} {...props}>
        {children}
    </ReachMenuButton>
)) as ForwardReferenceComponent<'button', MenuButtonProps>

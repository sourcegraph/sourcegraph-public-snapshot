import classNames from 'classnames'
import React from 'react'

import { ForwardReferenceComponent } from '../../types'

export const MenuDivider = React.forwardRef(({ children, className, ...props }, reference) => (
    <div ref={reference} {...props} className={classNames('dropdown-divider', className)} />
)) as ForwardReferenceComponent<'div'>

import React from 'react'

import classNames from 'classnames'

import { ForwardReferenceComponent } from '../../../types'
import { TypographyProps } from '../utils'

import { getLabelClassName } from './utils'

interface LabelProps extends React.HTMLAttributes<HTMLLabelElement>, TypographyProps {
    size?: 'small' | 'base'
    weight?: 'regular' | 'medium' | 'bold'
    isUnderline?: boolean
    isUppercase?: boolean
}

export const Label = React.forwardRef((props, reference) => {
    const {
        children,
        as: Component = 'label',
        size,
        weight,
        alignment,
        mode,
        isUnderline,
        isUppercase,
        className,
        ...rest
    } = props

    return (
        <Component
            ref={reference}
            className={classNames(
                getLabelClassName({ isUppercase, isUnderline, alignment, weight, size, mode }),
                className
            )}
            {...rest}
        >
            {children}
        </Component>
    )
}) as ForwardReferenceComponent<'label', LabelProps>

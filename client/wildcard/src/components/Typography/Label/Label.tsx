import React from 'react'

import classNames from 'classnames'

import { ForwardReferenceComponent } from '../../../types'
import { getAlignmentStyle, getFontWeightStyle, getModeStyle, TypographyProps } from '../utils'

import typographyStyles from '../Typography.module.scss'
import styles from './Label.module.scss'

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
                styles.label,
                isUnderline && styles.labelUnderline,
                isUppercase && styles.labelUppercase,
                size === 'small' && typographyStyles.small,
                weight && getFontWeightStyle({ weight }),
                alignment && getAlignmentStyle({ alignment }),
                mode && getModeStyle({ mode }),
                mode === 'single-line' && styles.labelSingleLine,
                className
            )}
            {...rest}
        >
            {children}
        </Component>
    )
}) as ForwardReferenceComponent<'label', LabelProps>

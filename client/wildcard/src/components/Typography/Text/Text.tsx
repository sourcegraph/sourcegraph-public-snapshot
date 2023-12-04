import React, { type ReactNode } from 'react'

import classNames from 'classnames'

import type { ForwardReferenceComponent } from '../../../types'
import type { TYPOGRAPHY_SIZES, TYPOGRAPHY_WEIGHTS } from '../constants'
import { getModeStyle, getAlignmentStyle, type TypographyProps, getFontWeightStyle } from '../utils'

import typographyStyles from '../Typography.module.scss'

interface TextProps extends React.HTMLAttributes<HTMLParagraphElement>, TypographyProps {
    size?: typeof TYPOGRAPHY_SIZES[number]
    weight?: typeof TYPOGRAPHY_WEIGHTS[number]
    children?: ReactNode | ReactNode[] | undefined
}

export const Text = React.forwardRef(function Text(
    { children, className, size, weight, as: Component = 'p', alignment, mode, ...props },
    reference
) {
    return (
        <Component
            className={classNames(
                size === 'small' && typographyStyles.small,
                weight && getFontWeightStyle({ weight }),
                alignment && getAlignmentStyle({ alignment }),
                mode && getModeStyle({ mode }),
                className
            )}
            ref={reference}
            {...props}
        >
            {children}
        </Component>
    )
}) as ForwardReferenceComponent<'p', TextProps>

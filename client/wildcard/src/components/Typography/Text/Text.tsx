import classNames from 'classnames'
import React from 'react'

import { ForwardReferenceComponent } from '../../../types'
import { TYPOGRAPHY_SIZES, TYPOGRAPHY_WEIGHTS } from '../constants'
import typographyStyles from '../Typography.module.scss'
import { getModeStyle, getAlignmentStyle, TypographyProps, getFontWeightStyle } from '../utils'

interface TextProps extends React.HTMLAttributes<HTMLParagraphElement>, TypographyProps {
    size?: typeof TYPOGRAPHY_SIZES[number]
    weight?: typeof TYPOGRAPHY_WEIGHTS[number]
}

export const Text = React.forwardRef(
    ({ children, className, size, weight, as: Component = 'p', alignment, mode }, reference) => (
        <Component
            className={classNames(
                size === 'small' && typographyStyles.small,
                weight && getFontWeightStyle({ weight }),
                alignment && getAlignmentStyle({ alignment }),
                mode && getModeStyle({ mode }),
                className
            )}
            ref={reference}
        >
            {children}
        </Component>
    )
) as ForwardReferenceComponent<'p', TextProps>

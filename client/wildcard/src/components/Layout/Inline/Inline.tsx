import classNames from 'classnames'
import React from 'react'

import { SpaceTokens, Box } from '../Box/Box'
import stackStyles from '../Stack/Stack.module.scss'

import inlineStyles from './Inline.module.scss'

interface Inline {
    children: React.ReactNode
    space: SpaceTokens
    component?: 'div' | 'ol' | 'ul'
    align?: 'left' | 'right' | 'center'
    alignY?: 'bottom' | 'top' | 'center'
    reverse?: boolean
}

export const Inline: React.FunctionComponent<Inline> = ({ children, space, component, align, alignY, reverse }) => (
    <Box className={classNames(stackStyles.stack, stackStyles[`stack${space}`])}>
        {/* Is this box needed? */}
        <Box
            component={component}
            flexWrap="wrap"
            flexDirection="row"
            className={classNames(alignY && inlineStyles[alignY])}
        >
            {React.Children.map(children, child => (
                <Box marginLeft={space} marginTop={space} minWidth={0}>
                    {child}
                </Box>
            ))}
        </Box>
    </Box>
)

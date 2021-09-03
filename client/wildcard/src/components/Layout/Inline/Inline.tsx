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
    <Box
        component={component}
        flexWrap="nowrap"
        flexDirection="row"
        gap={space}
        className={classNames(stackStyles[`stack${space}`], alignY && inlineStyles[alignY])}
    >
        {children}
    </Box>
)

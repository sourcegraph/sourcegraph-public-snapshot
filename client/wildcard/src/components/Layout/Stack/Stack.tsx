import classNames from 'classnames'
import React from 'react'

import { SpaceTokens, Box } from '../Box/Box'

import stackStyles from './Stack.module.scss'

interface Stack {
    children: React.ReactNode
    space: SpaceTokens
    component?: 'div' | 'ol' | 'ul'
    align: 'left' | 'right' | 'center'
    dividers: boolean
}

export const Stack: React.FunctionComponent<Stack> = ({ children, space, align, dividers }) => (
    <Box paddingBottom={space} className={classNames(stackStyles.stack, stackStyles[`stack${space}`])}>
        {React.Children.map(children, child => (
            <Box paddingTop={space}>{child}</Box>
        ))}
    </Box>
)

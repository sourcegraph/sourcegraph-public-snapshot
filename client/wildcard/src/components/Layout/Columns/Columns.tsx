import classNames from 'classnames'
import React from 'react'

import { Box } from '../Box/Box'

import columnStyle from './Columns.module.scss'

export const Columns: React.FunctionComponent = ({ children }) => <Box display="flex">{children}</Box>

interface Column {
    children: React.ReactNode
    width?: '25' | '50' | 'content'
}

export const Column: React.FunctionComponent<Column> = ({ children, width }) => (
    <Box
        minWidth={0}
        width={width !== 'content' ? 'full' : undefined}
        flexShrink={width === 'content' ? 0 : undefined}
        justifyContent="flexStart"
        className={classNames(columnStyle.columns, width !== 'content' ? columnStyle[`columns${width}`] : null)}
    >
        {children}
    </Box>
)
// <div className={columnStyle.columns}>{children}</div>

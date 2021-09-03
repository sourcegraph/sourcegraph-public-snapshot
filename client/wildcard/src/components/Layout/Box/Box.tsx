import React, { forwardRef } from 'react'

import boxStyles from './Box.module.scss'

export type SpaceTokens = 'gutter' | 'large' | 'medium' | 'none' | 'small' | 'xlarge' | 'xsmall' | 'xxlarge' | 'xxsmall'

interface Box {
    component?: any
    className?: string
    children: React.ReactNode
}

interface Styles {
    display: 'block' | 'flex' | 'inline' | 'inlineBlock' | 'none'
    position: 'fixed' | 'relative' | 'absolute'
    borderRadius: SpaceTokens
    paddingBottom: SpaceTokens
    paddingRight: SpaceTokens
    paddingLeft: SpaceTokens
    paddingTop: SpaceTokens
    marginBottom: SpaceTokens
    marginRight: SpaceTokens
    marginLeft: SpaceTokens
    marginTop: SpaceTokens
    gap: SpaceTokens
    alignItems: 'center' | 'flexEnd' | 'flexStart'
    justifyContent: 'center' | 'flexEnd' | 'flexStart' | 'spaceBetween'
    flexGrow: 0 | 1
    flexWrap: 'nowrap' | 'wrap'
    flexDirection: 'column' | 'columnReverse' | 'row' | 'rowReverse'
    flexShrink: 0
    textAlign: 'left' | 'right' | 'center'
    background: any
    overflow: 'hidden' | 'scroll' | 'visible' | 'auto'
    userSelect: 'none'
    outline: 'none'
    opacity: 0
    zIndex: 0 | 1 | 2 | 'dropdown' | 'dropdownBackdrop' | 'modal' | 'modalBackdrop' | 'notification' | 'sticky'
    boxShadow: any
    cursor: 'default' | 'pointer'
    pointerEvents: 'none'
    top: 0
    bottom: 0
    left: 0
    right: 0
    height: 'full' | 'touchable'
    width: 'full' | 'touchable'
    minWidth: 0
    maxWidth: 'large' | 'medium' | 'small' | 'xsmall'
}

export interface BoxProps extends Box, Partial<Styles> {}

export const getCSSfromProps = (styleProps): string =>
    Object.keys(styleProps)
        .map(name => boxStyles[`${name}${styleProps[name]}`])
        .join(' ')

export const Box = forwardRef<HTMLElement, BoxProps>(
    ({ component: Component = 'div', className, ...props }, reference) => {
        const { children, ...styles } = props
        // const classNamesFromProps = Object.keys(styles).map(name => boxStyles[`${name}${props[name]}`])

        return (
            <Component className={`${className || ''} ${getCSSfromProps(styles)}`} ref={reference}>
                {children}
            </Component>
        )
    }
)

Box.displayName = 'Box'

import React, { createContext, type ReactNode, useCallback, useContext, useEffect, useMemo, useState } from 'react'

import classNames from 'classnames'
import { noop } from 'lodash'
import FocusLock from 'react-focus-lock'

import type { ForwardReferenceComponent } from '../..'

import styles from './Collapse.module.scss'

type CollapseControlledProps =
    | { isOpen?: undefined; onOpenChange?: never; openByDefault?: boolean }
    | { isOpen: boolean; onOpenChange: (opened: boolean) => void; openByDefault?: boolean }

interface CollapseCommonProps {
    children:
        | React.FunctionComponent<React.PropsWithChildren<{ isOpen?: boolean; setOpen: (open: boolean) => void }>>
        | ReactNode
}

export type CollapseProps = CollapseControlledProps & CollapseCommonProps

interface CollapseContextData {
    isOpen?: boolean
    setOpen: (opened: boolean) => void
}
const DEFAULT_CONTEXT_VALUE: CollapseContextData = {
    isOpen: false,
    setOpen: noop,
}

const CollapseContext = createContext<CollapseContextData>(DEFAULT_CONTEXT_VALUE)

export const Collapse: React.FunctionComponent<CollapseProps> = React.memo(function Collapse(props) {
    const { children, isOpen, openByDefault, onOpenChange = noop } = props
    const [isInternalOpen, setInternalOpen] = useState<boolean>(Boolean(openByDefault))
    const isControlled = isOpen !== undefined
    const isCollapseOpen = isControlled ? isOpen : isInternalOpen
    const ChildrenComponent = typeof children === 'function' && children

    const setOpen = useCallback(
        (opened: boolean) => {
            if (!isControlled) {
                setInternalOpen(opened)
                return
            }

            onOpenChange(opened)
        },
        [isControlled, onOpenChange]
    )

    const context = useMemo(
        () => ({
            isOpen: isCollapseOpen,
            setOpen,
        }),
        [isCollapseOpen, setOpen]
    )

    return (
        <CollapseContext.Provider value={context}>
            {ChildrenComponent ? <ChildrenComponent isOpen={isCollapseOpen} setOpen={setOpen} /> : children}
        </CollapseContext.Provider>
    )
})

interface CollapseHeaderProps extends React.HTMLAttributes<HTMLButtonElement> {
    focusLocked?: boolean
}

export const CollapseHeader = React.forwardRef(function CollapseHeader(props, reference) {
    const { children, className, as: Component = 'button', onClick = noop, focusLocked, ...attributes } = props
    const { setOpen, isOpen } = useContext(CollapseContext)
    const [focusLock, setFocusLock] = useState(false)

    useEffect(() => {
        if (focusLocked) {
            requestAnimationFrame(() => {
                setFocusLock(true)
            })
        }

        return () => {
            setFocusLock(false)
        }
    }, [focusLocked])

    const handleClick: React.MouseEventHandler<HTMLButtonElement> = event => {
        setOpen(!isOpen)
        onClick(event)
    }

    const contentElement = (
        <Component
            className={className}
            ref={reference}
            onClick={handleClick}
            role="button"
            aria-expanded={isOpen}
            {...attributes}
        >
            {children}
        </Component>
    )

    if (!focusLocked) {
        return contentElement
    }

    return (
        <FocusLock disabled={focusLock} returnFocus={true}>
            {contentElement}
        </FocusLock>
    )
}) as ForwardReferenceComponent<'button', CollapseHeaderProps>

interface CollapsePanelProps {
    forcedRender?: boolean
}

export const CollapsePanel = React.forwardRef(function CollapsePanel(props, reference) {
    const { forcedRender = true, children, className, as: Component = 'div', ...attributes } = props
    const { isOpen } = useContext(CollapseContext)

    // When we enforce rendering we always render a DOM element and hide it with
    // display: none CSS rule, on other cases when we explicitly say forcedRender={false}
    // we render/not render DOM element based on isOpen state
    const shouldRender = forcedRender ? true : isOpen

    if (!shouldRender) {
        return null
    }

    return (
        <Component
            className={classNames(styles.collapse, isOpen && styles.show, className)}
            ref={reference}
            {...attributes}
        >
            {children}
        </Component>
    )
}) as ForwardReferenceComponent<'div', CollapsePanelProps>

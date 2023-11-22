import { forwardRef, type HTMLAttributes, useContext } from 'react'

import classNames from 'classnames'
import { noop } from 'lodash'
import { createPortal } from 'react-dom'
import { useMergeRefs } from 'use-callback-ref'

import { PopoverContext } from '../../contexts/internal-context'
import { PopoverRoot } from '../../contexts/public-context'

import style from './PopoverTail.module.scss'

enum PopoverSize {
    sm = 'sm',
    md = 'md',
    lg = 'lg',
}

const sizeClasses: Partial<Record<PopoverSize, string>> = {
    [PopoverSize.sm]: style.tailSizeSm,
    [PopoverSize.md]: style.tailSizeMd,
    [PopoverSize.lg]: style.tailSizeLg,
}

interface PopoverTailProps extends HTMLAttributes<HTMLElement> {
    size?: PopoverSize | `${PopoverSize}`
    forceRender?: boolean
}

export const PopoverTail = forwardRef<HTMLDivElement, PopoverTailProps>((props, ref) => {
    const { size = PopoverSize.md, forceRender = false, className, ...attributes } = props

    const { setTailElement, isOpen: isContextOpen } = useContext(PopoverContext)
    const { renderRoot } = useContext(PopoverRoot)

    const setContextTail = forceRender ? noop : setTailElement
    const tailRef = useMergeRefs<HTMLDivElement>([ref, setContextTail])

    const isOpen = forceRender ? true : isContextOpen

    if (!isOpen) {
        return null
    }

    return createPortal(
        <div ref={tailRef} className={classNames(style.tail, className, sizeClasses[size])} {...attributes}>
            <div className={style.tailInner} />
        </div>,
        renderRoot ?? document.body
    )
})

import { FunctionComponent, HTMLAttributes, useContext } from 'react'

import classNames from 'classnames'
import { createPortal } from 'react-dom'

import { PopoverContext } from '../../contexts/internal-context'
import { PopoverRoot } from '../../contexts/public-context'

import style from './PopoverTail.module.scss'

enum PopoverSize {
    sm = 'sm',
    md = 'md',
    lg = 'lg',
}

const sizeClasses: Record<PopoverSize, string> = {
    // sm is set by default (no styles are needed)
    [PopoverSize.sm]: '',
    [PopoverSize.md]: style.tailSizeMd,
    [PopoverSize.lg]: style.tailSizeLg,
}

interface PopoverTailProps extends HTMLAttributes<SVGElement> {
    size?: PopoverSize | `${PopoverSize}`
}

export const PopoverTail: FunctionComponent<PopoverTailProps> = props => {
    const { size = PopoverSize.sm } = props
    const { setTailElement, isOpen } = useContext(PopoverContext)
    const { renderRoot } = useContext(PopoverRoot)

    if (!isOpen) {
        return null
    }

    return createPortal(
        <div ref={setTailElement} className={classNames(style.tail, sizeClasses[size])}>
            <div className={style.tailInner} />
        </div>,
        renderRoot ?? document.body
    )
}

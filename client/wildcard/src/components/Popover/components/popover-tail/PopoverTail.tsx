import { FunctionComponent, HTMLAttributes, useContext } from 'react'

import { createPortal } from 'react-dom'

import { PopoverContext } from '../../contexts/internal-context'
import { PopoverRoot } from '../../contexts/public-context'

import style from './PopoverTail.module.scss'

interface PopoverTailProps extends HTMLAttributes<SVGElement> {}

export const PopoverTail: FunctionComponent<PopoverTailProps> = props => {
    const { setTailElement, isOpen } = useContext(PopoverContext)
    const { renderRoot } = useContext(PopoverRoot)

    if (!isOpen) {
        return null
    }

    return createPortal(
        <svg {...props} width="17.2" height="11" viewBox="0 0 200 130" className={style.tail} ref={setTailElement}>
            <path d="M0,0 L100,130 200,0" className={style.tailTrianglePath} />
        </svg>,
        renderRoot ?? document.body
    )
}

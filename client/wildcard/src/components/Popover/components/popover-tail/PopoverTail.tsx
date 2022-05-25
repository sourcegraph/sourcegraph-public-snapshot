import { FunctionComponent, HTMLAttributes, useContext } from 'react'

import { createPortal } from 'react-dom'

import { PopoverContext } from '../../context'

import style from './PopoverTail.module.scss'

interface PopoverTailProps extends HTMLAttributes<HTMLDivElement> {}

export const PopoverTail: FunctionComponent<PopoverTailProps> = props => {
    const { setTailElement } = useContext(PopoverContext)

    return createPortal(<div {...props} className={style.tail} ref={setTailElement} />, document.body)
}

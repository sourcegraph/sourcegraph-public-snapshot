import React from 'react'

import classNames from 'classnames'
import { isArray } from 'lodash'

import { type ForwardReferenceComponent } from '../..'

import styles from './KbdBadge.module.scss'

export interface Keybind {
    modifier: string
    selector: string[] | string
}

export interface KbdBadgeProps {
    shortCut: Keybind
    onClick?: () => void
    className?: string
}

/**
 * KbdBadge Element
 */
export const KbdBadge = React.forwardRef(function KbdBadge(
    { shortCut, className, onClick, as: Component = 'span', ...attributes },
    reference
) {
    const shortCutKeys = createKbd(shortCut)

    return (
        <Component className={classNames(className)} ref={reference} {...attributes}>
            <div className={classNames(styles.keybind, onClick && styles.withCursor)} onClick={onClick}>
                {shortCutKeys[0]} {shortCutKeys[1]}
            </div>
        </Component >
    )
}) as ForwardReferenceComponent<'span', KbdBadgeProps>

const createKbd = (kb: Keybind): string[] => {
    let shortcut = [kb.modifier]
    let keys = ''
    if (isArray(kb.selector)) {
        for (const p of kb.selector) {
            keys += ' + ' + p.toUpperCase()
        }
    } else {
        keys = kb.selector
    }

    shortcut.push(keys)
    return shortcut
}

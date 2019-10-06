import React from 'react'
import * as GQL from '../../../../shared/src/graphql/schema'

export interface SideEffectListItemContext {}

interface Props extends SideEffectListItemContext {
    sideEffect: GQL.ISideEffect

    className?: string
}

/**
 * An item in a list of side effects.
 */
export const SideEffectListItem: React.FunctionComponent<Props> = ({ sideEffect, className = '' }) => (
    <li className={`list-group-item d-flex align-items-center ${className}`}>
        {sideEffect.title}
        {sideEffect.detail && <span className="ml-2 mt-1 text-muted small">{sideEffect.detail}</span>}
    </li>
)

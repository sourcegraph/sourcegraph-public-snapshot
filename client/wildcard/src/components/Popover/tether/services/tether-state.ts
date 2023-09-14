import { getPositions } from './geometry'
import { type TetherState, getPositionState } from './tether-position'
import type { TetherLayout } from './types'

/**
 * Calculates and returns the most suitable position and tooltip element
 * size value based on layout settings.
 *
 * @param layout
 */
export function getState(layout: TetherLayout): TetherState | null {
    let state: TetherState | null = null

    for (const position of getPositions(layout.position, layout.flipping)) {
        const current = getPositionState(layout, position)
        state ??= current

        if (current === null) {
            continue
        }

        if (state !== null && current.elementArea > state.elementArea) {
            state = current
        }

        // If element bounds equals to null that means that non of constraints were applied
        // therefore we have enough space to render all content inside tooltip without any
        // visual restriction. Pick current state and stop further calculations
        if (current.elementBounds === null) {
            break
        }
    }

    return state
}

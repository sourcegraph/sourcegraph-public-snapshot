import { Facet, StateEffect, StateEffectType, StateField, Text } from '@codemirror/state'

import { Position, Range } from '@sourcegraph/extension-api-types'

/**
 * This is a trailing throttle implementation that allows the callback to query
 * whether it was canceled or not (i.e. whether or not the function was called
 * again). This can be useful if the callback performs asynchronous work.
 */
export function throttle<P extends unknown[]>(
    callback: ({ isCanceled }: { isCanceled: boolean }, ...args: P) => void,
    timeout: number
): (...args: P) => void {
    let running = false
    let lastTimeCalled = 0
    let lastArguments: P

    return (...args) => {
        lastArguments = args

        if (!running) {
            running = true
            const timeCalled = (lastTimeCalled = Date.now())
            setTimeout(() => {
                running = false
                callback(
                    {
                        get isCanceled() {
                            return timeCalled !== lastTimeCalled
                        },
                    },
                    ...lastArguments
                )
            }, timeout)
        }
    }
}

/**
 * Converts line/character positions to document offsets.
 */
export function positionToOffset(textDocument: Text, position: Position): number {
    // Position is 0-based
    return textDocument.line(position.line + 1).from + position.character
}

/**
 * Helper function create an effect and a field to act as input provider for a
 * facet.
 */
export function createFacetInputField<T>(facet: Facet<T, unknown>, initial: T): [StateField<T>, StateEffectType<T>] {
    const setFieldValue = StateEffect.define<T>()
    const field = StateField.define<T>({
        create: () => initial,
        update(value, transaction) {
            for (const effect of transaction.effects) {
                if (effect.is(setFieldValue)) {
                    return effect.value
                }
            }
            return value
        },
        provide: field => facet.from(field),
    })

    return [field, setFieldValue]
}

export function viewPortChanged(
    previous: { from: number; to: number } | null,
    next: { from: number; to: number }
): boolean {
    return previous?.from !== next.from || previous.to !== next.to
}

export function sortRangeValuesByStart<T extends { range: Range }>(values: T[]): T[] {
    return values.sort(({ range: rangeA }, { range: rangeB }) =>
        rangeA.start.line === rangeB.start.line
            ? rangeA.start.character - rangeB.start.character
            : rangeA.start.line - rangeB.start.line
    )
}

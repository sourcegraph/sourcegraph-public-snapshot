import { getTicks } from '@visx/scale'
import type { AnyD3Scale } from '@visx/scale/lib/types/Scale'
import { format } from 'd3-format'

const SI_PREFIX_FORMATTER = format('~s')

export function formatYTick(number: number): string {
    // D3 formatter doesn't support float numbers properly
    if (!Number.isInteger(number)) {
        return number.toString()
    }

    return SI_PREFIX_FORMATTER(number)
}

const MINIMUM_NUMBER_OF_TICKS = 2

export interface GetScaleTicksOptions {
    scale: AnyD3Scale
    space: number
    pixelsPerTick?: number
}

export function getXScaleTicks<T>(input: GetScaleTicksOptions): T[] {
    const { scale, space, pixelsPerTick = 80 } = input

    // Calculate desirable number of ticks
    const numberTicks = Math.max(MINIMUM_NUMBER_OF_TICKS, Math.floor(space / pixelsPerTick))

    let filteredTicks = getTicks(scale)

    while (filteredTicks.length > numberTicks) {
        filteredTicks = getHalvedTicks(filteredTicks)
    }

    return filteredTicks
}

/**
 * Returns list of not formatted (raw) Y axis ticks.
 * Example: 1000, 1500, 2000, ...
 *
 * Number of lines (ticks) is based on chart height value and our expectation
 * around label density on the chart (no more than 1 tick in each 40px, see
 * HEIGHT_PER_TICK const)
 *
 * Ticks are constrained to integers.
 */

export function getYScaleTicks(input: GetScaleTicksOptions): number[] {
    const { scale, space, pixelsPerTick = 40 } = input

    // Generate max density ticks (d3 scale generation)
    const ticks: number[] = getTicks(scale).filter(Number.isInteger) as number[]

    if (ticks.length <= 2) {
        return ticks
    }

    // Calculate desirable number of ticks
    const numberTicks = Math.max(MINIMUM_NUMBER_OF_TICKS, Math.floor(space / pixelsPerTick))

    let filteredTicks = ticks

    while (filteredTicks.length > numberTicks) {
        filteredTicks = getHalvedTicks(filteredTicks)
    }

    return filteredTicks
}

/**
 * Cut off half of tick elements from the list based on
 * original number of ticks. With odd number of original ticks
 * removes all even index ticks with even number removes all
 * odd index ticks.
 */
function getHalvedTicks<T>(ticks: T[]): T[] {
    const isOriginTickLengthOdd = !(ticks.length % 2)
    const filteredTicks = []

    for (let index = ticks.length; index >= 1; index--) {
        if (isOriginTickLengthOdd) {
            if (index % 2 === 0) {
                filteredTicks.unshift(ticks[index - 1])
            }
        } else if (index % 2) {
            filteredTicks.unshift(ticks[index - 1])
        }
    }

    return filteredTicks
}

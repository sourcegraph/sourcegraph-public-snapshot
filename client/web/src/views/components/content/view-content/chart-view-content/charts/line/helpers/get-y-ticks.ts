import { getTicks } from '@visx/scale'
import { AnyD3Scale } from '@visx/scale/lib/types/Scale'

const HEIGHT_PER_TICK = 40

/**
 * Returns list of not formatted (raw) Y axis ticks.
 * Example: 1000, 1500, 2000, ...
 *
 * Number of lines (ticks) is based on chart height value and our expectation
 * around label density on the chart (no more than 1 tick in each 40px, see
 * HEIGHT_PER_TICK const)
 */
export function getYTicks(scale: AnyD3Scale, height: number): number[] {
    // Generate max density ticks (d3 scale generation)
    const ticks: number[] = getTicks(scale)

    if (ticks.length <= 2) {
        return ticks
    }

    // Calculate desirable number of ticks
    const numberTicks = Math.max(1, Math.floor(height / HEIGHT_PER_TICK))

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
function getHalvedTicks(ticks: number[]): number[] {
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

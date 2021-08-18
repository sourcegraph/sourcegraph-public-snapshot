import { getTicks } from '@visx/scale'
import { AnyD3Scale } from '@visx/scale/lib/types/Scale'

import { numberFormatter } from '../components/TickComponent'

const APPROXIMATE_SYMBOL_WIDTH = 11

export function getYAxisWidth<Scale extends AnyD3Scale>(scale: Scale, numberTicks: number): number {
    const ticksValues = getTicks(scale, numberTicks)
    const ticksLengths = ticksValues
        .map(numberFormatter)
        .map(value =>
            value
                .split('')
                // Filter all dots from the label symbols to avoid unnecessary
                // width increasing (dots take just a few pixels)
                .filter(symbol => symbol !== '.')
        )
        .map(value => value.length)
    const maxNumberSymbolsInTicks = Math.max(...ticksLengths)

    return maxNumberSymbolsInTicks * APPROXIMATE_SYMBOL_WIDTH
}

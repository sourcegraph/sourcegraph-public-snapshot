import { getTicks } from '@visx/scale'
import { AnyD3Scale } from '@visx/scale/lib/types/Scale'

import { numberFormatter } from '../components/TickComponent'

const APPROXIMATE_SYMBOL_WIDTH = 11
const MINIMAL_NUMBER_OF_LABEL_SYMBOLS = 2

export function getYAxisWidth<Scale extends AnyD3Scale>(scale: Scale, numberTicks: number): number {
    const ticksValues = getTicks(scale, numberTicks)
    const ticksLengths = ticksValues.map(
        value =>
            numberFormatter(value)
                .split('')
                // Filter all dots from the label symbols to avoid unnecessary
                // width increasing (dots take just a few pixels)
                .filter(symbol => symbol !== '.').length
    )

    const maxNumberSymbolsInTicks = Math.max(...ticksLengths, MINIMAL_NUMBER_OF_LABEL_SYMBOLS)

    return maxNumberSymbolsInTicks * APPROXIMATE_SYMBOL_WIDTH
}

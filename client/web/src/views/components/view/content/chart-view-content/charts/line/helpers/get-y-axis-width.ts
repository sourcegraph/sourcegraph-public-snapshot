import { numberFormatter } from '../components/TickComponent'

const APPROXIMATE_SYMBOL_WIDTH = 11
const MINIMAL_NUMBER_OF_LABEL_SYMBOLS = 3

/**
 * Returns width of Y label chart part based on max number of character
 * in a longest Y axis label.
 *
 * @param ticksValues - generated Y labels (ticks) strings - 0.1, 12, 42.2k, 2M
 */
export function getYAxisWidth(ticksValues: number[]): number {
    const ticksLengths = ticksValues.map(value => numberFormatter(value).split('').length)

    const maxNumberSymbolsInTicks = Math.max(...ticksLengths, MINIMAL_NUMBER_OF_LABEL_SYMBOLS)

    return maxNumberSymbolsInTicks * APPROXIMATE_SYMBOL_WIDTH
}

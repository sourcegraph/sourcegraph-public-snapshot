/**
 * # LED numerals exercise
 *
 * Write a command line program which takes a number and displays the “LED” version of
 * this number using underscores and vertical bars:
 *
 * $ led 757
 *  _   _  _
 *   | |_   |
 *   |  _|  |
 *
 * The first goal is just to get the above working. This likely will not take all of the
 * time. Use the stretch goals below to fill the time:
 *
 *     - Add a second parameter for bar length:
 *         $ led 757 2
 *          __   __  __
 *            | |      |
 *            | |__    |
 *            |    |   |
 *            |  __|   |
 *     - How do we test this for correctness? Write the tests.
 *     - How do we test this for performance? Write the tests.
 */

export const writeLED = (led: number): void => {
    let firstRow = ''
    let secondRow = ''
    let thirdRow = ''

    // Split the LED number into digits
    const digits = led.toString().split('')
    // Get the LED pattern for each digit and build the rows with it
    for (const digit of digits) {
        firstRow += getTopLEDPattern(parseInt(digit, 10)) + ' '
        secondRow += getMiddleLEDPattern(parseInt(digit, 10)) + ' '
        thirdRow += getBottomLEDPattern(parseInt(digit, 10)) + ' '
    }

    console.log(`${firstRow}\n${secondRow}\n${thirdRow}`)
}

// TODO: Handle errors if digit > 9
const getTopLEDPattern = (digit: number): string => {
    switch (digit) {
        case 0:
        case 2:
        case 3:
        case 5:
        case 6:
        case 7:
        case 8:
        case 9:
            return ' _ '
        case 1:
        case 4:
        default:
            return '   '
    }
}

const getMiddleLEDPattern = (digit: number): string => {
    switch (digit) {
        case 0:
            return '| |'
        case 1:
            return ' | '
        case 2:
        case 3:
            return ' _|'
        case 4:
        case 8:
        case 9:
            return '|_|'
        case 5:
        case 6:
            return '|_ '
        case 7:
        default:
            return '  |'
    }
}

const getBottomLEDPattern = (digit: number): string => {
    switch (digit) {
        case 0:
        case 6:
        case 8:
            return '|_|'
        case 1:
            return ' | '
        case 2:
            return '|_ '
        case 3:
        case 5:
        case 9:
            return ' _|'
        case 4:
        case 7:
        default:
            return '  |'
    }
}

// writeLED(1234567890)
// writeLED(10000)

const writeLED2 = (led: number, barLength = 1): void => {
    let topRow = ''
    let topVerticalsRow = ''
    let middleRow = ''
    let bottomVerticalsRow = ''
    let bottomRow = ''

    // Split the LED number into digits
    const digits = led.toString().split('')
    // Get the LED pattern for each digit and build the rows with it
    for (const digit of digits) {
        topRow += getTopLEDPattern2(parseInt(digit, 10), barLength) + ' '
        topVerticalsRow += getTopVerticalsRowLEDPattern(parseInt(digit, 10), barLength) + ' '
        middleRow += getMiddleLEDPattern2(parseInt(digit, 10), barLength) + ' '
        bottomVerticalsRow += getBottomVerticalsRowLEDPattern(parseInt(digit, 10), barLength) + ' '
        bottomRow += getBottomLEDPattern2(parseInt(digit, 10), barLength) + ' '
    }

    const topVerticals = new Array(barLength - 1).fill(topVerticalsRow).join('\n')
    const bottomVerticals = new Array(barLength - 1).fill(bottomVerticalsRow).join('\n')

    console.log(
        `${topRow}\n${topVerticals}${barLength > 1 ? '\n' : ''}${middleRow}\n${bottomVerticals}${
            barLength > 1 ? '\n' : ''
        }${bottomRow}`
    )
}

// TODO: Handle errors if digit > 9
const getTopLEDPattern2 = (digit: number, barLength: number): string => {
    switch (digit) {
        case 0:
        case 2:
        case 3:
        case 5:
        case 6:
        case 7:
        case 8:
        case 9:
            return ' ' + '_'.repeat(barLength) + ' '
        case 1:
        case 4:
        default:
            return ' ' + ' '.repeat(barLength) + ' '
    }
}

const getTopVerticalsRowLEDPattern = (digit: number, barLength: number): string => {
    switch (digit) {
        case 0:
        case 4:
        case 8:
        case 9:
            return '|' + ' '.repeat(barLength) + '|'
        case 1:
        case 5:
        case 6:
            return '|' + ' '.repeat(barLength) + ' '
        case 2:
        case 3:
        case 7:
        default:
            return ' ' + ' '.repeat(barLength) + '|'
    }
}

const getMiddleLEDPattern2 = (digit: number, barLength: number): string => {
    switch (digit) {
        case 0:
            return '|' + ' '.repeat(barLength) + '|'
        case 1:
            return '|' + ' '.repeat(barLength) + ' '
        case 2:
        case 3:
            return ' ' + '_'.repeat(barLength) + '|'
        case 4:
        case 8:
        case 9:
            return '|' + '_'.repeat(barLength) + '|'
        case 5:
        case 6:
            return '|' + '_'.repeat(barLength) + ' '
        case 7:
        default:
            return ' ' + ' '.repeat(barLength) + '|'
    }
}

const getBottomVerticalsRowLEDPattern = (digit: number, barLength: number): string => {
    switch (digit) {
        case 0:
        case 6:
        case 8:
            return '|' + ' '.repeat(barLength) + '|'
        case 1:
        case 2:
            return '|' + ' '.repeat(barLength) + ' '
        case 3:
        case 4:
        case 5:
        case 7:
        case 9:
        default:
            return ' ' + ' '.repeat(barLength) + '|'
    }
}

const getBottomLEDPattern2 = (digit: number, barLength: number): string => {
    switch (digit) {
        case 0:
        case 6:
        case 8:
            return '|' + '_'.repeat(barLength) + '|'
        case 1:
            return '|' + ' '.repeat(barLength) + ' '
        case 2:
            return '|' + '_'.repeat(barLength) + ' '
        case 3:
        case 5:
        case 9:
            return ' ' + '_'.repeat(barLength) + '|'
        case 4:
        case 7:
        default:
            return ' ' + ' '.repeat(barLength) + '|'
    }
}

// writeLED2(1234567890, 3)
writeLED2(parseInt(process.argv[2], 10), parseInt(process.argv[3], 10) || 1)

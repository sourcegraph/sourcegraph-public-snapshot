import chalk from 'chalk'

export function printSuccessBanner(lines: string[], log = console.log.bind(console)): void {
    const lineLength = getLineLength(lines)
    const banner = '='.repeat(lineLength)
    const emptyLine = ' '.repeat(lineLength)

    log(chalk.bgGreenBright.black(banner))
    log(chalk.bgGreenBright.black(emptyLine))
    for (const line of lines) {
        log(chalk.bgGreenBright.black(padLine(line, lineLength)))
        log(chalk.bgGreenBright.black(emptyLine))
    }
    log(chalk.bgGreenBright.black(banner))
}

function padLine(content: string, length: number): string {
    const spaceRequired = length - content.length
    const half = spaceRequired / 2
    let line = `${' '.repeat(half)}${content}${' '.repeat(half)}`
    if (line.length < length) {
        line += ' '
    }
    return line
}

function getLineLength(lines: string[]): number {
    const longestLine = lines.reduce((accumulator, current) => {
        if (accumulator < current.length) {
            return current.length
        }
        return accumulator
    }, 0)

    return Math.max(40, longestLine + 10)
}

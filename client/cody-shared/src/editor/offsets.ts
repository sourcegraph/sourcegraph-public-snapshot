import { Position, Range } from '.'

/**
 * Utility class to convert line/character positions into offsets.
 */
export class DocumentOffsets {
    public lines: number[] = []

    constructor(public readonly content: string) {
        if (content) {
            this.lines.push(0)
            let index = 1
            while (index < content.length) {
                if (content[index] === '\n') {
                    this.lines.push(index + 1)
                }
                index++
            }
            if (content.length !== this.lines[this.lines.length - 1]) {
                this.lines.push(content.length) // sentinel value
            }
        }
    }

    public offset(position: Position): number {
        const lineStartOffset = this.lines[position.line]
        return lineStartOffset + position.character
    }

    public rangeSlice(range: Range): string {
        const start = this.offset(range.start)
        const end = this.offset(range.end)

        return this.content.slice(start, end)
    }
}

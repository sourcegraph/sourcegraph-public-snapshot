import { Position, TextDocument } from './protocol'

/**
 * Utility class to convert line/character positions into offsets.
 */
export class DocumentOffsets {
    private lines: number[] = []
    constructor(public readonly document: TextDocument) {
        if (document.content) {
            this.lines.push(0)
            let index = 1
            while (index < document.content.length) {
                if (document.content[index] === '\n') {
                    this.lines.push(index + 1)
                }
                index++
            }
            if (document.content.length !== this.lines[this.lines.length - 1]) {
                this.lines.push(document.content.length) // sentinel value
            }
        }
    }
    public offset(position: Position): number {
        const lineStartOffset = this.lines[position.line]
        return lineStartOffset + position.character
    }
}

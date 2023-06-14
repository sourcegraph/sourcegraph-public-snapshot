import { Position, TextDocument } from './protocol'

export class Offsets {
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
                this.lines.push(document.content.length) // sentinel value used for binary search
            }
        }
    }
    public offset(position: Position): number {
        const lineStartOffset = this.lines[position.line]
        return lineStartOffset + position.character
    }
}

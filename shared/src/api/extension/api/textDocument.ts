import * as sourcegraph from 'sourcegraph'
import { PrefixSumComputer } from '../../../util/prefixSumComputer'
import { getWordAtText } from '../../../util/wordHelpers'
import { Position } from '../types/position'
import { Range } from '../types/range'

/** @internal */
export class ExtDocument implements sourcegraph.TextDocument {
    private _eol: string
    private _lines: string[]

    constructor(private model: Pick<sourcegraph.TextDocument, 'uri' | 'languageId' | 'text'>) {
        this._eol = getEOL(model.text || '')
        this._lines = model.text !== undefined ? model.text.split(this._eol) : []
    }

    public readonly uri = this.model.uri
    public readonly languageId = this.model.languageId
    public readonly text = this.model.text

    public offsetAt(position: sourcegraph.Position): number {
        this.throwIfNoModelText()
        position = this.validatePosition(position)
        return this.lineStarts.getAccumulatedValue(position.line - 1) + position.character
    }

    public positionAt(offset: number): sourcegraph.Position {
        this.throwIfNoModelText()
        offset = Math.floor(offset)
        offset = Math.max(0, offset)

        const out = this.lineStarts.getIndexOf(offset)
        const lineLength = this._lines[out.index].length
        const character = Math.min(out.remainder, lineLength) // ensure we return a valid position
        return new Position(out.index, character)
    }

    public validatePosition(position: sourcegraph.Position): sourcegraph.Position {
        this.throwIfNoModelText()
        if (!(position instanceof Position)) {
            throw new TypeError('invalid argument')
        }

        let { line, character } = position
        let hasChanged = false

        if (line < 0) {
            line = 0
            character = 0
            hasChanged = true
        } else if (line >= this._lines.length) {
            line = this._lines.length - 1
            character = this._lines[line].length
            hasChanged = true
        } else {
            const maxCharacter = this._lines[line].length
            if (character < 0) {
                character = 0
                hasChanged = true
            } else if (character > maxCharacter) {
                character = maxCharacter
                hasChanged = true
            }
        }

        if (!hasChanged) {
            return position
        }
        return new Position(line, character)
    }

    public validateRange(range: sourcegraph.Range): sourcegraph.Range {
        this.throwIfNoModelText()
        if (!(range instanceof Range)) {
            throw new TypeError('invalid argument')
        }

        const start = this.validatePosition(range.start)
        const end = this.validatePosition(range.end)

        if (start === range.start && end === range.end) {
            return range
        }
        return new Range(start.line, start.character, end.line, end.character)
    }

    public getWordRangeAtPosition(position: sourcegraph.Position): sourcegraph.Range | undefined {
        this.throwIfNoModelText()
        position = this.validatePosition(position)
        const wordAtText = getWordAtText(position.character, this._lines[position.line])
        if (wordAtText) {
            return new Range(position.line, wordAtText.startColumn, position.line, wordAtText.endColumn)
        }
        return undefined
    }

    // Memoize computation of line starts.
    private _lineStarts: PrefixSumComputer | null = null
    private get lineStarts(): PrefixSumComputer {
        if (!this._lineStarts) {
            const eolLength = this._eol.length
            const linesLength = this._lines.length
            const lineStartValues = new Uint32Array(linesLength)
            for (let i = 0; i < linesLength; i++) {
                lineStartValues[i] = this._lines[i].length + eolLength
            }
            this._lineStarts = new PrefixSumComputer(lineStartValues)
        }
        return this._lineStarts
    }

    private throwIfNoModelText(): void {
        if (this.model.text === undefined) {
            throw new Error('model text is not available')
        }
    }

    public toJSON(): any {
        return this.model
    }
}

/**
 * Detects the end-of-line character in the text (either \n, \r\n, or \r).
 */
export function getEOL(text: string): string {
    for (let i = 0; i < text.length; i++) {
        const ch = text.charAt(i)
        if (ch === '\r') {
            if (i + 1 < text.length && text.charAt(i + 1) === '\n') {
                return '\r\n'
            }
            return '\r'
        }
        if (ch === '\n') {
            return '\n'
        }
    }
    return '\n'
}

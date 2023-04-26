import assert from 'assert'

import { updateRange } from './tracked-range'

// A position, a subset of vscode.Position, for testing tracked ranges.
class SimplePosition {
    constructor(public line: number, public character: number) {}
    public isAfter(other: SimplePosition): boolean {
        return other.line < this.line || (other.line === this.line && other.character < this.character)
    }
    public isAfterOrEqual(other: SimplePosition): boolean {
        return this.isAfter(other) || this.isEqual(other)
    }
    public isBefore(other: SimplePosition): boolean {
        return !this.isAfterOrEqual(other)
    }
    public isBeforeOrEqual(other: SimplePosition): boolean {
        return !this.isAfter(other)
    }
    public isEqual(other: SimplePosition): boolean {
        return this.line === other.line && this.character === other.character
    }
    public translate(lineDelta?: number, characterDelta?: number): SimplePosition {
        return new SimplePosition(this.line + (lineDelta || 0), this.character + (characterDelta || 0))
    }
}

// A range, a subset of vscode.Range, for testing tracked ranges.
class SimpleRange {
    constructor(public start: SimplePosition, public end: SimplePosition) {}
    public with(start: SimplePosition, end: SimplePosition): SimpleRange {
        return start.isEqual(this.start) && end.isEqual(this.end) ? this : new SimpleRange(start, end)
    }
}

// Creates a position.
function pos(line: number, character: number): SimplePosition {
    return new SimplePosition(line, character)
}

// Creates a range.
function rng(start: SimplePosition, end: SimplePosition): SimpleRange {
    return new SimpleRange(start, end)
}

// Given some text and a range, serializes the range. Range is denoted with []s.
// For example, "Hello, [world]!"
function show(text: string, range: SimpleRange): string {
    const buffer = []
    let line = 0
    let beginningOfLine = 0
    for (let i = 0; i <= text.length; i++) {
        const position = new SimplePosition(line, i - beginningOfLine)
        if (position.isEqual(range.start)) {
            buffer.push('[')
        }
        if (position.isEqual(range.end)) {
            buffer.push(']')
        }
        if (i < text.length) {
            const ch = text[i]
            buffer.push(ch)
            if (ch === '\n') {
                line++
                beginningOfLine = i + 1
            }
        }
    }
    return buffer.join('')
}

// Given a spec with a couple of ranges specified using [] and (),
// returns the text and the ranges. The "tracked" range uses [], the
// "edited" range uses ().
//
// For example:
// "[He(llo,] world)!" produces text "Hello, world!" with the "tracked"
// range encompassing "Hello," and the "edited" encompassing "llo, world"
//
// Does not actually track or edit those ranges--just uses this naming
// convention because all of the tests *do* track and edit ranges.
function parse(spec: string): { tracked: SimpleRange; edited: SimpleRange; text: string } {
    const buffer = []
    let trackedStart
    let trackedEnd
    let editedStart
    let editedEnd
    let line = 0
    let beginningOfLine = 0
    let i = 0
    for (const ch of spec) {
        const here = pos(line, i - beginningOfLine)
        switch (ch) {
            case '[':
                assert(!trackedStart, 'multiple starting range [s')
                trackedStart = here
                break
            case ']':
                assert(trackedStart, 'missing start [')
                assert(!trackedEnd, 'multiple ending ]s')
                trackedEnd = here
                break
            case '(':
                assert(!editedStart, 'multiple starting range (s')
                editedStart = here
                break
            case ')':
                assert(editedStart, 'missing start (')
                assert(!editedEnd, 'multiple ending )s')
                editedEnd = here
                break
            case '\n':
                line++
                beginningOfLine = i + 1
            // fallthrough
            default:
                i++
                buffer.push(ch)
        }
    }

    assert(trackedStart && trackedEnd && editedStart && editedEnd, 'ranges should be specified with [], ()')

    return {
        tracked: rng(trackedStart, trackedEnd),
        edited: rng(editedStart, editedEnd),
        text: buffer.join(''),
    }
}

// Replaces a range with the specified text.
function edit(text: string, range: SimpleRange, replacement: string): string {
    const buffer = []
    let line = 0
    let beginningOfLine = 0
    for (let i = 0; i < text.length; i++) {
        const here = pos(line, i - beginningOfLine)
        const ch = text[i]
        if (here.isEqual(range.start)) {
            buffer.push(replacement)
        } else if (here.isBefore(range.start) || here.isAfterOrEqual(range.end)) {
            buffer.push(ch)
        }
        if (ch === '\n') {
            line++
            beginningOfLine = i + 1
        }
    }
    return buffer.join('')
}

// Given a spec with a tracked range in [], an edited range in (),
// replaces () with the specified text; applies range tracking;
// and serializes the resulting text and tracked range.
function track(spec: string, replacement: string): string {
    const scenario = parse(spec)
    const editedText = edit(scenario.text, scenario.edited, replacement)
    const updatedRange = updateRange(scenario.tracked, { range: scenario.edited, text: replacement })
    return updatedRange ? show(editedText, updatedRange) : editedText
}

describe('Tracked Range', () => {
    it('show helper function displays ranges', () => {
        expect(show('hello\nworld', rng(pos(0, 2), pos(1, 3)))).toBe('he[llo\nwor]ld')
    })
    it('parse helper function extracts ranges', () => {
        expect(parse('he[ll(o],\nw)orld')).toStrictEqual({
            tracked: rng(pos(0, 2), pos(0, 5)),
            edited: rng(pos(0, 4), pos(1, 1)),
            text: 'hello,\nworld',
        })
    })
    it('edit helper function accurately edits text', () => {
        const scenario = parse('(hello,\n) world![]')
        expect(edit(scenario.text, scenario.edited, 'goodbye')).toBe('goodbye world!')
    })
    it('Handles single-character deletion before the range', () => {
        expect(track('he(l)lo\nw[or]ld', '')).toBe('helo\nw[or]ld')
    })
})

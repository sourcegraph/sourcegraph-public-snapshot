/*

// Debugging helpers

function time<T>(label: string, f: () => T): T {
    const start = performance.now()
    const result = f()
    const end = performance.now()
    console.log(label, end - start, 'msec')
    return result
}

function dumpProgram(program: Uint16Array, ops: Op[], a: string, b: string): void {
    const buffer = []
    buffer.push(`  ^${Array.prototype.map.call(a, render).join('')}\n`)
    for (let iB = 0; iB <= b.length; iB++) {
        buffer.push(`${iB === 0 ? '^' : render(b[iB - 1])}|`)
        for (let iA = 0; iA <= a.length; iA++) {
            buffer.push(program[iB * (a.length + 1) + iA].toString(26).slice(0, 1))
        }
        buffer.push('\n')
    }
    buffer.push(`  ^${Array.prototype.map.call(a, render).join('')}\n`)
    for (let iB = 0; iB <= b.length; iB++) {
        buffer.push(`${iB === 0 ? '^' : render(b[iB - 1])}|`)
        for (let iA = 0; iA <= a.length; iA++) {
            buffer.push(ops[iB * (a.length + 1) + iA])
        }
        buffer.push('\n')
    }
    console.log(buffer.join(''))
}
*/

function render(ch: string): string {
    return ch === '\n' ? '@' : ch
}

export function dumpUse(use: Uint8Array, a: string, b: string): void {
    const buffer = []
    buffer.push(`  ^${Array.prototype.map.call(a, render).join('')}\n`)
    for (let iB = 0; iB <= b.length; iB++) {
        buffer.push(`${iB === 0 ? '^' : render(b[iB - 1])}|`)
        for (let iA = 0; iA <= a.length; iA++) {
            buffer.push(use[iB * (a.length + 1) + iA].toString())
        }
        buffer.push('\n')
    }
    console.log(buffer.join(''))
}

// * the root of the program at the start of the strings
// I insert from B
// X delete from A
// - accept a character the same in A and B
// R replace a character in A with a different one in B
// TODO: Use a cheaper representation than strings for the operations.
type Op = '*' | 'I' | 'X' | '-' | 'R'

// Computes the longest common subsequence of strings a and b.
// Returns a boolean program of |b|+1 rows and |a|+1 columns in
// row-major format. The 0th row and column can be ignored. If
// the program[i, j] is true then the longest common subsequence
// of a and b uses b[i-1] and a[j-1].
export function longestCommonSubsequence(a: string, b: string): Uint8Array {
    // TODO: Diff should be higher quality in cases like this:
    //
    // I(// This )S(f)I(u)S(n)I(ction)S( frozzle)I(s widgets)
    // I(fn frozzle)...
    //
    // Here we have eaten into the subsequent line. It would be
    // better to do:
    //
    // I(// This function frozzles widgets)
    // S(fn frozzle)
    //
    // The former is prone to conflicts where one agent is editing the
    // comment and the other is editing the function signature.

    // Construct a dynamic program of edits.
    const lenA = a.length
    const lenB = b.length
    const program = new Uint16Array((lenA + 1) * (lenB + 1))
    const ops = new Array<Op>((lenA + 1) * (lenB + 1))
    ops[0] = '*'
    // Top row: Delete all of the characters in A.
    for (let i = 1; i <= lenA; i++) {
        program[i] = i
        ops[i] = 'X'
    }
    // Left column: Insert all of the characters in B.
    for (let i = 1; i <= lenB; i++) {
        program[i * (lenA + 1)] = i
        ops[i * (lenA + 1)] = 'I'
    }
    for (let iB = 1; iB <= lenB; iB++) {
        const chB = b[iB - 1]
        for (let iA = 1; iA <= lenA; iA++) {
            const chA = a[iA - 1]
            const costDeleteA = program[iB * (lenA + 1) + iA - 1] + 1
            const costInsertB = program[(iB - 1) * (lenA + 1) + iA] + 1
            const costSkipReplace = program[(iB - 1) * (lenA + 1) + iA - 1] + (chA === chB ? 0 : 2)
            const cost = Math.min(costDeleteA, costInsertB, costSkipReplace)
            program[iB * (lenA + 1) + iA] = cost
            ops[iB * (lenA + 1) + iA] =
                cost === costSkipReplace ? (chA === chB ? '-' : 'R') : cost === costDeleteA ? 'X' : 'I'
        }
    }
    // dumpProgram(program, ops, a, b)
    const use = new Uint8Array((lenA + 1) * (lenB + 1))
    let i = lenA
    let j = lenB
    while (i !== 0 || j !== 0) {
        const op = ops[j * (lenA + 1) + i]
        switch (op) {
            case '-':
                use[j * (lenA + 1) + i] = 1
                i--
                j--
                break
            case 'R':
                i--
                j--
                break
            case 'X':
                i--
                break
            case 'I':
                j--
                break
            default:
                throw new Error('unreachable')
        }
    }
    return use
}

type Chunk = string[]

function computeChunks(original: string, a: string, b: string): Chunk[] {
    const useA = longestCommonSubsequence(original, a)
    // dumpUse(useA, original, a)
    const useB = longestCommonSubsequence(original, b)
    // dumpUse(useB, original, b)
    const chunks: Chunk[] = []
    let lO = 0
    let lA = 0
    let lB = 0
    outer: while (true) {
        for (let i = 1; lO + i <= original.length && (lA + i <= a.length || lB + i <= b.length); i++) {
            if (
                lA + i <= a.length &&
                useA[(lA + i) * (original.length + 1) + lO + i] &&
                lB + i <= b.length &&
                useB[(lB + i) * (original.length + 1) + lO + i]
            ) {
                // Skipping stable pieces.
                continue
            }
            if (i > 1) {
                // We found a stable chunk.
                chunks.push([a.slice(lA, lA + i - 1), original.slice(lO, lO + i - 1), b.slice(lB, lB + i - 1)])
                lO += i - 1
                lA += i - 1
                lB += i - 1
                continue outer
            }
            // i === 1
            // Skipping unstable pieces.
            for (let nextO = lO + 1; nextO <= original.length; nextO++) {
                // Find an A
                let nextA
                let nextB
                for (nextA = 1; nextA <= a.length; nextA++) {
                    if (useA[nextA * (original.length + 1) + nextO]) {
                        break
                    }
                }
                if (nextA > a.length) {
                    continue
                }
                // Find a B
                for (nextB = 1; nextB <= b.length; nextB++) {
                    if (useB[nextB * (original.length + 1) + nextO]) {
                        break
                    }
                }
                if (nextB > b.length) {
                    continue
                }
                // Output an unstable chunk
                chunks.push([a.slice(lA, nextA - 1), original.slice(lO, nextO - 1), b.slice(lB, nextB - 1)])
                lO = nextO - 1
                lA = nextA - 1
                lB = nextB - 1
                continue outer
            }
            break outer
        }
        break
    }
    // TODO: Could strip a prefix/suffix here
    if (lO <= original.length || lA <= a.length || lB <= b.length) {
        chunks.push([a.slice(lA), original.slice(lO), b.slice(lB)])
    }
    return chunks
}

export interface Position {
    line: number
    character: number
}

export interface Range {
    start: Position
    end: Position
}

export interface Edit {
    text: string
    range: Range
}

export interface Diff {
    originalText: string
    bufferText: string
    mergedText: string | undefined
    // TODO: We can use the presence of mergedText to indicate clean
    clean: boolean
    conflicts: Range[]
    edits: Edit[]
    highlights: Range[]
}

// Rolls over characters in text and computes the updated position to the end
// of the text.
function updatedPosition(position: Position, text: string): Position {
    // TODO: Handle Mac-style \r by itself.
    let { line, character } = position
    for (const ch of text) {
        if (ch === '\n') {
            line++
            character = 0
        } else {
            character++
        }
    }
    return { line, character }
}

// Given a source text, and two evolutions a and b, computes:
// - Whether the diff can be applied without conflicts.
// - A set of insertions and deletions to apply in b.
// - A set of ranges to highlight in the merged text, explaining edits in a.
// "a" is the text produced by Cody, which is treated as "foreign".
// "b" is the text produced by the human (perhaps applying Cody), which is
// treated as known.
export function computeDiff(original: string, a: string, b: string, bStart: Position): Diff {
    const chunks = computeChunks(original, a, b)
    const edits = []
    const conflicts = []
    const postEditHighlights = []
    let clean = true
    let originalPos = bStart
    let mergedPos = bStart
    const mergedText: string[] = []
    for (const chunk of chunks) {
        const originalEnd = updatedPosition(originalPos, chunk[2])
        if (chunk[1] === chunk[2] && chunk[0] !== chunk[1]) {
            // Changed by robot
            edits.push({
                kind: 'insert',
                text: chunk[0],
                range: {
                    start: originalPos,
                    end: originalEnd,
                },
            })
            mergedText.push(chunk[0])
            const mergedEnd = updatedPosition(mergedPos, chunk[0])
            if (clean) {
                postEditHighlights.push({ start: mergedPos, end: mergedEnd })
            }
            mergedPos = mergedEnd
        } else if (chunk[1] === chunk[0] && chunk[1] !== chunk[2]) {
            // Changed by human
            mergedPos = updatedPosition(mergedPos, chunk[2])
            mergedText.push(chunk[2])
        } else if (chunk[0] === chunk[2]) {
            // Changed by both, to the same thing
            mergedPos = updatedPosition(mergedPos, chunk[2])
            mergedText.push(chunk[2])
        } else {
            // Conflict! chunk[0]/chunk[2]`
            conflicts.push({ start: originalPos, end: originalEnd })
            clean = false
            // We give up on tracking positions.
        }
        originalPos = originalEnd
    }
    return {
        originalText: original,
        bufferText: b,
        mergedText: clean ? mergedText.join('') : undefined,
        clean,
        conflicts,
        edits,
        highlights: postEditHighlights,
    }
}

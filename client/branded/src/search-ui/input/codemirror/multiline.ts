import { type ChangeSpec, EditorState, Facet, type Extension } from '@codemirror/state'
import { EditorView } from '@codemirror/view'

const replacePattern = /[\n\r↵]+/g

/**
 * Facet indiciating whether multi-line mode is enabled or not.
 */
export const multiLineEnabled = Facet.define<boolean, boolean>({
    combine(values) {
        return values.some(value => value)
    },
})

/**
 * Returns an extension that enables/disables multi-line mode.
 * If multi-line is enabled, line wrapping is enabled by default.
 * If it is disabled, entering a line break (via keyboard or pasting)
 * will replace consecutive line breaks into a single space.
 *
 * NOTE 1: If a submit handler is assigned to the query input then pressing
 * enter won't insert a line break anyway. In that case, this extensions ensures
 * that line breaks are stripped from pasted input.
 *
 * NOTE 2: Line breaks from the initial value will have to be manually stripped, e.g.
 * with the {@link toSingleLine} function.
 */
export function multiline(multiline: boolean): Extension {
    return [
        multiline
            ? // Automatically enable line wrapping in multi-line mode
              EditorView.lineWrapping
            : // NOTE: If a submit handler is assigned to the query input then the pressing
              // enter won't insert a line break anyway. In that case, this extensions ensures
              // that line breaks are stripped from pasted input.
              singleLine,
        multiLineEnabled.of(multiline),
    ]
}

/**
 * Replaces all consecutive line breaks with a single space.
 */
export function toSingleLine(value: string): string {
    return value.replaceAll(replacePattern, ' ')
}

/**
 * An extension that enforces that the input will be single line. Consecutive
 * line breaks will be replaced by a single space.
 */
const singleLine = EditorState.transactionFilter.of(transaction => {
    if (!transaction.docChanged) {
        return transaction
    }

    const newText = transaction.newDoc.sliceString(0)
    const changes: ChangeSpec[] = []

    // new RegExp(...) creates a copy of the regular expression so that we have
    // our own stateful copy for using `exec` below.
    const lineBreakPattern = new RegExp(replacePattern)
    let match: RegExpExecArray | null = null
    while ((match = lineBreakPattern.exec(newText))) {
        // Insert space for line breaks following non-whitespace characters
        if (match.index > 0 && !/\s/.test(newText[match.index - 1])) {
            changes.push({ from: match.index, to: match.index + match[0].length, insert: ' ' })
        } else {
            // Otherwise remove it
            changes.push({ from: match.index, to: match.index + match[0].length })
        }
    }

    return changes.length > 0 ? [transaction, { changes, sequential: true }] : transaction
})

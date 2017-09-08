import * as moment from 'moment'
import { BlameState, contextKey, store } from 'sourcegraph/blame/store'

function limitString(s: string, n: number, dotdotdot: boolean): string {
    if (s.length > n) {
        if (dotdotdot) {
            return s.substring(0, n - 1) + '…'
        }
        return s.substring(0, n)
    }
    return s
}

/**
 * setLineBlameContent sets the given line's blame content.
 */
function setLineBlameContent(line: number, blameContent: string, rev?: string): void {
    // Remove blame class from all other lines.
    const currentlyBlamed = document.querySelectorAll('.blob td.code>.blame')
    for (const blame of currentlyBlamed) {
        blame.parentNode!.removeChild(blame)
    }

    if (line > 0) {
        // Add blame element to the target line's code cell.
        const cells = document.querySelectorAll('.blob td.code')
        const cell = cells[line - 1]
        if (!cell) {
            return
        }

        const blame = document.createElement('span')
        blame.classList.add('blame')
        blame.setAttribute('data-blame', blameContent)
        if (rev) {
            blame.setAttribute('data-blame-rev', rev)
        }
        if (cell.textContent === '\n') {
            /*
                Empty line, so appendChild would place this on the next line
                after \n not at the start before \n. Only empty lines contain a
                newline character.
            */
            cell.insertBefore(blame, cell.firstChild)
        } else {
            cell.appendChild(blame)
        }
    }
}

store.subscribe((state: BlameState) => {
    state = store.getValue()

    // Clear the blame content on whatever line it was already on.
    setLineBlameContent(-1, '')

    if (!state.context) {
        return
    }
    const hunks = state.hunksByLoc.get(contextKey(state.context))
    if (!hunks) {
        if (state.displayLoading) {
            setLineBlameContent(state.context.line, 'loading ◌')
        }
        return
    }
    const hunk = hunks[0]
    if (!hunk.author || !hunk.author.person) {
        // Clear the blame content on whatever line it was already on.
        setLineBlameContent(-1, '')
        return
    }

    const timeSince = moment(hunk.author.date, 'YYYY-MM-DD HH:mm:ss ZZ UTC').fromNow()
    const blameContent = `${hunk.author.person.name}, ${timeSince} • ${limitString(hunk.message, 80, true)} ${limitString(hunk.rev, 6, false)}`

    setLineBlameContent(state.context.line, blameContent, hunk.rev)
})

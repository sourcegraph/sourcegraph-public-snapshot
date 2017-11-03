import formatDistance from 'date-fns/formatDistance'
import { limitString } from '../util'
import { parseCommitDateString } from '../util/time'
import { BlameData } from './index'

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

export function clearLineBlameContent(): void {
    setLineBlameContent(-1, '')
}

export function setLineBlame(data: BlameData): void {
    clearLineBlameContent()

    if (!data.hunks || data.hunks.length === 0) {
        if (data.loading) {
            setLineBlameContent(data.ctx.range.start.line, 'loading ◌')
        }
        return
    }
    const hunk = data.hunks[0]
    if (!hunk || !hunk.author || !hunk.author.person) {
        return clearLineBlameContent()
    }

    const timeSince = formatDistance(parseCommitDateString(hunk.author.date), new Date(), { addSuffix: true })
    const blameContent = `${hunk.author.person.name}, ${timeSince} • ${limitString(hunk.message, 80, true)} ${limitString(hunk.rev, 6, false)}`

    setLineBlameContent(data.ctx.range.start.line, blameContent, hunk.rev)
}

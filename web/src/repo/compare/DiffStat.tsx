import * as React from 'react'
import { pluralize } from '../../../../shared/src/util/strings'

const NUM_SQUARES = 5

/** Displays a diff stat (visual representation of added, changed, and deleted lines in a diff). */
export const DiffStat: React.FunctionComponent<{
    /** Number of additions (added lines). */
    added: number

    /** Number of changes (changed lines). */
    changed: number

    /** Number of deletions (deleted lines). */
    deleted: number

    className?: string
}> = ({ added, changed, deleted, className = '' }) => {
    const total = added + changed + deleted
    const numSquares = Math.min(NUM_SQUARES, total)
    let addedSquares = allocateSquares(added, total)
    let changedSquares = allocateSquares(changed, total)
    let deletedSquares = allocateSquares(deleted, total)

    // Make sure we have exactly numSquares squares.
    const totalSquares = addedSquares + changedSquares + deletedSquares
    if (totalSquares < numSquares) {
        const deficit = numSquares - totalSquares
        if (deleted > changed && deleted > added) {
            deletedSquares += deficit
        } else if (changed > added && changed > deleted) {
            changedSquares += deficit
        } else {
            addedSquares += deficit
        }
    } else if (totalSquares > numSquares) {
        const surplus = numSquares - totalSquares
        if (deleted <= changed && deleted <= added) {
            deletedSquares -= surplus
        } else if (changed < added && changed < deleted) {
            changedSquares -= surplus
        } else {
            addedSquares -= surplus
        }
    }

    const squares: ('added' | 'changed' | 'deleted')[] = Array(addedSquares)
        .fill('added')
        .concat(
            Array(changedSquares).fill('changed'),
            Array(deletedSquares).fill('deleted'),
            Array(NUM_SQUARES - numSquares).fill('empty')
        )

    const labels: string[] = []
    if (added > 0) {
        labels.push(`${added} ${pluralize('addition', added)}`)
    }
    if (changed > 0) {
        labels.push(`${changed} ${pluralize('change', changed)}`)
    }
    if (deleted > 0) {
        labels.push(`${deleted} ${pluralize('deletion', deleted)}`)
    }
    return (
        <div className={`diff-stat ${className}`} data-tooltip={labels.join(', ')}>
            <small className="diff-stat__total">{total}</small>
            {squares.map((verb, i) => (
                <div key={i} className={`diff-stat__square diff-stat__${verb}`} />
            ))}
        </div>
    )
}

function allocateSquares(n: number, total: number): number {
    if (total === 0) {
        return 0
    }
    return Math.max(Math.round(n / total), n > 0 ? 1 : 0)
}

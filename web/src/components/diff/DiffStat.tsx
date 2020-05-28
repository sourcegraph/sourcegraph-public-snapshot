import * as React from 'react'
import { numberWithCommas, pluralize } from '../../../../shared/src/util/strings'

const NUM_SQUARES = 5

/** Displays a diff stat (visual representation of added, changed, and deleted lines in a diff). */
export const DiffStat: React.FunctionComponent<{
    /** Number of additions (added lines). */
    added: number

    /** Number of changes (changed lines). */
    changed: number

    /** Number of deletions (deleted lines). */
    deleted: number

    /* Show +/- numbers, not just the total change count. */
    expandedCounts?: boolean

    className?: string
}> = ({ added, changed, deleted, expandedCounts = false, className = '' }) => {
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
        labels.push(`${numberWithCommas(added)} ${pluralize('addition', added)}`)
    }
    if (changed > 0) {
        labels.push(`${numberWithCommas(changed)} ${pluralize('change', changed)}`)
    }
    if (deleted > 0) {
        labels.push(`${numberWithCommas(deleted)} ${pluralize('deletion', deleted)}`)
    }
    return (
        <div className={`diff-stat ${className}`} data-tooltip={labels.join(', ')}>
            {expandedCounts ? (
                <span className="diff-stat__total font-weight-bold">
                    <span className="diff-stat__text-added mr-1">+{numberWithCommas(added)}</span>
                    {changed > 0 && (
                        <span className="diff-stat__text-changed mr-1">&bull;{numberWithCommas(changed)}</span>
                    )}
                    <span className="diff-stat__text-deleted mr-1">&minus;{numberWithCommas(deleted)}</span>
                </span>
            ) : (
                <small className="diff-stat__total">{numberWithCommas(total + changed)}</small>
            )}
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

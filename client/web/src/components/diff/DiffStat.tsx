import * as React from 'react'

import classNames from 'classnames'

import { numberWithCommas, pluralize } from '@sourcegraph/common'
import { Tooltip } from '@sourcegraph/wildcard'

import styles from './DiffStat.module.scss'

const NUM_SQUARES = 5

interface DiffProps {
    /** Number of additions (added lines). */
    added: number

    /** Number of changes (changed lines). */
    changed: number

    /** Number of deletions (deleted lines). */
    deleted: number
}

interface DiffStatProps extends DiffProps {
    /* Show +/- numbers, not just the total change count. */
    expandedCounts?: boolean

    className?: string
}

/** Displays a diff stat (visual representation of added, changed, and deleted lines in a diff). */
export const DiffStat: React.FunctionComponent<React.PropsWithChildren<DiffStatProps>> = React.memo(function DiffStat({
    added,
    changed,
    deleted,
    expandedCounts = false,
    className = '',
}) {
    const total = added + changed + deleted

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
        <div className="flex-1">
            <Tooltip content={labels.join(', ')}>
                <div className={classNames(styles.diffStat, className)} aria-label={labels.join(', ')}>
                    {expandedCounts ? (
                        <>
                            {/*
                                a11y-ignore
                                Rule: "color-contrast" (Elements must have sufficient color contrast)
                                GitHub issue: https://github.com/sourcegraph/sourcegraph/issues/33343
                            */}
                            <strong className="a11y-ignore text-success mr-1">+{numberWithCommas(added)}</strong>
                            {changed > 0 && (
                                <strong className="a11y-ignore text-warning mr-1">
                                    &bull;{numberWithCommas(changed)}
                                </strong>
                            )}
                            <strong className="a11y-ignore text-danger">&minus;{numberWithCommas(deleted)}</strong>
                        </>
                    ) : (
                        <small>{numberWithCommas(total + changed)}</small>
                    )}
                </div>
            </Tooltip>
        </div>
    )
})

export const DiffStatSquares: React.FunctionComponent<React.PropsWithChildren<DiffProps>> = React.memo(
    function DiffStatSquares({ added, changed, deleted }) {
        const total = added + changed + deleted
        const numberOfSquares = Math.min(NUM_SQUARES, total)
        let addedSquares = allocateSquares(added, total)
        let changedSquares = allocateSquares(changed, total)
        let deletedSquares = allocateSquares(deleted, total)

        // Make sure we have exactly numSquares squares.
        const totalSquares = addedSquares + changedSquares + deletedSquares
        if (totalSquares < numberOfSquares) {
            const deficit = numberOfSquares - totalSquares
            if (deleted > changed && deleted > added) {
                deletedSquares += deficit
            } else if (changed > added && changed > deleted) {
                changedSquares += deficit
            } else {
                addedSquares += deficit
            }
        } else if (totalSquares > numberOfSquares) {
            const surplus = numberOfSquares - totalSquares
            if (deleted <= changed && deleted <= added) {
                deletedSquares -= surplus
            } else if (changed < added && changed < deleted) {
                changedSquares -= surplus
            } else {
                addedSquares -= surplus
            }
        }

        const squares = new Array<'bg-success' | 'bg-warning' | 'bg-danger'>(addedSquares)
            .fill('bg-success')
            .concat(
                new Array<'bg-warning'>(changedSquares).fill('bg-warning'),
                new Array<'bg-danger'>(deletedSquares).fill('bg-danger'),
                new Array(NUM_SQUARES - numberOfSquares).fill(styles.empty)
            )

        return (
            <div className={styles.squares}>
                {squares.map((className, index) => (
                    // eslint-disable-next-line react/no-array-index-key
                    <div key={index} className={classNames(styles.square, className)} />
                ))}
            </div>
        )
    }
)

interface DiffStatStackProps extends DiffProps {
    className?: string
}

/** A commonly used combined vertical stack of a `DiffStat` and a `DiffStatSquares` */
export const DiffStatStack: React.FunctionComponent<React.PropsWithChildren<DiffStatStackProps>> = React.memo(
    function DiffStatStack({ className = '', ...props }) {
        return (
            <div className={classNames('d-flex flex-column align-items-center', className)}>
                <DiffStat className="pb-1" expandedCounts={true} {...props} />
                <DiffStatSquares {...props} />
            </div>
        )
    }
)

function allocateSquares(number: number, total: number): number {
    if (total === 0) {
        return 0
    }
    return Math.max(Math.round(number / total), number > 0 ? 1 : 0)
}

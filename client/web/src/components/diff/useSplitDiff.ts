import { useMemo } from 'react'
import { Hunk } from './DiffSplitHunk'
import { DiffHunkLineType } from '../../graphql-operations'

type HunkZipped = [Hunk[], Hunk | undefined, number]

const zipHunks = (hunks: Hunk[]): HunkZipped =>
    hunks.reduce(
        ([result, last, lastDeletionIndex], current, index): HunkZipped => {
            if (!last) {
                result.push(current)
                return [result, current, current.kind === DiffHunkLineType.DELETED ? index : -1]
            }

            if (current.kind === DiffHunkLineType.ADDED && lastDeletionIndex >= 0) {
                result.splice(lastDeletionIndex + 1, 0, current)
                return [result, current, lastDeletionIndex + 2]
            }

            result.push(current)

            // Preserve `lastDeletionIndex` if there are lines of deletions,
            // otherwise update it to the new deletion line
            let newLastDeletionIndex = -1
            if (current.kind === DiffHunkLineType.DELETED) {
                if (last.kind === DiffHunkLineType.DELETED) {
                    newLastDeletionIndex = lastDeletionIndex
                } else {
                    newLastDeletionIndex = index
                }
            }
            return [result, current, newLastDeletionIndex]
        },
        [[], undefined, -1] as HunkZipped
    )

export const useSplitDiff = (hunks: Hunk[]): { diff: Hunk[] } => {
    const [zippedHunks] = useMemo(() => zipHunks(hunks), [hunks])
    return { diff: zippedHunks }
}

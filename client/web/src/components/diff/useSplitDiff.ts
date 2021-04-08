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
                // The new `lastDeletionIndex` may be out of range, but `splice` will fix it
                return [result, current, lastDeletionIndex + 2]
            }

            result.push(current)

            // Keep the `lastDeletionIndex` if there are lines of deletions,
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

// const groupHunks = (hunks: Hunk[]): (Hunk | undefined)[][] => {
//     const elements: (Hunk | undefined)[][] = []
//     // This could be a very complex reduce call, use `for` loop seems to make it a little more readable
//     for (let index = 0; index < hunks.length; index++) {
//         const current = hunks[index]

//         // A normal change is displayed on both side
//         if (current.kind === DiffHunkLineType.UNCHANGED) {
//             elements.push([current, current])
//         } else if (current.kind === DiffHunkLineType.DELETED) {
//             const next = hunks[index + 1]
//             // If an insert change is following a delete change, they should be displayed side by side
//             console.log(next)
//             if (next?.kind === DiffHunkLineType.ADDED) {
//                 index = index + 1
//                 elements.push([current, next])
//             } else {
//                 elements.push([current, undefined])
//             }
//         } else {
//             elements.push([undefined, current])
//         }
//     }

//     return elements
// }

export const useSplitDiff = (hunks: Hunk[]): { diff: Hunk[] } => {
    const [zippedHunks] = useMemo(() => zipHunks(hunks), [hunks])
    // const diff = useMemo(() => groupHunks(zippedHunks), [zippedHunks])

    return { diff: zippedHunks }
}

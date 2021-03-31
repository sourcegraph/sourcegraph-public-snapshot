import { useMemo } from 'react'

enum DiffHunkLineType {
    ADDED = 'ADDED',
    UNCHANGED = 'UNCHANGED',
    DELETED = 'DELETED',
}

interface Hunk {
    kind: DiffHunkLineType
    html: string
    line: number
    oldLine: number
    newLine: number
}

type HunkZipped = [Hunk[], Hunk | null, number]

const zipHunks = (hunks: Hunk[]) =>
    hunks.reduce(
        ([result, last, lastDeletionIndex], current, i): HunkZipped => {
            if (!last) {
                result.push(current)
                return [result, current, current.kind === DiffHunkLineType.DELETED ? i : -1]
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
                    newLastDeletionIndex = i
                }
            }
            return [result, current, newLastDeletionIndex]
        },
        <HunkZipped>[[], null, -1]
    )

const groupHunks = (hunks: Hunk[]) => {
    const elements = []

    // This could be a very complex reduce call, use `for` loop seems to make it a little more readable
    for (let i = 0; i < hunks.length; i++) {
        const current = hunks[i]

        // A normal change is displayed on both side
        if (current.kind === DiffHunkLineType.UNCHANGED) {
            elements.push([current, current])
        } else if (current.kind === DiffHunkLineType.DELETED) {
            const next = hunks[i + 1]
            // If an insert change is following a delete change, they should be displayed side by side
            if (next.kind === DiffHunkLineType.ADDED) {
                i = i + 1
                elements.push([current, next])
            } else {
                elements.push([current, null])
            }
        } else {
            elements.push([null, current])
        }
    }

    return elements
}

export const useSplitDiff = (hunks: Hunk[]) => {
    const [zippedHunks] = useMemo(() => zipHunks(hunks), [hunks])
    const diff = useMemo(() => groupHunks(zippedHunks), [zippedHunks])

    return { diff }
}

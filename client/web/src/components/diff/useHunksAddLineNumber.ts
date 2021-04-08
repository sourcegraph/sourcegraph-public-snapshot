import { Hunk } from './DiffSplitHunk'

enum DiffHunkLineType {
    ADDED = 'ADDED',
    UNCHANGED = 'UNCHANGED',
    DELETED = 'DELETED',
}

export const useHunksAddLineNumber = (
    hunks: {
        kind: DiffHunkLineType
        html: string
    }[],
    newStartLine: number,
    oldStartLine: number,
    fileDiffAnchor: string
): { hunksWithLineNumber: Hunk[] } => {
    let oldLine = oldStartLine
    let newLine = newStartLine

    const hunksWithLineNumber = hunks.map((hunkWithLineNumber: { kind: DiffHunkLineType; html: string }) => {
        const hunk: Hunk = {
            kind: hunkWithLineNumber.kind,
            html: hunkWithLineNumber.html,
            oldLine: 0,
            newLine: 0,
            anchor: '',
        }
        if (hunkWithLineNumber.kind === DiffHunkLineType.DELETED) {
            oldLine++
            hunk.oldLine = oldLine - 1
            hunk.anchor = `${fileDiffAnchor}L${oldLine - 1}`
        }
        if (hunkWithLineNumber.kind === DiffHunkLineType.ADDED) {
            newLine++
            hunk.newLine = newLine - 1
            hunk.anchor = `${fileDiffAnchor}R${newLine - 1}`
        }
        if (hunkWithLineNumber.kind === DiffHunkLineType.UNCHANGED) {
            oldLine++
            newLine++
            hunk.newLine = newLine - 1
            hunk.oldLine = oldLine - 1
            hunk.anchor = `${fileDiffAnchor}L${oldLine - 1}`
        }

        return hunk
    })

    return { hunksWithLineNumber }
}

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

export const useHooksAddLineNumber = (
    hunks: Hunk[],
    newStartLine: number,
    oldStartLine: number
): { hunksWithLineNumber: Hunk[] } => {
    let oldLine = oldStartLine
    let newLine = newStartLine

    const hunksWithLineNumber = hunks.map((hunkWithLineNumber, index) => {
        if (hunkWithLineNumber.kind === DiffHunkLineType.DELETED) {
            oldLine++
            hunkWithLineNumber.oldLine = oldLine - 1
        }
        if (hunkWithLineNumber.kind === DiffHunkLineType.ADDED) {
            newLine++
            hunkWithLineNumber.newLine = newLine - 1
        }
        if (hunkWithLineNumber.kind === DiffHunkLineType.UNCHANGED) {
            oldLine++
            newLine++
            hunkWithLineNumber.newLine = newLine - 1
            hunkWithLineNumber.oldLine = oldLine - 1
        }

        return hunkWithLineNumber
    })

    return { hunksWithLineNumber }
}

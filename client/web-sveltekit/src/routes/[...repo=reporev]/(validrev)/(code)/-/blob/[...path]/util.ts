export enum FileViewMode {
    CodeFile = 2 ** 1,
    BinaryFile = 2 ** 2,
    AtRevision = 2 ** 3,
    Diff = 2 ** 4,
}

export function isFileViewMode(viewMode: FileViewMode, mode: FileViewMode): boolean {
    return (viewMode & mode) != 0
}

export enum CodeViewMode {
    Default = 'default',
    Code = 'code',
    Blame = 'blame',
}

/**
 * Converts the raw view mode value (e.g. from a URL parameter) to a ViewMode enum value.
 *
 * @param rawViewMode The raw view mode value.
 * @returns The ViewMode enum value.
 */
export function toCodeViewMode(rawViewMode: string | null | undefined): CodeViewMode {
    switch (rawViewMode?.toLowerCase()) {
        case 'code':
        case 'raw':
            return CodeViewMode.Code
        case 'blame':
            return CodeViewMode.Blame
        default:
            return CodeViewMode.Default
    }
}

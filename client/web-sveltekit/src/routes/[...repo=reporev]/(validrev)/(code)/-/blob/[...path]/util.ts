export enum ViewMode {
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
export function toViewMode(rawViewMode: string | null | undefined): ViewMode {
    switch (rawViewMode?.toLowerCase()) {
        case 'code':
        case 'raw':
            return ViewMode.Code
        case 'blame':
            return ViewMode.Blame
        default:
            return ViewMode.Default
    }
}

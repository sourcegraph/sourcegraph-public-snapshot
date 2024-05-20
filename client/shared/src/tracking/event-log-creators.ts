export type V2CodyCopyPages = 'search-result' | 'file-match' | 'notebook-file-block' | 'blob-view' | 'notebook-symbols'

export const V2CodyCopyPageTypes: { [key in V2CodyCopyPages]: number } = {
    'search-result': 1,
    'file-match': 2,
    'notebook-file-block': 3,
    'blob-view': 4,
    'notebook-symbols': 5,
}

/**
 * Returns an [EventName, argument, publicArgument] for a "code copied" to clipboard event
 *
 * @param page page or component from where copied
 */
export const codeCopiedEvent = (page: V2CodyCopyPages): [string, { page: string }, { page: string }] => [
    'CodeCopied',
    { page },
    { page },
]

/**
 * Returns an [EventName, argument, publicArgument] for a "code copied" to clipboard event
 *
 * @param page page or component from where copied
 *
 * TODO: this is an experiment to check how often users copy from SG, will be removed soon
 */
export const codeCopiedEvent = (page: string): [string, { page: string }, { page: string }] => [
    'CodeCopied',
    { page },
    { page },
]

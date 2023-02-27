import * as jsonc from 'jsonc-parser'

export const REPOSITORY_KEY = 'repository'
export const REVISIONS_KEY = 'revisions'
export const MAX_REVISION_LENGTH = 255

export interface RepositoryRevisions {
    [REPOSITORY_KEY]: string
    [REVISIONS_KEY]: string[]
}

const validateRevisions = (revisions: any[]): boolean =>
    revisions.every(revision => typeof revision === 'string' && revision.length < MAX_REVISION_LENGTH)

export function parseConfig(configJSON: string): RepositoryRevisions[] | null {
    const errors: jsonc.ParseError[] = []
    const config = jsonc.parse(configJSON || '[]', errors)

    if (!config || errors.length > 0 || !Array.isArray(config)) {
        return null
    }

    const validItems = config.every(
        item =>
            item &&
            typeof item === 'object' &&
            REPOSITORY_KEY in item &&
            REVISIONS_KEY in item &&
            typeof item[REPOSITORY_KEY] === 'string' &&
            Array.isArray(item[REVISIONS_KEY]) &&
            validateRevisions(item[REVISIONS_KEY])
    )

    return validItems ? (config as RepositoryRevisions[]) : null
}

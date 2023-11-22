import { at } from 'lodash'

import { type GitCommitFields, RepositoryType } from '../graphql-operations'

import { CodeHostType, FileExtension } from './constants'

export const isPerforceChangelistMappingEnabled = (): boolean =>
    window.context.experimentalFeatures.perforceChangelistMapping === 'enabled'

export const isPerforceDepotSource = (sourceType: string): boolean => sourceType === RepositoryType.PERFORCE_DEPOT

export const getRefType = (sourceType: RepositoryType | string): string =>
    isPerforceDepotSource(sourceType) ? 'changelist' : 'commit'

export const getCanonicalURL = (sourceType: RepositoryType | string, node: GitCommitFields): string =>
    isPerforceChangelistMappingEnabled() && isPerforceDepotSource(sourceType) && node.perforceChangelist
        ? node.perforceChangelist.canonicalURL
        : node.canonicalURL

export const getInitialSearchTerm = (repo: string): string => {
    const r = repo.split('/')
    return r.at(-1)?.trim() ?? ''
}

export const stringToCodeHostType = (codeHostType: string): CodeHostType => {
    switch (codeHostType) {
        case 'github': {
            return CodeHostType.GITHUB
        }
        case 'gitlab': {
            return CodeHostType.GITLAB
        }
        case 'bitbucketCloud': {
            return CodeHostType.BITBUCKETCLOUD
        }
        case 'gitolite': {
            return CodeHostType.GITOLITE
        }
        case 'awsCodeCommit': {
            return CodeHostType.AWSCODECOMMIT
        }
        case 'azureDevOps': {
            return CodeHostType.AZUREDEVOPS
        }
        default: {
            return CodeHostType.OTHER
        }
    }
}

export const contains = (arr: string[], target: string): boolean => {
    for (const i of arr) {
        if (i === target) {
            return true
        }
    }
    return false
}

export const getExtension = (file: string): { extension: FileExtension; isTest: boolean } => {
    const e = { extension: '' as FileExtension, isTest: false }
    const f = file.split('.')
    if (contains(f, 'test')) {
        e.isTest = true
    }

    const s = f.slice(1)
    if (contains(s, 'mod') || contains(s, 'sum')) {
        e.extension = 'go' as FileExtension
    } else {
        // This is what the linter wants *shrugs*
        e.extension = s[s.length - 1] as FileExtension
    }
    console.log(file, e)
    return e
}

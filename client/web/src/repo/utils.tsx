import { includes } from 'lodash'

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

export interface FileInfo {
    extension: FileExtension
    isTest: boolean
}

export const getFileInfo = (file: string, isDirectory: boolean): FileInfo => {
    const fileInfo = { extension: '' as FileExtension, isTest: false }
    fileInfo.isTest = isDirectory ? false : containsTest(file)

    if (file.split('.').length === 1) {
        return fileInfo
    }

    const f = file.split('.')
    // Last item in 'f' is file extension string
    fileInfo.extension = f.pop() as FileExtension
    return fileInfo
}

export const containsTest = (file: string): boolean => {
    const f = file.split('.')
    const matchTest1 = '_test'
    const matchTest2 = 'test_'
    for (const i of f) {
        if (i === 'test') {
            return true
        }
        if (includes(i, matchTest1) || includes(i, matchTest2)) {
            return true
        }
    }
    return false
}

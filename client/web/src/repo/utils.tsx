import { type GitCommitFields, RepositoryType } from '../graphql-operations'

import { CodeHostType } from './constants'
import {
    mdiLanguageTypescript,
    mdiLanguageCss3,
    mdiTestTube,
    mdiFileDocumentOutline,
    mdiReact,
    mdiSass,
    mdiLanguageGo,
    mdiCogOutline,
    mdiGraphql,
    mdiLanguageMarkdown,
    mdiLanguageJavascript,
    mdiCodeJson,
    mdiGit,
    mdiLanguageLua,
    mdiText,
    mdiLanguagePython,
    mdiDatabaseOutline,
} from "@mdi/js"

import styles from "./RepoRevisionSidebarFileTree.module.scss"

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

const contains = (arr: string[], target: string): boolean => {
    for (let i = 0; i < arr.length; i++) {
        if (arr[i] === target) {
            return true
        }
    }
    return false
}

export const getIcon = (file: string): { icon: string; iconClass: string } => {
    const s = file.split(".").slice(1)
    let extension: string

    if (s.length === 1) {
        extension = s[0]
    }

    // handle test files separately
    if (contains(s, "test")) {
        extension = "test"
    } else if (contains(s, "mod") || contains(s, "sum")) {
        extension = "go"
    } else {
        extension = s[s.length - 1]
    }

    switch (extension) {
        case "ts": {
            return { icon: mdiLanguageTypescript, iconClass: styles.blue }
        }
        case "tsx": {
            return { icon: mdiReact, iconClass: styles.blue }
        }
        case "js": {
            return { icon: mdiLanguageJavascript, iconClass: styles.yellow }
        }
        case "jsx": {
            return { icon: mdiReact, iconClass: styles.blue }
        }
        case "scss": {
            return { icon: mdiSass, iconClass: styles.pink }
        }
        case "css": {
            return { icon: mdiLanguageCss3, iconClass: styles.blue }
        }
        case "go": {
            return { icon: mdiLanguageGo, iconClass: styles.blue }
        }
        case "lua": {
            return { icon: mdiLanguageLua, iconClass: styles.blue }
        }
        case "yaml": {
            return { icon: mdiCogOutline, iconClass: styles.gray }
        }
        case "yml": {
            return { icon: mdiCogOutline, iconClass: styles.gray }
        }
        case "py": {
            return { icon: mdiLanguagePython, iconClass: styles.yellow }
        }
        case "sql": {
            return { icon: mdiDatabaseOutline, iconClass: styles.blue }
        }
        case "graphql": {
            return { icon: mdiGraphql, iconClass: styles.pink }
        }
        case "md": {
            return { icon: mdiLanguageMarkdown, iconClass: styles.blue }
        }
        case "txt": {
            return { icon: mdiText, iconClass: styles.defaultIcon }
        }
        case "test": {
            return { icon: mdiTestTube, iconClass: styles.blue }
        }
        case "json": {
            return { icon: mdiCodeJson, iconClass: styles.defaultIcon }
        }
        case "lock": {
            return { icon: mdiCodeJson, iconClass: styles.defaultIcon }
        }
        case "gitignore": {
            return { icon: mdiGit, iconClass: styles.red }
        }
        default: {
            return { icon: mdiFileDocumentOutline, iconClass: styles.defaultIcon }
        }
    }
}

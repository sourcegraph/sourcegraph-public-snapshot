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
    mdiDocker,
    mdiHeart,
    mdiNpm,
    mdiConsole,
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

export const contains = (arr: string[], target: string): boolean => {
    for (let i = 0; i < arr.length; i++) {
        if (arr[i] === target) {
            return true
        }
    }
    return false
}

export const getExtension = (file: string): string => {
    const testRegex = /test/;
    if (testRegex.test(file)) {
        return "test"
    }

    const s = file.split(".").slice(1)
    // if (s.length === 1) {
    //     return s[0]
    // }

    // handle test files separately
    if (contains(s, "mod") || contains(s, "sum")) {
        return "go"
    } else {
        return s[s.length - 1]
    }
}

export const getIcon = (file: string, isBranch: boolean): { icon: string; iconClass: string } => {
    if (isBranch) {
        return { icon: mdiFileDocumentOutline, iconClass: styles.defaultIcon }
    }
    const extension = getExtension(file)

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
            return { icon: mdiTestTube, iconClass: styles.cyan }
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
        case "gitattributes": {
            return { icon: mdiGit, iconClass: styles.red }
        }
        case "dockerignore": {
            return { icon: mdiDocker, iconClass: styles.blue }
        }
        case "bazel": {
            return { icon: mdiHeart, iconClass: styles.green }
        }
        case "bzl": {
            return { icon: mdiHeart, iconClass: styles.green }
        }
        case "bazelignore": {
            return { icon: mdiHeart, iconClass: styles.green }
        }
        case "bazeliskrc": {
            return { icon: mdiHeart, iconClass: styles.green }
        }
        case "bazelrc": {
            return { icon: mdiHeart, iconClass: styles.green }
        }
        case "bazelversion": {
            return { icon: mdiHeart, iconClass: styles.green }
        }
        case "npmrc": {
            return { icon: mdiNpm, iconClass: styles.red }
        }
        case "sh": {
            return { icon: mdiConsole, iconClass: styles.defaultIcon }
        }

        default: {
            return { icon: mdiFileDocumentOutline, iconClass: styles.defaultIcon }
        }
    }
}


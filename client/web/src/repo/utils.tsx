import { type GitCommitFields, RepositoryType } from '../graphql-operations'

import { CodeHostType } from './constants'
import { CiSettings, CiTextAlignLeft } from 'react-icons/ci'
import { MdOutlineScience } from 'react-icons/md'
import { GoDatabase, GoTerminal } from 'react-icons/go'
import {
    SiC,
    SiCplusplus,
    SiCsharp,
    SiCssmodules,
    SiDocker,
    SiGit,
    SiGo,
    SiGraphql,
    SiHtml5,
    SiJavascript,
    SiJson,
    SiKotlin,
    SiLua,
    SiMarkdown,
    SiNixos,
    SiNpm,
    SiPerl,
    SiPhp,
    SiPython,
    SiR,
    SiReact,
    SiRubygems,
    SiRust,
    SiScala,
    SiSvg,
    SiTypescript,
    SiZig
} from 'react-icons/si'

import styles from "./RepoRevisionSidebarFileTree.module.scss"
import { IconType } from '@sourcegraph/wildcard'
import { FaJava, FaRegFile, FaSass } from 'react-icons/fa'
import { CustomIcon } from '@sourcegraph/wildcard/src/components/Icon'

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
    const f = file.split(".")
    if (contains(f, "test")) {
        return "test"
    }

    const s = f.slice(1)

    if (contains(s, "mod") || contains(s, "sum")) {
        return "go"
    } else {
        return s[s.length - 1]
    }
}

export const getIcon = (file: string, isBranch: boolean): {
    icon: CustomIcon;
    iconClass: string;
} => {
    if (isBranch) {
        return { icon: FaRegFile, iconClass: styles.defaultIcon }
    }
    const extension = getExtension(file)

    switch (extension) {
        case "c": {
            return { icon: SiC, iconClass: styles.blue }
        }
        case "cc": {
            return { icon: SiCplusplus, iconClass: styles.blue }
        }
        case "cs": {
            return { icon: SiCsharp, iconClass: styles.blue }
        }
        case "css": {
            return { icon: SiCssmodules, iconClass: styles.blue }
        }
        case "dockerignore": {
            return { icon: SiDocker, iconClass: styles.blue }
        }
        case "gitignore": {
            return { icon: SiGit, iconClass: styles.red }
        }
        case "gitattributes": {
            return { icon: SiGit, iconClass: styles.red }
        }
        case "go": {
            return { icon: SiGo, iconClass: styles.blue }
        }
        case "graphql": {
            return { icon: SiGraphql, iconClass: styles.pink }
        }
        case "html": {
            return { icon: SiHtml5, iconClass: styles.blue }
        }
        case "java": {
            return { icon: FaJava, iconClass: styles.defaultIcon }
        }
        case "js": {
            return { icon: SiJavascript, iconClass: styles.yellow }
        }
        case "json": {
            return { icon: SiJson, iconClass: styles.defaultIcon }
        }
        case "jsx": {
            return { icon: SiReact, iconClass: styles.yellow }
        }
        case "kt": {
            return { icon: SiKotlin, iconClass: styles.green }
        }
        case "lock": {
            return { icon: SiJson, iconClass: styles.defaultIcon }
        }
        case "lua": {
            return { icon: SiLua, iconClass: styles.blue }
        }
        case "md": {
            return { icon: SiMarkdown, iconClass: styles.blue }
        }
        case "ncl": {
            return { icon: CiSettings, iconClass: styles.gray }
        }
        case "nix": {
            return { icon: SiNixos, iconClass: styles.gray }
        }
        case "npmrc": {
            return { icon: SiNpm, iconClass: styles.red }
        }
        case "php": {
            return { icon: SiPhp, iconClass: styles.defaultIcon }
        }
        case "pl": {
            return { icon: SiPerl, iconClass: styles.defaultIcon }
        }
        case "py": {
            return { icon: SiPython, iconClass: styles.yellow }
        }
        case "r": {
            return { icon: SiR, iconClass: styles.red }
        }
        case "rb": {
            return { icon: SiRubygems, iconClass: styles.red }
        }
        case "rs": {
            return { icon: SiRust, iconClass: styles.defaultIcon }
        }
        case "scala": {
            return { icon: SiScala, iconClass: styles.red }
        }
        case "scss": {
            return { icon: FaSass, iconClass: styles.pink }
        }
        case "sh": {
            return { icon: GoTerminal, iconClass: styles.defaultIcon }
        }
        case "sql": {
            return { icon: GoDatabase, iconClass: styles.blue }
        }
        case "svg": {
            return { icon: SiSvg, iconClass: styles.blue }
        }
        case "test": {
            return { icon: MdOutlineScience, iconClass: styles.defaultIcon }
        }
        case "ts": {
            return { icon: SiTypescript, iconClass: styles.blue }
        }
        case "tsx": {
            return { icon: SiReact, iconClass: styles.blue }
        }
        case "txt": {
            return { icon: CiTextAlignLeft, iconClass: styles.defaultIcon }
        }
        case "yaml": {
            return { icon: CiSettings, iconClass: styles.gray }
        }
        case "yml": {
            return { icon: CiSettings, iconClass: styles.gray }
        }
        case "zig": {
            return { icon: SiZig, iconClass: styles.yellow }
        }
        default: {
            return { icon: FaRegFile, iconClass: styles.defaultIcon }
        }
    }
}


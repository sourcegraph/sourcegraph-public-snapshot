import { ComponentType } from 'react'

import { CiSettings, CiTextAlignLeft } from 'react-icons/ci'
import { FaJava, FaSass } from 'react-icons/fa'
import { GoDatabase, GoTerminal } from 'react-icons/go'
import {
    SiAssemblyscript,
    SiC,
    SiClojure,
    SiCplusplus,
    SiCsharp,
    SiCssmodules,
    SiDart,
    SiDocker,
    SiFortran,
    SiFsharp,
    SiGit,
    SiGo,
    SiGraphql,
    SiHaskell,
    SiHtml5,
    SiJavascript,
    SiJulia,
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
    SiRuby,
    SiRust,
    SiScala,
    SiSvg,
    SiSwift,
    SiTypescript,
    SiVisualbasic,
    SiZig,
} from 'react-icons/si'
import { VscJson } from 'react-icons/vsc'

import styles from './RepoRevisionSidebarFileTree.module.scss'

export const LogsPageTabs = {
    COMMANDS: 0,
    SYNCLOGS: 1,
} as const

export enum CodeHostType {
    GITHUB = 'github',
    BITBUCKETCLOUD = 'bitbucketCloud',
    BITBUCKETSERVER = 'bitbucketServer',
    GITLAB = 'gitlab',
    GITOLITE = 'gitolite',
    AWSCODECOMMIT = 'awsCodeCommit',
    AZUREDEVOPS = 'azureDevOps',
    OTHER = 'other',
}

export enum FileExtension {
    ASSEMBLY = 'asm',
    BASH = 'sh',
    BASIC = 'vb',
    C = 'c',
    // CLOJURE file extensions: .clj .cljs .cljr .cljc .edn
    CLOJURE = 'clj',
    CPP = 'cc',
    CSHARP = 'cs',
    CSS = 'css',
    DART = 'dart',
    DOCKERIGNORE = 'dockerignore',
    // FORTRAN file extension: f, for, ftn
    FORTRAN = 'f',
    // F# file extensions: fs, fsi, fsx
    FSHARP = 'fs',
    GITIGNORE = 'gitignore',
    GITATTRIBUTES = 'gitattributes',
    GO = 'go',
    GRAPHQL = 'graphql',
    HASKELL = 'hs',
    HTML = 'html',
    JAVA = 'java',
    JAVASCRIPT = 'js',
    JSON = 'json',
    JSX = 'jsx',
    JULIA = 'jl',
    KOTLIN = 'kt',
    LOCKFILE = 'lock',
    LUA = 'lua',
    MARKDOWN = 'md',
    NCL = 'ncl',
    NIX = 'nix',
    NPM = 'npmrc',
    PHP = 'php',
    PERL = 'pl',
    PYTHON = 'py',
    R = 'r',
    RUBY = 'rb',
    RUST = 'rs',
    SCALA = 'scala',
    SASS = 'scss',
    SQL = 'sql',
    SVG = 'svg',
    SWIFT = 'swift',
    TEST = 'test',
    TYPESCRIPT = 'ts',
    TSX = 'tsx',
    TEXT = 'txt',
    YAML = 'yaml',
    YML = 'yml',
    ZIG = 'zig',
}

type CustomIcon = ComponentType<{ className?: string }>
/*
 * We use the react-icons package instead of material design icons for two reasons:
 * 1) Many of mdi's programming language icons will be deprecated soon.
 * 2) They are missing quite a few icons that are needed when displaying file types.
 */
export const FILE_ICONS: Map<FileExtension, { icon: CustomIcon; iconClass: string }> = new Map([
    [FileExtension.ASSEMBLY, { icon: SiAssemblyscript, iconClass: styles.defaultIcon }],
    [FileExtension.BASH, { icon: GoTerminal, iconClass: styles.defaultIcon }],
    [FileExtension.BASIC, { icon: SiVisualbasic, iconClass: styles.defaultIcon }],
    [FileExtension.C, { icon: SiC, iconClass: styles.blue }],
    [FileExtension.CLOJURE, { icon: SiClojure, iconClass: styles.defaultIcon }],
    [FileExtension.CPP, { icon: SiCplusplus, iconClass: styles.blue }],
    [FileExtension.CSHARP, { icon: SiCsharp, iconClass: styles.blue }],
    [FileExtension.CSS, { icon: SiCssmodules, iconClass: styles.blue }],
    [FileExtension.DART, { icon: SiDart, iconClass: styles.blue }],
    [FileExtension.DOCKERIGNORE, { icon: SiDocker, iconClass: styles.blue }],
    [FileExtension.FORTRAN, { icon: SiFortran, iconClass: styles.defaultIcon }],
    [FileExtension.FSHARP, { icon: SiFsharp, iconClass: styles.blue }],
    [FileExtension.GITIGNORE, { icon: SiGit, iconClass: styles.red }],
    [FileExtension.GITATTRIBUTES, { icon: SiGit, iconClass: styles.red }],
    [FileExtension.GO, { icon: SiGo, iconClass: styles.blue }],
    [FileExtension.GRAPHQL, { icon: SiGraphql, iconClass: styles.pink }],
    [FileExtension.HASKELL, { icon: SiHaskell, iconClass: styles.blue }],
    [FileExtension.HTML, { icon: SiHtml5, iconClass: styles.blue }],
    [FileExtension.JAVA, { icon: FaJava, iconClass: styles.defaultIcon }],
    [FileExtension.JAVASCRIPT, { icon: SiJavascript, iconClass: styles.yellow }],
    [FileExtension.JSX, { icon: SiReact, iconClass: styles.yellow }],
    [FileExtension.JSON, { icon: VscJson, iconClass: styles.defaultIcon }],
    [FileExtension.JULIA, { icon: SiJulia, iconClass: styles.defaultIcon }],
    [FileExtension.KOTLIN, { icon: SiKotlin, iconClass: styles.green }],
    [FileExtension.LOCKFILE, { icon: VscJson, iconClass: styles.defaultIcon }],
    [FileExtension.LUA, { icon: SiLua, iconClass: styles.blue }],
    [FileExtension.MARKDOWN, { icon: SiMarkdown, iconClass: styles.blue }],
    [FileExtension.NCL, { icon: CiSettings, iconClass: styles.defaultIcon }],
    [FileExtension.NIX, { icon: SiNixos, iconClass: styles.gray }],
    [FileExtension.NPM, { icon: SiNpm, iconClass: styles.red }],
    [FileExtension.PHP, { icon: SiPhp, iconClass: styles.defaultIcon }],
    [FileExtension.PERL, { icon: SiPerl, iconClass: styles.defaultIcon }],
    [FileExtension.PYTHON, { icon: SiPython, iconClass: styles.yellow }],
    [FileExtension.R, { icon: SiR, iconClass: styles.red }],
    [FileExtension.RUBY, { icon: SiRuby, iconClass: styles.red }],
    [FileExtension.RUST, { icon: SiRust, iconClass: styles.defaultIcon }],
    [FileExtension.SCALA, { icon: SiScala, iconClass: styles.red }],
    [FileExtension.SASS, { icon: FaSass, iconClass: styles.pink }],
    [FileExtension.SQL, { icon: GoDatabase, iconClass: styles.blue }],
    [FileExtension.SVG, { icon: SiSvg, iconClass: styles.blue }],
    [FileExtension.SWIFT, { icon: SiSwift, iconClass: styles.blue }],
    [FileExtension.TYPESCRIPT, { icon: SiTypescript, iconClass: styles.blue }],
    [FileExtension.TSX, { icon: SiReact, iconClass: styles.blue }],
    [FileExtension.TEXT, { icon: CiTextAlignLeft, iconClass: styles.defaultIcon }],
    [FileExtension.YAML, { icon: CiSettings, iconClass: styles.defaultIcon }],
    [FileExtension.YML, { icon: CiSettings, iconClass: styles.defaultIcon }],
    [FileExtension.ZIG, { icon: SiZig, iconClass: styles.yellow }],
])

import {
    mdiCodeJson,
    mdiCog,
    mdiConsole,
    mdiDocker,
    mdiFileCodeOutline,
    mdiFileJpgBox,
    mdiFilePngBox,
    mdiGit,
    mdiGraphql,
    mdiLanguageC,
    mdiLanguageCpp,
    mdiLanguageCsharp,
    mdiLanguageCss3,
    mdiLanguageGo,
    mdiLanguageHaskell,
    mdiLanguageHtml5,
    mdiLanguageJava,
    mdiLanguageJavascript,
    mdiLanguageKotlin,
    mdiLanguageLua,
    mdiLanguageMarkdown,
    mdiLanguagePhp,
    mdiLanguagePython,
    mdiLanguageR,
    mdiLanguageRuby,
    mdiLanguageRust,
    mdiLanguageSwift,
    mdiLanguageTypescript,
    mdiNix,
    mdiNpm,
    mdiReact,
    mdiSass,
    mdiSvg,
    mdiText,
} from '@mdi/js'

import { containsTest } from '$lib/web'

const BLUE = 'var(--blue)'
const PINK = 'var(--pink)'
const YELLOW = 'var(--yellow)'
const RED = 'var(--red)'
const GREEN = 'var(--green)'
const CYAN = 'var(--blue)'
const DEFAULT_ICON = 'var(--gray-07)'

export enum FileExtension {
    ASSEMBLY = 'asm',
    BASH = 'sh',
    BASIC = 'vb',
    C = 'c',
    CLOJURE_CLJ = 'clj',
    CLOJURE_CLJS = 'cljs',
    CLOJURE_CLJR = 'cljr',
    CLOJURE_CLJC = 'cljc',
    CLOJURE_EDN = 'edn',
    COFFEE = 'coffee',
    CPP = 'cc',
    CSHARP = 'cs',
    CSS = 'css',
    DART = 'dart',
    DEFAULT = 'default',
    DOCKERFILE = 'Dockerfile',
    DOCKERIGNORE = 'dockerignore',
    FORTRAN_F = 'f',
    FORTRAN_FOR = 'for',
    FORTRAN_FTN = 'ftn',
    FSHARP = 'fs',
    FSI = 'fsi',
    FSX = 'fsx',
    GEMFILE = 'Gemfile',
    GIF = 'gif',
    GIFF = 'giff',
    GITIGNORE = 'gitignore',
    GITATTRIBUTES = 'gitattributes',
    GO = 'go',
    GOMOD = 'mod',
    GOSUM = 'sum',
    GRAPHQL = 'graphql',
    GROOVY = 'groovy',
    HASKELL = 'hs',
    HTML = 'html',
    JAVA = 'java',
    JAVASCRIPT = 'js',
    JPG = 'jpg',
    JPEG = 'jpeg',
    JSON = 'json',
    JSX = 'jsx',
    JULIA = 'jl',
    KOTLIN = 'kt',
    LOCKFILE = 'lock',
    LUA = 'lua',
    MARKDOWN = 'md',
    MDX = 'mdx',
    NCL = 'ncl',
    NIX = 'nix',
    NPM = 'npmrc',
    OCAML = 'ml',
    PHP = 'php',
    PERL = 'pl',
    PERL_PM = 'pm',
    PNG = 'png',
    POWERSHELL_PS1 = 'ps1',
    POWERSHELL_PSM1 = 'psm1',
    PYTHON = 'py',
    R = 'r',
    R_CAP = 'R',
    RUBY = 'rb',
    RUST = 'rs',
    SCALA = 'scala',
    SASS = 'scss',
    SQL = 'sql',
    SVELTE = 'svelte',
    SVG = 'svg',
    SWIFT = 'swift',
    TEST = 'test',
    TOML = 'toml',
    TYPESCRIPT = 'ts',
    TSX = 'tsx',
    TEXT = 'txt',
    YAML = 'yaml',
    YML = 'yml',
    ZIG = 'zig',
}

const FILE_ICONS = new Map(Object.values(FileExtension).map(extension => [extension, getIconAttributes(extension)]))

interface IconInfo {
    svgPath: string
    color: string
}

interface FileInfo {
    icon: IconInfo
    isTest: boolean
}

const getColor = (extension: FileExtension) => {
    switch (extension) {
        case FileExtension.MARKDOWN:
        case FileExtension.MDX:
        case FileExtension.SVG: {
            return BLUE
        }
        case FileExtension.GITIGNORE:
        case FileExtension.GITATTRIBUTES:
        case FileExtension.HTML:
        case FileExtension.NPM:
        case FileExtension.R:
        case FileExtension.R_CAP:
        case FileExtension.RUBY:
        case FileExtension.GEMFILE: {
            return RED
        }
        case FileExtension.GOMOD:
        case FileExtension.GOSUM:
        case FileExtension.GRAPHQL:
        case FileExtension.SASS: {
            return PINK
        }
        case (FileExtension.JAVASCRIPT, FileExtension.PYTHON): {
            return YELLOW
        }
        case FileExtension.KOTLIN: {
            return GREEN
        }
        case FileExtension.PHP: {
            return CYAN
        }
        default: {
            return DEFAULT_ICON
        }
    }
}

// TODO: Explore other icon libraries or make our own.
// Most programming language Material Design icons will be
// deprecated very soon, and will not be included in the next release.
// Additionally, it doesn't have all the icons we need anyway.
const getIconAttributes = (extension: FileExtension): IconInfo => {
    const color = getColor(extension)
    switch (extension) {
        // TODO: Ideally, we would switch to a new icon library
        // so we don't have to default to a general icon
        // for the languages in the first case block.
        case FileExtension.ASSEMBLY:
        case FileExtension.BASIC:
        case FileExtension.CLOJURE_CLJ:
        case FileExtension.CLOJURE_CLJC:
        case FileExtension.CLOJURE_CLJR:
        case FileExtension.CLOJURE_EDN:
        case FileExtension.CLOJURE_CLJS:
        case FileExtension.DART:
        case FileExtension.DEFAULT:
        case FileExtension.FORTRAN_F:
        case FileExtension.FORTRAN_FOR:
        case FileExtension.FORTRAN_FTN:
        case FileExtension.FSHARP:
        case FileExtension.FSI:
        case FileExtension.FSX:
        case FileExtension.GROOVY:
        case FileExtension.JULIA:
        case FileExtension.OCAML:
        case FileExtension.PERL:
        case FileExtension.PERL_PM:
        case FileExtension.SCALA:
        case FileExtension.SVELTE:
        case FileExtension.SQL:
        case FileExtension.TEST:
        case FileExtension.ZIG: {
            return { svgPath: mdiFileCodeOutline, color }
        }
        case FileExtension.BASH:
        case FileExtension.POWERSHELL_PS1:
        case FileExtension.POWERSHELL_PSM1: {
            return { svgPath: mdiConsole, color }
        }
        case FileExtension.C: {
            return { svgPath: mdiLanguageC, color }
        }
        case FileExtension.RUST:
        case FileExtension.TOML: {
            return { svgPath: mdiLanguageRust, color }
        }
        case FileExtension.CPP: {
            return { svgPath: mdiLanguageCpp, color }
        }
        case FileExtension.CSHARP: {
            return { svgPath: mdiLanguageCsharp, color }
        }
        case FileExtension.CSS: {
            return { svgPath: mdiLanguageCss3, color }
        }
        case FileExtension.DOCKERFILE:
        case FileExtension.DOCKERIGNORE: {
            return { svgPath: mdiDocker, color }
        }
        case FileExtension.GITIGNORE:
        case FileExtension.GITATTRIBUTES: {
            return { svgPath: mdiGit, color }
        }
        case FileExtension.GO: {
            return { svgPath: mdiLanguageGo, color }
        }
        case FileExtension.GRAPHQL: {
            return { svgPath: mdiGraphql, color }
        }
        case FileExtension.HASKELL: {
            return { svgPath: mdiLanguageHaskell, color }
        }
        case FileExtension.HTML: {
            return { svgPath: mdiLanguageHtml5, color }
        }
        case FileExtension.JAVA: {
            return { svgPath: mdiLanguageJava, color }
        }
        case FileExtension.JAVASCRIPT: {
            return { svgPath: mdiLanguageJavascript, color }
        }
        case FileExtension.JPG:
        case FileExtension.JPEG: {
            return { svgPath: mdiFileJpgBox, color }
        }
        case FileExtension.JSON:
        case FileExtension.LOCKFILE: {
            return { svgPath: mdiCodeJson, color }
        }
        case FileExtension.JSX:
        case FileExtension.TSX: {
            return { svgPath: mdiReact, color }
        }
        case FileExtension.KOTLIN: {
            return { svgPath: mdiLanguageKotlin, color }
        }
        case FileExtension.LUA: {
            return { svgPath: mdiLanguageLua, color }
        }
        case FileExtension.MARKDOWN:
        case FileExtension.MDX: {
            return { svgPath: mdiLanguageMarkdown, color }
        }
        case FileExtension.NCL:
        case FileExtension.NIX: {
            return { svgPath: mdiNix, color }
        }
        case FileExtension.NPM: {
            return { svgPath: mdiNpm, color }
        }
        case FileExtension.PHP: {
            return { svgPath: mdiLanguagePhp, color }
        }
        case FileExtension.PNG: {
            return { svgPath: mdiFilePngBox, color }
        }
        case FileExtension.PYTHON: {
            return { svgPath: mdiLanguagePython, color }
        }
        case FileExtension.R:
        case FileExtension.R_CAP: {
            return { svgPath: mdiLanguageR, color }
        }
        case FileExtension.RUBY:
        case FileExtension.GEMFILE: {
            return { svgPath: mdiLanguageRuby, color }
        }
        case FileExtension.SASS: {
            return { svgPath: mdiSass, color }
        }
        case FileExtension.SVG: {
            return { svgPath: mdiSvg, color }
        }
        case FileExtension.SWIFT: {
            return { svgPath: mdiLanguageSwift, color }
        }
        case FileExtension.TYPESCRIPT: {
            return { svgPath: mdiLanguageTypescript, color }
        }
        case FileExtension.TEXT: {
            return { svgPath: mdiText, color }
        }
        case FileExtension.YAML:
        case FileExtension.YML: {
            return { svgPath: mdiCog, color }
        }
        default: {
            return { svgPath: mdiConsole, color }
        }
    }
}

const getExtension = (file: string): FileExtension => {
    return file.split('.').at(-1) as FileExtension
}

export const getFileInfo = (file: string): FileInfo => {
    const extension = getExtension(file)
    const icon = FILE_ICONS.get(extension)

    if (icon) {
        return {
            icon,
            isTest: containsTest(file),
        }
    }

    return {
        icon: {
            svgPath: mdiFileCodeOutline,
            color: DEFAULT_ICON,
        },
        isTest: false,
    }
}

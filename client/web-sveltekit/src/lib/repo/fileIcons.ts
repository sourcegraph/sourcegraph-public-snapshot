import { containsTest } from "$lib/web"
import {
    mdiBash,
    mdiCodeJson,
    mdiDocker,
    mdiFileCode,
    mdiFileGifBox,
    mdiFileJpgBox,
    mdiFilePngBox,
    mdiGit,
    mdiGraphql,
    mdiLanguageC,
    mdiLanguageCpp,
    mdiLanguageCsharp,
    mdiLanguageCss3,
    mdiLanguageFortran,
    mdiLanguageGo,
    mdiLanguageHaskell,
    mdiLanguageHtml5,
    mdiLanguageJava,
    mdiLanguageJavascript,
    mdiLanguageKotlin,
    mdiLanguageLua,
    mdiLanguageMarkdown,
    mdiLanguageMarkdownOutline,
    mdiLanguagePhp,
    mdiLanguagePython,
    mdiLanguageR,
    mdiLanguageRuby,
    mdiLanguageRust,
    mdiLanguageSwift,
    mdiLanguageTypescript,
    mdiNpm,
    mdiReact,
    mdiSvg,
    mdiText,
} from '@mdi/js'

// Icon Colors
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
    GEMFILE = 'Gemfile',
    RUST = 'rs',
    SCALA = 'scala',
    SASS = 'scss',
    SQL = 'sql',
    SVELTE = 'svelte',
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

const getIcon = (extension: string): string => {
    switch (extension) {
        case FileExtension.ASSEMBLY: {
            return mdiFileCode
        }
        case FileExtension.BASH: {
            return mdiBash
        }
        case FileExtension.BASIC: {
            return mdiFileCode
        }
        case FileExtension.C: {
            return mdiLanguageC
        }
        case FileExtension.CLOJURE_CLJ: {
            return mdiFileCode
        }
        case FileExtension.CLOJURE_CLJC: {
            return mdiFileCode
        }
        case FileExtension.CLOJURE_CLJR: {
            return mdiFileCode
        }
        case FileExtension.CLOJURE_CLJS: {
            return mdiFileCode
        }
        case FileExtension.CLOJURE_EDN: {
            return mdiFileCode
        }
        case FileExtension.CPP: {
            return mdiLanguageCpp
        }
        case FileExtension.CSHARP: {
            return mdiLanguageCsharp
        }
        case FileExtension.CSS: {
            return mdiLanguageCss3
        }
        case FileExtension.DART: {
            return mdiFileCode
        }
        case FileExtension.DEFAULT: {
            return mdiFileCode
        }
        case FileExtension.DOCKERFILE: {
            return mdiDocker
        }
        case FileExtension.DOCKERIGNORE: {
            return mdiDocker
        }
        case FileExtension.FORTRAN_F: {
            return mdiLanguageFortran
        }
        case FileExtension.FORTRAN_FOR: {
            return mdiLanguageFortran
        }
        case FileExtension.FORTRAN_FTN: {
            return mdiLanguageFortran
        }
        case FileExtension.FSHARP: {
            return mdiFileCode
        }
        case FileExtension.FSI: {
            return mdiFileCode
        }
        case FileExtension.FSX: {
            return mdiFileCode
        }
        case FileExtension.GIF: {
            return mdiFileGifBox
        }
        case FileExtension.GIFF: {
            return mdiFileGifBox
        }
        case FileExtension.GITIGNORE: {
            return mdiGit
        }
        case FileExtension.GITATTRIBUTES: {
            return mdiGit
        }
        case FileExtension.GO: {
            return mdiLanguageGo
        }
        case FileExtension.GOMOD: {
            return mdiLanguageGo
        }
        case FileExtension.GOSUM: {
            return mdiLanguageGo
        }
        case FileExtension.GRAPHQL: {
            return mdiGraphql
        }
        case FileExtension.GROOVY: {
            return mdiFileCode
        }
        case FileExtension.HASKELL: {
            return mdiLanguageHaskell
        }
        case FileExtension.HTML: {
            return mdiLanguageHtml5
        }
        case FileExtension.JAVA: {
            return mdiLanguageJava
        }
        case FileExtension.JAVASCRIPT: {
            return mdiLanguageJavascript
        }
        case FileExtension.JPG: {
            return mdiFileJpgBox
        }
        case FileExtension.JPEG: {
            return mdiFileJpgBox
        }
        case FileExtension.JSON: {
            return mdiCodeJson
        }
        case FileExtension.JSX: {
            return mdiReact
        }
        case FileExtension.JULIA: {
            return mdiFileCode
        }
        case FileExtension.KOTLIN: {
            return mdiLanguageKotlin
        }
        case FileExtension.LOCKFILE: {
            return mdiCodeJson
        }
        case FileExtension.LUA: {
            return mdiLanguageLua
        }
        case FileExtension.MARKDOWN: {
            return mdiLanguageMarkdown
        }
        case FileExtension.MDX: {
            return mdiLanguageMarkdownOutline
        }
        case FileExtension.NCL: {
            return mdiFileCode
        }
        case FileExtension.NIX: {
            return mdiFileCode
        }
        case FileExtension.NPM: {
            return mdiNpm
        }
        case FileExtension.OCAML: {
            return mdiFileCode
        }
        case FileExtension.PHP: {
            return mdiLanguagePhp
        }
        case FileExtension.PERL: {
            return mdiFileCode
        }
        case FileExtension.PERL_PM: {
            return mdiFileCode
        }
        case FileExtension.PNG: {
            return mdiFilePngBox
        }
        case FileExtension.POWERSHELL_PS1: {
            return mdiBash
        }
        case FileExtension.POWERSHELL_PSM1: {
            return mdiBash
        }
        case FileExtension.PYTHON: {
            return mdiLanguagePython
        }
        case FileExtension.R: {
            return mdiLanguageR
        }
        case FileExtension.R_CAP: {
            return mdiLanguageR
        }
        case FileExtension.RUBY: {
            return mdiLanguageRuby
        }
        case FileExtension.GEMFILE: {
            return mdiLanguageRuby
        }
        case FileExtension.RUST: {
            return mdiLanguageRust
        }
        case FileExtension.SCALA: {
            return mdiFileCode
        }
        case FileExtension.SASS: {
            return mdiLanguageCss3
        }
        case FileExtension.SQL: {
            return mdiFileCode
        }
        case FileExtension.SVELTE: {
            return mdiFileCode
        }
        case FileExtension.SVG: {
            return mdiSvg
        }
        case FileExtension.SWIFT: {
            return mdiLanguageSwift
        }
        case FileExtension.TEST: {
            return mdiFileCode
        }
        case FileExtension.TYPESCRIPT: {
            return mdiLanguageTypescript
        }
        case FileExtension.TSX: {
            return mdiReact
        }
        case FileExtension.TEXT: {
            return mdiText
        }
        case FileExtension.YAML: {
            return mdiCodeJson
        }
        case FileExtension.YML: {
            return mdiCodeJson
        }
        case FileExtension.ZIG: {
            return mdiFileCode
        }
        default: {
            return mdiFileCode
        }
    }
}

const getColor = (extension: string): string => {
    switch (extension) {
        case FileExtension.ASSEMBLY: {
            return DEFAULT_ICON
        }
        case FileExtension.BASH: {
            return DEFAULT_ICON
        }
        case FileExtension.BASIC: {
            return DEFAULT_ICON
        }
        case FileExtension.C: {
            return BLUE
        }
        case FileExtension.CLOJURE_CLJ: {
            return DEFAULT_ICON
        }
        case FileExtension.CLOJURE_CLJC: {
            return DEFAULT_ICON
        }
        case FileExtension.CLOJURE_CLJR: {
            return DEFAULT_ICON
        }
        case FileExtension.CLOJURE_CLJS: {
            return DEFAULT_ICON
        }
        case FileExtension.CLOJURE_EDN: {
            return DEFAULT_ICON
        }
        case FileExtension.CPP: {
            return BLUE
        }
        case FileExtension.CSHARP: {
            return BLUE
        }
        case FileExtension.CSS: {
            return BLUE
        }
        case FileExtension.DART: {
            return BLUE
        }
        case FileExtension.DEFAULT: {
            return DEFAULT_ICON
        }
        case FileExtension.DOCKERFILE: {
            return BLUE
        }
        case FileExtension.DOCKERIGNORE: {
            return BLUE
        }
        case FileExtension.FORTRAN_F: {
            return DEFAULT_ICON
        }
        case FileExtension.FORTRAN_FOR: {
            return DEFAULT_ICON
        }
        case FileExtension.FORTRAN_FTN: {
            return DEFAULT_ICON
        }
        case FileExtension.FSHARP: {
            return DEFAULT_ICON
        }
        case FileExtension.FSI: {
            return DEFAULT_ICON
        }
        case FileExtension.FSX: {
            return DEFAULT_ICON
        }
        case FileExtension.GIF: {
            return DEFAULT_ICON
        }
        case FileExtension.GIFF: {
            return DEFAULT_ICON
        }
        case FileExtension.GITIGNORE: {
            return RED
        }
        case FileExtension.GITATTRIBUTES: {
            return RED
        }
        case FileExtension.GO: {
            return BLUE
        }
        case FileExtension.GOMOD: {
            return PINK
        }
        case FileExtension.GOSUM: {
            return PINK
        }
        case FileExtension.GROOVY: {
            return BLUE
        }
        case FileExtension.GRAPHQL: {
            return PINK
        }
        case FileExtension.HASKELL: {
            return BLUE
        }
        case FileExtension.HTML: {
            return BLUE
        }
        case FileExtension.JAVA: {
            return DEFAULT_ICON
        }
        case FileExtension.JAVASCRIPT: {
            return YELLOW
        }
        case FileExtension.JPG: {
            return DEFAULT_ICON
        }
        case FileExtension.JPEG: {
            return DEFAULT_ICON
        }
        case FileExtension.JSON: {
            return DEFAULT_ICON
        }
        case FileExtension.JSX: {
            return BLUE
        }
        case FileExtension.JULIA: {
            return DEFAULT_ICON
        }
        case FileExtension.KOTLIN: {
            return GREEN
        }
        case FileExtension.LOCKFILE: {
            return DEFAULT_ICON
        }
        case FileExtension.LUA: {
            return BLUE
        }
        case FileExtension.MARKDOWN: {
            return BLUE
        }
        case FileExtension.MDX: {
            return BLUE
        }
        case FileExtension.NCL: {
            return DEFAULT_ICON
        }
        case FileExtension.NIX: {
            return DEFAULT_ICON
        }
        case FileExtension.NPM: {
            return RED
        }
        case FileExtension.OCAML: {
            return DEFAULT_ICON
        }
        case FileExtension.PHP: {
            return CYAN
        }
        case FileExtension.PERL: {
            return DEFAULT_ICON
        }
        case FileExtension.PERL_PM: {
            return DEFAULT_ICON
        }
        case FileExtension.PNG: {
            return DEFAULT_ICON
        }
        case FileExtension.POWERSHELL_PS1: {
            return DEFAULT_ICON
        }
        case FileExtension.POWERSHELL_PSM1: {
            return DEFAULT_ICON
        }
        case FileExtension.PYTHON: {
            return YELLOW
        }
        case FileExtension.R: {
            return RED
        }
        case FileExtension.R_CAP: {
            return RED
        }
        case FileExtension.RUBY: {
            return RED
        }
        case FileExtension.GEMFILE: {
            return RED
        }
        case FileExtension.RUST: {
            return DEFAULT_ICON
        }
        case FileExtension.SCALA: {
            return DEFAULT_ICON
        }
        case FileExtension.SASS: {
            return PINK
        }
        case FileExtension.SQL: {
            return DEFAULT_ICON
        }
        case FileExtension.SVELTE: {
            return DEFAULT_ICON
        }
        case FileExtension.SVG: {
            return BLUE
        }
        case FileExtension.SWIFT: {
            return BLUE
        }
        case FileExtension.TEST: {
            return DEFAULT_ICON
        }
        case FileExtension.TYPESCRIPT: {
            return BLUE
        }
        case FileExtension.TSX: {
            return BLUE
        }
        case FileExtension.TEXT: {
            return DEFAULT_ICON
        }
        case FileExtension.YAML: {
            return DEFAULT_ICON
        }
        case FileExtension.YML: {
            return DEFAULT_ICON
        }
        case FileExtension.ZIG: {
            return DEFAULT_ICON
        }
        default: {
            return DEFAULT_ICON
        }
    }
}

export interface IconInfoSvelte {
    svgPath: string
    color: string
}

const getIconInfo = (extension: FileExtension) => {
    return {
        svgPath: getIcon(extension),
        color: getColor(extension),
    }
}

export const FILE_ICONS: Map<FileExtension, IconInfoSvelte> = new Map([
    [FileExtension.ASSEMBLY, { ...getIconInfo(FileExtension.ASSEMBLY) }],
    [FileExtension.BASH, { ...getIconInfo(FileExtension.BASH) }],
    [FileExtension.BASIC, { ...getIconInfo(FileExtension.BASIC) }],
    [FileExtension.C, { ...getIconInfo(FileExtension.C) }],
    [FileExtension.CLOJURE_CLJ, { ...getIconInfo(FileExtension.CLOJURE_CLJ) }],
    [FileExtension.CLOJURE_CLJC, { ...getIconInfo(FileExtension.CLOJURE_CLJC) }],
    [FileExtension.CLOJURE_CLJR, { ...getIconInfo(FileExtension.CLOJURE_CLJR) }],
    [FileExtension.CLOJURE_CLJS, { ...getIconInfo(FileExtension.CLOJURE_CLJS) }],
    [FileExtension.CLOJURE_EDN, { ...getIconInfo(FileExtension.CLOJURE_EDN) }],
    [FileExtension.COFFEE, { ...getIconInfo(FileExtension.COFFEE) }],
    [FileExtension.CPP, { ...getIconInfo(FileExtension.CPP) }],
    [FileExtension.CSHARP, { ...getIconInfo(FileExtension.CSHARP) }],
    [FileExtension.CSS, { ...getIconInfo(FileExtension.CSS) }],
    [FileExtension.DART, { ...getIconInfo(FileExtension.DART) }],
    [FileExtension.DEFAULT, { ...getIconInfo(FileExtension.DEFAULT) }],
    [FileExtension.DOCKERFILE, { ...getIconInfo(FileExtension.DOCKERFILE) }],
    [FileExtension.DOCKERIGNORE, { ...getIconInfo(FileExtension.DOCKERIGNORE) }],
    [FileExtension.FORTRAN_F, { ...getIconInfo(FileExtension.FORTRAN_F) }],
    [FileExtension.FORTRAN_FOR, { ...getIconInfo(FileExtension.FORTRAN_FOR) }],
    [FileExtension.FORTRAN_FTN, { ...getIconInfo(FileExtension.FORTRAN_FTN) }],
    [FileExtension.FSHARP, { ...getIconInfo(FileExtension.FSHARP) }],
    [FileExtension.FSI, { ...getIconInfo(FileExtension.FSI) }],
    [FileExtension.FSX, { ...getIconInfo(FileExtension.FSX) }],
    [FileExtension.GIF, { ...getIconInfo(FileExtension.GIF) }],
    [FileExtension.GIFF, { ...getIconInfo(FileExtension.GIFF) }],
    [FileExtension.GITIGNORE, { ...getIconInfo(FileExtension.GITIGNORE) }],
    [FileExtension.GITATTRIBUTES, { ...getIconInfo(FileExtension.GITATTRIBUTES) }],
    [FileExtension.GO, { ...getIconInfo(FileExtension.GO) }],
    [FileExtension.GOMOD, { ...getIconInfo(FileExtension.GOMOD) }],
    [FileExtension.GOSUM, { ...getIconInfo(FileExtension.GOSUM) }],
    [FileExtension.GROOVY, { ...getIconInfo(FileExtension.GROOVY) }],
    [FileExtension.GRAPHQL, { ...getIconInfo(FileExtension.GRAPHQL) }],
    [FileExtension.HASKELL, { ...getIconInfo(FileExtension.HASKELL) }],
    [FileExtension.HTML, { ...getIconInfo(FileExtension.HTML) }],
    [FileExtension.JAVA, { ...getIconInfo(FileExtension.JAVA) }],
    [FileExtension.JAVASCRIPT, { ...getIconInfo(FileExtension.JAVASCRIPT) }],
    [FileExtension.JPG, { ...getIconInfo(FileExtension.JPG) }],
    [FileExtension.JPEG, { ...getIconInfo(FileExtension.JPEG) }],
    [FileExtension.JSX, { ...getIconInfo(FileExtension.JSX) }],
    [FileExtension.JSON, { ...getIconInfo(FileExtension.JSON) }],
    [FileExtension.JULIA, { ...getIconInfo(FileExtension.JULIA) }],
    [FileExtension.KOTLIN, { ...getIconInfo(FileExtension.KOTLIN) }],
    [FileExtension.LOCKFILE, { ...getIconInfo(FileExtension.LOCKFILE) }],
    [FileExtension.LUA, { ...getIconInfo(FileExtension.LUA) }],
    [FileExtension.MARKDOWN, { ...getIconInfo(FileExtension.MARKDOWN) }],
    [FileExtension.MDX, { ...getIconInfo(FileExtension.MDX) }],
    [FileExtension.NCL, { ...getIconInfo(FileExtension.NCL) }],
    [FileExtension.NIX, { ...getIconInfo(FileExtension.NIX) }],
    [FileExtension.NPM, { ...getIconInfo(FileExtension.NPM) }],
    [FileExtension.OCAML, { ...getIconInfo(FileExtension.OCAML) }],
    [FileExtension.PHP, { ...getIconInfo(FileExtension.PHP) }],
    [FileExtension.PERL, { ...getIconInfo(FileExtension.PERL) }],
    [FileExtension.PERL_PM, { ...getIconInfo(FileExtension.PERL_PM) }],
    [FileExtension.PNG, { ...getIconInfo(FileExtension.PNG) }],
    [FileExtension.POWERSHELL_PS1, { ...getIconInfo(FileExtension.POWERSHELL_PS1) }],
    [FileExtension.POWERSHELL_PSM1, { ...getIconInfo(FileExtension.POWERSHELL_PSM1) }],
    [FileExtension.PYTHON, { ...getIconInfo(FileExtension.PYTHON) }],
    [FileExtension.R, { ...getIconInfo(FileExtension.R) }],
    [FileExtension.R_CAP, { ...getIconInfo(FileExtension.R_CAP) }],
    [FileExtension.RUBY, { ...getIconInfo(FileExtension.RUBY) }],
    [FileExtension.GEMFILE, { ...getIconInfo(FileExtension.GEMFILE) }],
    [FileExtension.RUST, { ...getIconInfo(FileExtension.RUST) }],
    [FileExtension.SCALA, { ...getIconInfo(FileExtension.SCALA) }],
    [FileExtension.SASS, { ...getIconInfo(FileExtension.SASS) }],
    [FileExtension.SQL, { ...getIconInfo(FileExtension.SQL) }],
    [FileExtension.SVELTE, { ...getIconInfo(FileExtension.SVELTE) }],
    [FileExtension.SVG, { ...getIconInfo(FileExtension.SVG) }],
    [FileExtension.SWIFT, { ...getIconInfo(FileExtension.SWIFT) }],
    [FileExtension.TYPESCRIPT, { ...getIconInfo(FileExtension.TYPESCRIPT) }],
    [FileExtension.TSX, { ...getIconInfo(FileExtension.TSX) }],
    [FileExtension.TEXT, { ...getIconInfo(FileExtension.TEXT) }],
    [FileExtension.YAML, { ...getIconInfo(FileExtension.YAML) }],
    [FileExtension.YML, { ...getIconInfo(FileExtension.YML) }],
    [FileExtension.ZIG, { ...getIconInfo(FileExtension.ZIG) }],
])

interface FileInfo {
    extension: FileExtension
    isTest: boolean
}

export const getFileInfo = (file: string): FileInfo => {
    const extension = file.split('.').at(-1)
    const isValidExtension = Object.values(FileExtension).includes(extension as FileExtension)

    if (extension && isValidExtension) {
        return {
            extension: extension as FileExtension,
            isTest: containsTest(file),
        }
    }

    return {
        extension: 'default' as FileExtension,
        isTest: false,
    }
}

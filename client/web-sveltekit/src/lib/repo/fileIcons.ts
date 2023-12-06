import {
    mdiBash,
    mdiCodeJson,
    mdiFileCode,
    mdiFileJpgBox,
    mdiFilePngBox,
    mdiLanguageMarkdown,
    mdiLanguageMarkdownOutline,
    mdiNix,
    mdiSvg,
    mdiText,
} from '@mdi/js'

import { containsTest } from '$lib/web'

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

interface IconInfo {
    value: string
    kind: 'mdi' | 'devicon'
}

const getIcon = (extension: string): IconInfo => {
    switch (extension) {
        case FileExtension.ASSEMBLY: {
            return { value: mdiFileCode, kind: 'mdi' }
        }
        case FileExtension.BASH: {
            return { value: mdiBash, kind: 'mdi' }
        }
        case FileExtension.BASIC: {
            return { value: mdiFileCode, kind: 'mdi' }
        }
        case FileExtension.C: {
            return { value: 'c', kind: 'devicon' }
        }
        case FileExtension.CLOJURE_CLJ: {
            return { value: 'clojure', kind: 'devicon' }
        }
        case FileExtension.CLOJURE_CLJC: {
            return { value: 'clojure', kind: 'devicon' }
        }
        case FileExtension.CLOJURE_CLJR: {
            return { value: 'clojure', kind: 'devicon' }
        }
        case FileExtension.CLOJURE_CLJS: {
            return { value: 'clojurescript', kind: 'devicon' }
        }
        case FileExtension.CLOJURE_EDN: {
            return { value: 'clojure', kind: 'devicon' }
        }
        case FileExtension.CPP: {
            return { value: 'cplusplus', kind: 'devicon' }
        }
        case FileExtension.CSHARP: {
            return { value: 'csharp', kind: 'devicon' }
        }
        case FileExtension.CSS: {
            return { value: 'css3', kind: 'devicon' }
        }
        case FileExtension.DART: {
            return { value: 'dart', kind: 'devicon' }
        }
        case FileExtension.DEFAULT: {
            return { value: mdiFileCode, kind: 'mdi' }
        }
        case FileExtension.DOCKERFILE: {
            return { value: 'docker', kind: 'devicon' }
        }
        case FileExtension.DOCKERIGNORE: {
            return { value: 'docker', kind: 'devicon' }
        }
        case FileExtension.FORTRAN_F: {
            return { value: mdiFileCode, kind: 'mdi' }
        }
        case FileExtension.FORTRAN_FOR: {
            return { value: mdiFileCode, kind: 'mdi' }
        }
        case FileExtension.FORTRAN_FTN: {
            return { value: mdiFileCode, kind: 'mdi' }
        }
        case FileExtension.FSHARP: {
            return { value: 'fsharp', kind: 'devicon' }
        }
        case FileExtension.FSI: {
            return { value: 'fsharp', kind: 'devicon' }
        }
        case FileExtension.FSX: {
            return { value: 'fsharp', kind: 'devicon' }
        }
        case FileExtension.GITIGNORE: {
            return { value: 'git', kind: 'devicon' }
        }
        case FileExtension.GITATTRIBUTES: {
            return { value: 'git', kind: 'devicon' }
        }
        case FileExtension.GO: {
            return { value: 'go', kind: 'devicon' }
        }
        case FileExtension.GRAPHQL: {
            return { value: 'graphql', kind: 'devicon' }
        }
        case FileExtension.GROOVY: {
            return { value: 'groovy', kind: 'devicon' }
        }
        case FileExtension.HASKELL: {
            return { value: 'haskell', kind: 'devicon' }
        }
        case FileExtension.HTML: {
            return { value: 'html5', kind: 'devicon' }
        }
        case FileExtension.JAVA: {
            return { value: 'java', kind: 'devicon' }
        }
        case FileExtension.JAVASCRIPT: {
            return { value: 'javascript', kind: 'devicon' }
        }
        case FileExtension.JPG: {
            return { value: mdiFileJpgBox, kind: 'mdi' }
        }
        case FileExtension.JPEG: {
            return { value: mdiFileJpgBox, kind: 'mdi' }
        }
        case FileExtension.JSON: {
            return { value: mdiCodeJson, kind: 'mdi' }
        }
        case FileExtension.JSX: {
            return { value: 'react', kind: 'devicon' }
        }
        case FileExtension.JULIA: {
            return { value: 'julia', kind: 'devicon' }
        }
        case FileExtension.KOTLIN: {
            return { value: 'kotlin', kind: 'devicon' }
        }
        case FileExtension.LOCKFILE: {
            return { value: mdiCodeJson, kind: 'mdi' }
        }
        case FileExtension.LUA: {
            return { value: 'lua', kind: 'devicon' }
        }
        case FileExtension.MARKDOWN: {
            return { value: mdiLanguageMarkdownOutline, kind: 'mdi' }
        }
        case FileExtension.MDX: {
            return { value: mdiLanguageMarkdown, kind: 'mdi' }
        }
        case FileExtension.NCL: {
            return { value: mdiNix, kind: 'mdi' }
        }
        case FileExtension.NIX: {
            return { value: mdiNix, kind: 'mdi' }
        }
        case FileExtension.NPM: {
            return { value: 'npm', kind: 'devicon' }
        }
        case FileExtension.OCAML: {
            return { value: 'ocaml', kind: 'devicon' }
        }
        case FileExtension.PHP: {
            return { value: 'php', kind: 'devicon' }
        }
        case FileExtension.PERL: {
            return { value: 'perl', kind: 'devicon' }
        }
        case FileExtension.PERL_PM: {
            return { value: 'perl', kind: 'devicon' }
        }
        case FileExtension.PNG: {
            return { value: mdiFilePngBox, kind: 'mdi' }
        }
        case FileExtension.POWERSHELL_PS1: {
            return { value: mdiBash, kind: 'mdi' }
        }
        case FileExtension.POWERSHELL_PSM1: {
            return { value: mdiBash, kind: 'mdi' }
        }
        case FileExtension.PYTHON: {
            return { value: 'python', kind: 'devicon' }
        }
        case FileExtension.R: {
            return { value: 'r', kind: 'devicon' }
        }
        case FileExtension.R_CAP: {
            return { value: 'r', kind: 'devicon' }
        }
        case FileExtension.RUBY: {
            return { value: 'ruby', kind: 'devicon' }
        }
        case FileExtension.GEMFILE: {
            return { value: 'ruby', kind: 'devicon' }
        }
        case FileExtension.RUST: {
            return { value: 'rust', kind: 'devicon' }
        }
        case FileExtension.SCALA: {
            return { value: 'scala', kind: 'devicon' }
        }
        case FileExtension.SASS: {
            return { value: 'sass', kind: 'devicon' }
        }
        case FileExtension.SQL: {
            return { value: mdiFileCode, kind: 'mdi' }
        }
        case FileExtension.SVELTE: {
            return { value: 'svelte', kind: 'devicon' }
        }
        case FileExtension.SVG: {
            return { value: mdiSvg, kind: 'mdi' }
        }
        case FileExtension.SWIFT: {
            return { value: 'swift', kind: 'devicon' }
        }
        case FileExtension.TEST: {
            return { value: mdiFileCode, kind: 'mdi' }
        }
        case FileExtension.TYPESCRIPT: {
            return { value: 'typescript', kind: 'devicon' }
        }
        case FileExtension.TSX: {
            return { value: 'react', kind: 'devicon' }
        }
        case FileExtension.TEXT: {
            return { value: mdiText, kind: 'mdi' }
        }
        case FileExtension.YAML: {
            return { value: mdiCodeJson, kind: 'mdi' }
        }
        case FileExtension.YML: {
            return { value: mdiCodeJson, kind: 'mdi' }
        }
        case FileExtension.ZIG: {
            return { value: 'zig', kind: 'devicon' }
        }
        default: {
            return { value: mdiFileCode, kind: 'mdi' }
        }
    }
}

const getColor = (extension: string): string => {
    switch (extension) {
        case (FileExtension.C,
        FileExtension.CPP,
        FileExtension.CSHARP,
        FileExtension.CSS,
        FileExtension.DART,
        FileExtension.DOCKERFILE,
        FileExtension.DOCKERIGNORE,
        FileExtension.GO,
        FileExtension.GROOVY,
        FileExtension.HASKELL,
        FileExtension.JSX,
        FileExtension.LUA,
        FileExtension.MARKDOWN,
        FileExtension.MDX,
        FileExtension.SVG,
        FileExtension.SWIFT,
        FileExtension.TYPESCRIPT,
        FileExtension.TSX): {
            return BLUE
        }
        case (FileExtension.GITIGNORE,
        FileExtension.GITATTRIBUTES,
        FileExtension.HTML,
        FileExtension.NPM,
        FileExtension.R,
        FileExtension.R_CAP,
        FileExtension.RUBY,
        FileExtension.GEMFILE): {
            return RED
        }
        case (FileExtension.GOMOD, FileExtension.GOSUM, FileExtension.GRAPHQL, FileExtension.SASS): {
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

export interface IconInfoSvelte {
    info: {
        value: string
        kind: 'mdi' | 'devicon'
    }
    // color will be ignored if kind is "devicon"
    color: string
}

const getIconInfo = (extension: FileExtension): IconInfoSvelte => {
    return {
        info: getIcon(extension),
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

/* export const FILE_ICONS: Map<FileExtension, IconInfoSvelte> = new Map([
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
])*/

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

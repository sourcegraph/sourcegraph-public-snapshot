import {
    mdiCodeJson,
    mdiConsole,
    mdiCog,
    mdiFileCode,
    mdiFileJpgBox,
    mdiFilePngBox,
    mdiLanguageMarkdown,
    mdiNix,
    mdiSvg,
    mdiText,
} from '@mdi/js'

import { containsTest } from '$lib/web'

// Devicons do not use these as colors are built in
// Material Design Icons need a color
const BLUE = 'var(--blue)'
const PINK = 'var(--pink)'
const YELLOW = 'var(--yellow)'
const RED = 'var(--red)'
const GREEN = 'var(--green)'
const CYAN = 'var(--blue)'
const DEFAULT_ICON = 'var(--gray-07)'

// We use the file extension to determine icon value, kind, color
// In the two helper functions getIconAttributes() and getColor()
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
    TOML = 'toml',
    TYPESCRIPT = 'ts',
    TSX = 'tsx',
    TEXT = 'txt',
    YAML = 'yaml',
    YML = 'yml',
    ZIG = 'zig',
}

// We use DviClass to enumerate Devicon classes
// These values will be interpolated in the actual component:
//      <i class={`devicon-${DviClass.JAVA}-plain`}>
// Visit https://devicon.dev/ for correct class names
enum DviClass {
    C = 'c',
    CLOJURE = 'clojure',
    CLOJURE_SCRIPT = 'clojurescript',
    CPP = 'cplusplus',
    CSHARP = 'csharp',
    CSS = 'css3',
    DART = 'dart',
    DOCKER = 'docker',
    FSHARP = 'fsharp',
    GIT = 'git',
    GO = 'go',
    GRAPHQL = 'graphql',
    GROOVY = 'groovy',
    HASKELL = 'haskell',
    HTML5 = 'html5',
    JAVA = 'java',
    JAVASCRIPT = 'javascript',
    JSON = 'json',
    JULIA = 'julia',
    KOTLIN = 'kotlin',
    LUA = 'lua',
    NPM = 'npm',
    OCAML = 'ocaml',
    PERL = 'perl',
    PHP = 'php',
    PYTHON = 'python',
    R = 'r',
    REACT = 'react',
    RUBY = 'ruby',
    RUST = 'rust',
    SASS = 'sass',
    SCALA = 'scala',
    SVELTE = 'svelte',
    SWIFT = 'swift',
    TYPESCRIPT = 'typescript',
    VUE = 'vue',
    ZIG = 'zig',
}

// Add different values to the Kind if using a different library.
enum Kind {
    MDI = 'mdi',
    DEVICON = 'devicon',
}

interface FileInfo {
    extension: FileExtension
    // @TODO: render test files with test-indicator
    isTest: boolean
}

// getFileInfo extracts the file extension from the full file name
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

interface IconSvelte {
    info: IconInfo
    // color ignored if 'kind' is 'devicon'
    color: string
}

interface IconInfo {
    value: string
    kind: 'mdi' | 'devicon'
}

const getIconSvelte = (extension: FileExtension): IconSvelte => {
    return {
        info: getIconAttributes(extension),
        color: getColor(extension),
    }
}

const getIconAttributes = (extension: string): IconInfo => {
    switch (extension) {
        case FileExtension.ASSEMBLY:
        case FileExtension.BASIC:
        case FileExtension.DEFAULT:
        case FileExtension.FORTRAN_F:
        case FileExtension.FORTRAN_FOR:
        case FileExtension.FORTRAN_FTN:
        case FileExtension.SQL:
        case FileExtension.TEST: {
            return {
                value: mdiFileCode,
                kind: Kind.MDI,
            }
        }
        case FileExtension.BASH:
        case FileExtension.POWERSHELL_PS1:
        case FileExtension.POWERSHELL_PSM1: {
            return {
                value: mdiConsole,
                kind: Kind.MDI,
            }
        }
        case FileExtension.C: {
            return {
                value: DviClass.C,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.RUST:
        case FileExtension.TOML: {
            return {
                value: DviClass.RUST,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.CLOJURE_CLJ:
        case FileExtension.CLOJURE_CLJC:
        case FileExtension.CLOJURE_CLJR:
        case FileExtension.CLOJURE_EDN: {
            return {
                value: DviClass.CLOJURE,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.CLOJURE_CLJS: {
            return {
                value: DviClass.CLOJURE_SCRIPT,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.CPP: {
            return {
                value: DviClass.CPP,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.CSHARP: {
            return {
                value: DviClass.CSHARP,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.CSS: {
            return {
                value: DviClass.CSS,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.DART: {
            return {
                value: DviClass.DART,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.DOCKERFILE:
        case FileExtension.DOCKERIGNORE: {
            return {
                value: DviClass.DOCKER,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.FSHARP:
        case FileExtension.FSI:
        case FileExtension.FSX: {
            return {
                value: DviClass.FSHARP,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.GITIGNORE:
        case FileExtension.GITATTRIBUTES: {
            return {
                value: DviClass.GIT,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.GO: {
            return {
                value: DviClass.GO,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.GRAPHQL: {
            return {
                value: DviClass.GRAPHQL,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.GROOVY: {
            return {
                value: DviClass.GROOVY,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.HASKELL: {
            return {
                value: DviClass.HASKELL,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.HTML: {
            return {
                value: DviClass.HTML5,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.JAVA: {
            return {
                value: DviClass.JAVA,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.JAVASCRIPT: {
            return {
                value: DviClass.JAVASCRIPT,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.JPG:
        case FileExtension.JPEG: {
            return {
                value: mdiFileJpgBox,
                kind: Kind.MDI,
            }
        }
        case FileExtension.JSON:
        case FileExtension.LOCKFILE: {
            return {
                value: mdiCodeJson,
                kind: Kind.MDI,
            }
        }
        case FileExtension.JSX:
        case FileExtension.TSX: {
            return {
                value: DviClass.REACT,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.JULIA: {
            return {
                value: DviClass.JULIA,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.KOTLIN: {
            return {
                value: DviClass.KOTLIN,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.LUA: {
            return {
                value: DviClass.LUA,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.MARKDOWN:
        case FileExtension.MDX: {
            return {
                value: mdiLanguageMarkdown,
                kind: Kind.MDI,
            }
        }
        case FileExtension.NCL:
        case FileExtension.NIX: {
            return {
                value: mdiNix,
                kind: Kind.MDI,
            }
        }
        case FileExtension.NPM: {
            return {
                value: DviClass.NPM,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.OCAML: {
            return {
                value: DviClass.OCAML,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.PHP: {
            return {
                value: DviClass.PHP,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.PERL:
        case FileExtension.PERL_PM: {
            return {
                value: DviClass.PERL,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.PNG: {
            return {
                value: mdiFilePngBox,
                kind: Kind.MDI,
            }
        }
        case FileExtension.PYTHON: {
            return {
                value: DviClass.PYTHON,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.R:
        case FileExtension.R_CAP: {
            return {
                value: DviClass.R,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.RUBY:
        case FileExtension.GEMFILE: {
            return {
                value: DviClass.RUBY,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.SCALA: {
            return {
                value: DviClass.SCALA,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.SASS: {
            return {
                value: DviClass.SASS,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.SVELTE: {
            return {
                value: DviClass.SVELTE,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.SVG: {
            return {
                value: mdiSvg,
                kind: Kind.MDI,
            }
        }
        case FileExtension.SWIFT: {
            return {
                value: DviClass.SWIFT,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.TYPESCRIPT: {
            return {
                value: DviClass.TYPESCRIPT,
                kind: Kind.DEVICON,
            }
        }
        case FileExtension.TEXT: {
            return {
                value: mdiText,
                kind: Kind.MDI,
            }
        }
        case FileExtension.YAML:
        case FileExtension.YML: {
            return {
                value: mdiCog,
                kind: Kind.MDI,
            }
        }
        case FileExtension.ZIG: {
            return {
                value: DviClass.ZIG,
                kind: Kind.DEVICON,
            }
        }
        default: {
            return {
                value: mdiFileCode,
                kind: Kind.MDI,
            }
        }
    }
}

// only relevant for MDI icons
const getColor = (extension: string): string => {
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

export const FILE_ICONS: Map<FileExtension, IconSvelte> = new Map([
    [FileExtension.ASSEMBLY, { ...getIconSvelte(FileExtension.ASSEMBLY) }],
    [FileExtension.BASH, { ...getIconSvelte(FileExtension.BASH) }],
    [FileExtension.BASIC, { ...getIconSvelte(FileExtension.BASIC) }],
    [FileExtension.C, { ...getIconSvelte(FileExtension.C) }],
    [FileExtension.CLOJURE_CLJ, { ...getIconSvelte(FileExtension.CLOJURE_CLJ) }],
    [FileExtension.CLOJURE_CLJC, { ...getIconSvelte(FileExtension.CLOJURE_CLJC) }],
    [FileExtension.CLOJURE_CLJR, { ...getIconSvelte(FileExtension.CLOJURE_CLJR) }],
    [FileExtension.CLOJURE_CLJS, { ...getIconSvelte(FileExtension.CLOJURE_CLJS) }],
    [FileExtension.CLOJURE_EDN, { ...getIconSvelte(FileExtension.CLOJURE_EDN) }],
    [FileExtension.COFFEE, { ...getIconSvelte(FileExtension.COFFEE) }],
    [FileExtension.CPP, { ...getIconSvelte(FileExtension.CPP) }],
    [FileExtension.CSHARP, { ...getIconSvelte(FileExtension.CSHARP) }],
    [FileExtension.CSS, { ...getIconSvelte(FileExtension.CSS) }],
    [FileExtension.DART, { ...getIconSvelte(FileExtension.DART) }],
    [FileExtension.DEFAULT, { ...getIconSvelte(FileExtension.DEFAULT) }],
    [FileExtension.DOCKERFILE, { ...getIconSvelte(FileExtension.DOCKERFILE) }],
    [FileExtension.DOCKERIGNORE, { ...getIconSvelte(FileExtension.DOCKERIGNORE) }],
    [FileExtension.FORTRAN_F, { ...getIconSvelte(FileExtension.FORTRAN_F) }],
    [FileExtension.FORTRAN_FOR, { ...getIconSvelte(FileExtension.FORTRAN_FOR) }],
    [FileExtension.FORTRAN_FTN, { ...getIconSvelte(FileExtension.FORTRAN_FTN) }],
    [FileExtension.FSHARP, { ...getIconSvelte(FileExtension.FSHARP) }],
    [FileExtension.FSI, { ...getIconSvelte(FileExtension.FSI) }],
    [FileExtension.FSX, { ...getIconSvelte(FileExtension.FSX) }],
    [FileExtension.GIF, { ...getIconSvelte(FileExtension.GIF) }],
    [FileExtension.GIFF, { ...getIconSvelte(FileExtension.GIFF) }],
    [FileExtension.GITIGNORE, { ...getIconSvelte(FileExtension.GITIGNORE) }],
    [FileExtension.GITATTRIBUTES, { ...getIconSvelte(FileExtension.GITATTRIBUTES) }],
    [FileExtension.GO, { ...getIconSvelte(FileExtension.GO) }],
    [FileExtension.GOMOD, { ...getIconSvelte(FileExtension.GOMOD) }],
    [FileExtension.GOSUM, { ...getIconSvelte(FileExtension.GOSUM) }],
    [FileExtension.GROOVY, { ...getIconSvelte(FileExtension.GROOVY) }],
    [FileExtension.GRAPHQL, { ...getIconSvelte(FileExtension.GRAPHQL) }],
    [FileExtension.HASKELL, { ...getIconSvelte(FileExtension.HASKELL) }],
    [FileExtension.HTML, { ...getIconSvelte(FileExtension.HTML) }],
    [FileExtension.JAVA, { ...getIconSvelte(FileExtension.JAVA) }],
    [FileExtension.JAVASCRIPT, { ...getIconSvelte(FileExtension.JAVASCRIPT) }],
    [FileExtension.JPG, { ...getIconSvelte(FileExtension.JPG) }],
    [FileExtension.JPEG, { ...getIconSvelte(FileExtension.JPEG) }],
    [FileExtension.JSX, { ...getIconSvelte(FileExtension.JSX) }],
    [FileExtension.JSON, { ...getIconSvelte(FileExtension.JSON) }],
    [FileExtension.JULIA, { ...getIconSvelte(FileExtension.JULIA) }],
    [FileExtension.KOTLIN, { ...getIconSvelte(FileExtension.KOTLIN) }],
    [FileExtension.LOCKFILE, { ...getIconSvelte(FileExtension.LOCKFILE) }],
    [FileExtension.LUA, { ...getIconSvelte(FileExtension.LUA) }],
    [FileExtension.MARKDOWN, { ...getIconSvelte(FileExtension.MARKDOWN) }],
    [FileExtension.MDX, { ...getIconSvelte(FileExtension.MDX) }],
    [FileExtension.NCL, { ...getIconSvelte(FileExtension.NCL) }],
    [FileExtension.NIX, { ...getIconSvelte(FileExtension.NIX) }],
    [FileExtension.NPM, { ...getIconSvelte(FileExtension.NPM) }],
    [FileExtension.OCAML, { ...getIconSvelte(FileExtension.OCAML) }],
    [FileExtension.PHP, { ...getIconSvelte(FileExtension.PHP) }],
    [FileExtension.PERL, { ...getIconSvelte(FileExtension.PERL) }],
    [FileExtension.PERL_PM, { ...getIconSvelte(FileExtension.PERL_PM) }],
    [FileExtension.PNG, { ...getIconSvelte(FileExtension.PNG) }],
    [FileExtension.POWERSHELL_PS1, { ...getIconSvelte(FileExtension.POWERSHELL_PS1) }],
    [FileExtension.POWERSHELL_PSM1, { ...getIconSvelte(FileExtension.POWERSHELL_PSM1) }],
    [FileExtension.PYTHON, { ...getIconSvelte(FileExtension.PYTHON) }],
    [FileExtension.R, { ...getIconSvelte(FileExtension.R) }],
    [FileExtension.R_CAP, { ...getIconSvelte(FileExtension.R_CAP) }],
    [FileExtension.RUBY, { ...getIconSvelte(FileExtension.RUBY) }],
    [FileExtension.GEMFILE, { ...getIconSvelte(FileExtension.GEMFILE) }],
    [FileExtension.RUST, { ...getIconSvelte(FileExtension.RUST) }],
    [FileExtension.SCALA, { ...getIconSvelte(FileExtension.SCALA) }],
    [FileExtension.SASS, { ...getIconSvelte(FileExtension.SASS) }],
    [FileExtension.SQL, { ...getIconSvelte(FileExtension.SQL) }],
    [FileExtension.SVELTE, { ...getIconSvelte(FileExtension.SVELTE) }],
    [FileExtension.SVG, { ...getIconSvelte(FileExtension.SVG) }],
    [FileExtension.SWIFT, { ...getIconSvelte(FileExtension.SWIFT) }],
    [FileExtension.TYPESCRIPT, { ...getIconSvelte(FileExtension.TYPESCRIPT) }],
    [FileExtension.TSX, { ...getIconSvelte(FileExtension.TSX) }],
    [FileExtension.TEXT, { ...getIconSvelte(FileExtension.TEXT) }],
    [FileExtension.YAML, { ...getIconSvelte(FileExtension.YAML) }],
    [FileExtension.YML, { ...getIconSvelte(FileExtension.YML) }],
    [FileExtension.ZIG, { ...getIconSvelte(FileExtension.ZIG) }],
])

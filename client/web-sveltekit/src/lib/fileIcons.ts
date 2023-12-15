import {
    mdiCodeJson,
    mdiCog,
    mdiConsole,
    mdiDocker,
    mdiFileCodeOutline,
    mdiFileGifBox,
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
const GRAY = 'var(--gray-05)'

enum FileExtension {
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

interface IconInfo {
    svgPath: string
    color: string
}

interface FileInfo {
    icon: IconInfo
    isTest: boolean
}

function getColor(extension: FileExtension) {
    switch (extension) {
        case FileExtension.DOCKERFILE:
        case FileExtension.DOCKERIGNORE:
        case FileExtension.GO:
        case FileExtension.MARKDOWN:
        case FileExtension.MDX:
        case FileExtension.JSX:
        case FileExtension.TSX:
        case FileExtension.TYPESCRIPT:
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
        case FileExtension.JAVASCRIPT:
        case FileExtension.PYTHON: {
            return YELLOW
        }
        case FileExtension.KOTLIN: {
            return GREEN
        }
        case FileExtension.PHP: {
            return CYAN
        }
        default: {
            return GRAY
        }
    }
}

// TODO: Explore other icon libraries or make our own.
// Most programming language Material Design icons will be
// deprecated very soon, and will not be included in the next release.
// Additionally, it doesn't have all the icons we need anyway.
function getIcon(extension: FileExtension): string {
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
            return mdiFileCodeOutline
        }
        case FileExtension.BASH:
        case FileExtension.POWERSHELL_PS1:
        case FileExtension.POWERSHELL_PSM1: {
            return mdiConsole
        }
        case FileExtension.C: {
            return mdiLanguageC
        }
        case FileExtension.RUST:
        case FileExtension.TOML: {
            return mdiLanguageRust
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
        case FileExtension.DOCKERFILE:
        case FileExtension.DOCKERIGNORE: {
            return mdiDocker
        }
        case FileExtension.GIF:
        case FileExtension.GIFF: {
            return mdiFileGifBox
        }
        case FileExtension.GITIGNORE:
        case FileExtension.GITATTRIBUTES: {
            return mdiGit
        }
        case FileExtension.GO:
        case FileExtension.GOMOD:
        case FileExtension.GOSUM: {
            return mdiLanguageGo
        }
        case FileExtension.GRAPHQL: {
            return mdiGraphql
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
        case FileExtension.JPG:
        case FileExtension.JPEG: {
            return mdiFileJpgBox
        }
        case FileExtension.JSON:
        case FileExtension.LOCKFILE: {
            return mdiCodeJson
        }
        case FileExtension.JSX:
        case FileExtension.TSX: {
            return mdiReact
        }
        case FileExtension.KOTLIN: {
            return mdiLanguageKotlin
        }
        case FileExtension.LUA: {
            return mdiLanguageLua
        }
        case FileExtension.MARKDOWN:
        case FileExtension.MDX: {
            return mdiLanguageMarkdown
        }
        case FileExtension.NCL:
        case FileExtension.NIX: {
            return mdiNix
        }
        case FileExtension.NPM: {
            return mdiNpm
        }
        case FileExtension.PHP: {
            return mdiLanguagePhp
        }
        case FileExtension.PNG: {
            return mdiFilePngBox
        }
        case FileExtension.PYTHON: {
            return mdiLanguagePython
        }
        case FileExtension.R:
        case FileExtension.R_CAP: {
            return mdiLanguageR
        }
        case FileExtension.RUBY:
        case FileExtension.GEMFILE: {
            return mdiLanguageRuby
        }
        case FileExtension.SASS: {
            return mdiSass
        }
        case FileExtension.SVG: {
            return mdiSvg
        }
        case FileExtension.SWIFT: {
            return mdiLanguageSwift
        }
        case FileExtension.TYPESCRIPT: {
            return mdiLanguageTypescript
        }
        case FileExtension.TEXT: {
            return mdiText
        }
        case FileExtension.YAML:
        case FileExtension.YML: {
            return mdiCog
        }
        default: {
            return mdiFileCodeOutline
        }
    }
}

export function getFileInfo(file: string): FileInfo {
    const extension = file.split('.').at(-1)
    const icon = FILE_ICONS.get(extension as FileExtension)

    if (icon) {
        return {
            icon,
            isTest: containsTest(file),
        }
    }

    return {
        icon: {
            svgPath: mdiFileCodeOutline,
            color: GRAY,
        },
        isTest: false,
    }
}

const FILE_ICONS = new Map(
    Object.values(FileExtension).map(extension => [
        extension,
        { svgPath: getIcon(extension), color: getColor(extension) },
    ])
)

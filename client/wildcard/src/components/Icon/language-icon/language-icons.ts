import type { ComponentType } from 'react'

import { CiSettings, CiTextAlignLeft } from 'react-icons/ci'
import { FaCss3Alt, FaSass, FaVuejs } from 'react-icons/fa'
import { GoDatabase, GoTerminal } from 'react-icons/go'
import { GrJava } from 'react-icons/gr'
import { ImEarth } from 'react-icons/im'
import { MdGif } from 'react-icons/md'
import { PiFilePngLight } from 'react-icons/pi'
import {
    SiApachegroovy,
    SiC,
    SiClojure,
    SiCmake,
    SiCoffeescript,
    SiCplusplus,
    SiCrystal,
    SiCsharp,
    SiD,
    SiDart,
    SiDocker,
    SiEditorconfig,
    SiElixir,
    SiElm,
    SiErlang,
    SiFortran,
    SiFsharp,
    SiGit,
    SiGnuemacs,
    SiGo,
    SiGraphql,
    SiHaskell,
    SiHtml5,
    SiJavascript,
    SiJinja,
    SiJpeg,
    SiJulia,
    SiKotlin,
    SiLlvm,
    SiLua,
    SiMarkdown,
    SiNginx,
    SiNim,
    SiNixos,
    SiNpm,
    SiOcaml,
    SiPerl,
    SiPhp,
    SiPurescript,
    SiPython,
    SiR,
    SiRuby,
    SiRust,
    SiScala,
    SiSvelte,
    SiSvg,
    SiSwift,
    SiTerraform,
    SiToml,
    SiTypescript,
    SiUnrealengine,
    SiVim,
    SiVisualbasic,
    SiWebassembly,
    SiWolframmathematica,
    SiZig,
} from 'react-icons/si'
import { VscJson } from 'react-icons/vsc'

import styles from './LanguageIcon.module.scss'

export type CustomIcon = ComponentType<{ className?: string }>

export interface IconInfo {
    icon: CustomIcon
    className: string
}

/**
 * The keys of this map must be present in the list of `languageFilter.ALL_LANGUAGES`.
 *
 * This map is deliberately not public, use {@link getFileIconInfo} instead.
 *
 * See FIXME(id: language-detection) for context on why this map uses the
 * language as the key instead of something simpler like a file extension.
 */
export const FILE_ICONS_BY_LANGUAGE: Map<string, IconInfo> = new Map([
    ['Bash', { icon: GoTerminal, className: styles.defaultIcon }],
    ['BASIC', { icon: SiVisualbasic, className: styles.defaultIcon }],
    ['C', { icon: SiC, className: styles.blue }],
    ['C++', { icon: SiCplusplus, className: styles.blue }],
    ['C#', { icon: SiCsharp, className: styles.blue }],
    ['Clojure', { icon: SiClojure, className: styles.blue }],
    ['CMake', { icon: SiCmake, className: styles.defaultIcon }],
    ['CoffeeScript', { icon: SiCoffeescript, className: styles.defaultIcon }],

    // TODO: Decide icon for CSV?
    ['Crystal', { icon: SiCrystal, className: styles.blue }],
    ['CSS', { icon: FaCss3Alt, className: styles.blue }],
    ['D', { icon: SiD, className: styles.red }],
    ['Dart', { icon: SiDart, className: styles.blue }],
    ['Dockerfile', { icon: SiDocker, className: styles.blue }],
    ['EditorConfig', { icon: SiEditorconfig, className: styles.defaultIcon }],
    ['Elixir', { icon: SiElixir, className: styles.blue }],
    ['Elm', { icon: SiElm, className: styles.blue }],
    ['Emacs Lisp', { icon: SiGnuemacs, className: styles.defaultIcon }],
    ['Erlang', { icon: SiErlang, className: styles.blue }],
    ['Fortran', { icon: SiFortran, className: styles.defaultIcon }],
    ['Fortran Free Form', { icon: SiFortran, className: styles.defaultIcon }],
    ['F#', { icon: SiFsharp, className: styles.blue }],
    ['Git Attributes', { icon: SiGit, className: styles.red }],
    ['Go', { icon: SiGo, className: styles.blue }],
    ['Go Module', { icon: SiGo, className: styles.pink }],
    ['Go Checksums', { icon: SiGo, className: styles.pink }],
    ['Groovy', { icon: SiApachegroovy, className: styles.blue }],
    ['GraphQL', { icon: SiGraphql, className: styles.pink }],
    ['Haskell', { icon: SiHaskell, className: styles.blue }],
    ['HTML', { icon: SiHtml5, className: styles.blue }],
    ['HTML+ECR', { icon: SiCrystal, className: styles.blue }],
    ['HTML+EEX', { icon: SiElixir, className: styles.blue }],
    ['HTML+ERB', { icon: SiRuby, className: styles.blue }],
    ['HTML+PHP', { icon: SiPhp, className: styles.blue }],
    ['HTML+Razor', { icon: SiCsharp, className: styles.blue }],
    ['Ignore List', { icon: CiSettings, className: styles.defaultIcon }],
    ['Java', { icon: GrJava, className: styles.defaultIcon }],
    ['JavaScript', { icon: SiJavascript, className: styles.yellow }],
    ['Jinja', { icon: SiJinja, className: styles.defaultIcon }],
    ['JSON with Comments', { icon: VscJson, className: styles.defaultIcon }],
    ['JSON', { icon: VscJson, className: styles.defaultIcon }],
    ['JSON5', { icon: VscJson, className: styles.defaultIcon }],
    ['JSONLD', { icon: VscJson, className: styles.defaultIcon }],
    ['Julia', { icon: SiJulia, className: styles.defaultIcon }],
    ['Kotlin', { icon: SiKotlin, className: styles.green }],
    ['LLVM', { icon: SiLlvm, className: styles.gray }],
    ['Lua', { icon: SiLua, className: styles.blue }],
    ['Markdown', { icon: SiMarkdown, className: styles.blue }],
    ['Mathematica', { icon: SiWolframmathematica, className: styles.red }],

    // https://github.com/NCAR/ncl, not tweag/nickel
    ['NCL', { icon: ImEarth, className: styles.defaultIcon }],
    ['Nginx', { icon: SiNginx, className: styles.defaultIcon }],
    ['Nim', { icon: SiNim, className: styles.yellow }],
    ['Nix', { icon: SiNixos, className: styles.gray }],
    ['NPM Config', { icon: SiNpm, className: styles.red }],

    // Missing an icon for Objective-C
    ['OCaml', { icon: SiOcaml, className: styles.yellow }],
    ['PHP', { icon: SiPhp, className: styles.cyan }],
    ['Perl', { icon: SiPerl, className: styles.defaultIcon }],
    ['PLpgSQL', { icon: GoDatabase, className: styles.blue }],
    ['PowerShell', { icon: GoTerminal, className: styles.defaultIcon }],

    // Missing icon for Protobuf
    ['PureScript', { icon: SiPurescript, className: styles.defaultIcon }],
    ['Python', { icon: SiPython, className: styles.blue }],
    ['R', { icon: SiR, className: styles.red }],
    ['Ruby', { icon: SiRuby, className: styles.red }],
    ['Rust', { icon: SiRust, className: styles.defaultIcon }],
    ['Scala', { icon: SiScala, className: styles.red }],
    ['Sass', { icon: FaSass, className: styles.pink }],
    ['SCSS', { icon: FaSass, className: styles.pink }],
    ['SQL', { icon: GoDatabase, className: styles.blue }],

    // Missing icon for Starlark
    ['Svelte', { icon: SiSvelte, className: styles.red }],
    ['SVG', { icon: SiSvg, className: styles.yellow }],
    ['Swift', { icon: SiSwift, className: styles.blue }],
    ['Terraform', { icon: SiTerraform, className: styles.blue }],
    ['TSX', { icon: SiTypescript, className: styles.blue }],
    ['TypeScript', { icon: SiTypescript, className: styles.blue }],
    ['Text', { icon: CiTextAlignLeft, className: styles.defaultIcon }],

    // Missing icon for Thrift
    ['TOML', { icon: SiToml, className: styles.defaultIcon }],
    ['UnrealScript', { icon: SiUnrealengine, className: styles.defaultIcon }],
    ['VBA', { icon: SiVisualbasic, className: styles.blue }],
    ['VBScript', { icon: SiVisualbasic, className: styles.blue }],
    ['Vim Script', { icon: SiVim, className: styles.defaultIcon }],
    ['Vue', { icon: FaVuejs, className: styles.green }],
    ['WebAssembly', { icon: SiWebassembly, className: styles.blue }],
    ['XML', { icon: CiSettings, className: styles.defaultIcon }],
    ['YAML', { icon: CiSettings, className: styles.defaultIcon }],
    ['Zig', { icon: SiZig, className: styles.yellow }],
])

/**
 * DO NOT add any extensions here for which there are multiple different
 * file formats in practice which use the same extensions.
 *
 * For programming languages, update {@link FILE_ICONS_BY_LANGUAGE}.
 */
const BINARY_FILE_ICONS_BY_EXTENSION: Map<string, IconInfo> = new Map([
    ['gif', { icon: MdGif, className: styles.defaultIcon }],
    ['giff', { icon: MdGif, className: styles.defaultIcon }],
    ['jpg', { icon: SiJpeg, className: styles.yellow }],
    ['jpeg', { icon: SiJpeg, className: styles.yellow }],
    ['png', { icon: PiFilePngLight, className: styles.defaultIcon }],
])

/**
 *
 * See FIXME(id: language-detection) for context on why this takes a
 * languages argument instead of directly using the file extension
 * for determining the language.
 *
 * @param path The path of the file (or just its name).
 * @param language Alias to the file language name.
 * @returns undefined if the language is not a known language.
 */
export function getFileIconInfo(path: string, language: string): IconInfo | undefined {
    const extension = path.split('.').at(-1) ?? ''
    return BINARY_FILE_ICONS_BY_EXTENSION.get(extension) ?? FILE_ICONS_BY_LANGUAGE.get(language)
}

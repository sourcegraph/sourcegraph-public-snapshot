import type { ComponentType, SvelteComponent } from 'svelte'
import type { SvelteHTMLElements } from 'svelte/elements'

interface LanguageIcon {
    icon: ComponentType<SvelteComponent<SvelteHTMLElements['svg']>>
    color: string
}

const BLUE = 'var(--blue)'
const PINK = 'var(--pink)'
const YELLOW = 'var(--yellow)'
const RED = 'var(--red)'
const GREEN = 'var(--green)'
const CYAN = 'var(--blue)'
const GRAY = 'var(--gray-05)'
export const DEFAULT_ICON_COLOR = GRAY

/**
 * The keys of this map must be present in the list of `languageFilter.ALL_LANGUAGES`.
 *
 * This map is deliberately not public, use {@link getFileIconInfo} instead.
 *
 * See FIXME(id: language-detection) for context on why this map uses the
 * language as the key instead of something simpler like a file extension.
 */
export const FILE_ICONS_BY_LANGUAGE: Map<string, LanguageIcon> = new Map([
    ['Bash', { icon: ILucideFileTerminal, color: DEFAULT_ICON_COLOR }],
    ['BASIC', { icon: ISimpleIconsVisualbasic, color: DEFAULT_ICON_COLOR }],
    ['C', { icon: ISimpleIconsC, color: BLUE }],
    ['C++', { icon: ISimpleIconsCplusplus, color: BLUE }],
    ['C#', { icon: ISimpleIconsCsharp, color: BLUE }],
    ['Clojure', { icon: ISimpleIconsClojure, color: BLUE }],
    ['CMake', { icon: ISimpleIconsCmake, color: DEFAULT_ICON_COLOR }],
    ['CoffeeScript', { icon: ISimpleIconsCoffeescript, color: DEFAULT_ICON_COLOR }],

    // TODO: Decide icon for CSV?
    ['Crystal', { icon: ISimpleIconsCrystal, color: BLUE }],
    ['CSS', { icon: ISimpleIconsCss3, color: BLUE }],
    ['D', { icon: ISimpleIconsD, color: RED }],
    ['Dart', { icon: ISimpleIconsDart, color: BLUE }],
    ['Dockerfile', { icon: ISimpleIconsDocker, color: BLUE }],
    ['EditorConfig', { icon: ISimpleIconsEditorconfig, color: DEFAULT_ICON_COLOR }],
    ['Elixir', { icon: ISimpleIconsElixir, color: BLUE }],
    ['Elm', { icon: ISimpleIconsElm, color: BLUE }],
    ['Emacs Lisp', { icon: ISimpleIconsGnuemacs, color: DEFAULT_ICON_COLOR }],
    ['Erlang', { icon: ISimpleIconsErlang, color: BLUE }],
    ['Fortran', { icon: ISimpleIconsFortran, color: DEFAULT_ICON_COLOR }],
    ['Fortran Free Form', { icon: ISimpleIconsFortran, color: DEFAULT_ICON_COLOR }],
    ['F#', { icon: ISimpleIconsFsharp, color: BLUE }],
    ['Git Attributes', { icon: ISimpleIconsGit, color: RED }],
    ['Go', { icon: ISimpleIconsGo, color: BLUE }],
    ['Go Module', { icon: ISimpleIconsGo, color: PINK }],
    ['Go Checksums', { icon: ISimpleIconsGo, color: PINK }],
    ['Groovy', { icon: ISimpleIconsApachegroovy, color: BLUE }],
    ['GraphQL', { icon: ISimpleIconsGraphql, color: PINK }],
    ['Hack', { icon: IFileIconsHack, color: YELLOW }],
    ['Haskell', { icon: ISimpleIconsHaskell, color: BLUE }],
    ['HTML', { icon: ISimpleIconsHtml5, color: BLUE }],
    ['HTML+ECR', { icon: ISimpleIconsCrystal, color: BLUE }],
    ['HTML+EEX', { icon: ISimpleIconsElixir, color: BLUE }],
    ['HTML+ERB', { icon: ISimpleIconsRuby, color: BLUE }],
    ['HTML+PHP', { icon: ISimpleIconsPhp, color: BLUE }],
    ['HTML+Razor', { icon: ISimpleIconsCsharp, color: BLUE }],
    ['Ignore List', { icon: ILucideSettings, color: DEFAULT_ICON_COLOR }],
    ['Java', { icon: IDeviconPlainJava, color: DEFAULT_ICON_COLOR }],
    ['JavaScript', { icon: ISimpleIconsJavascript, color: YELLOW }],
    ['Jinja', { icon: ISimpleIconsJinja, color: DEFAULT_ICON_COLOR }],
    ['JSON with Comments', { icon: ILucideFileJson, color: DEFAULT_ICON_COLOR }],
    ['JSON', { icon: ILucideFileJson, color: DEFAULT_ICON_COLOR }],
    ['JSON5', { icon: ILucideFileJson, color: DEFAULT_ICON_COLOR }],
    ['JSONLD', { icon: ILucideFileJson, color: DEFAULT_ICON_COLOR }],
    ['Julia', { icon: ISimpleIconsJulia, color: DEFAULT_ICON_COLOR }],
    ['Kotlin', { icon: ISimpleIconsKotlin, color: GREEN }],
    ['LLVM', { icon: ISimpleIconsLlvm, color: GRAY }],
    ['Lua', { icon: ISimpleIconsLua, color: BLUE }],
    ['Markdown', { icon: ISimpleIconsMarkdown, color: BLUE }],
    ['Mathematica', { icon: ISimpleIconsWolframmathematica, color: RED }],

    // https://github.com/NCAR/ncl, not tweag/nickel
    ['NCL', { icon: ILucideEarth, color: DEFAULT_ICON_COLOR }],
    ['Nginx', { icon: ISimpleIconsNginx, color: DEFAULT_ICON_COLOR }],
    ['Nim', { icon: ISimpleIconsNim, color: YELLOW }],
    ['Nix', { icon: ISimpleIconsNixos, color: GRAY }],
    ['NPM Config', { icon: ISimpleIconsNpm, color: RED }],

    // Missing an icon for Objective-C
    ['OCaml', { icon: ISimpleIconsOcaml, color: YELLOW }],
    ['PHP', { icon: ISimpleIconsPhp, color: CYAN }],
    ['Perl', { icon: ISimpleIconsPerl, color: DEFAULT_ICON_COLOR }],
    ['PLpgSQL', { icon: ILucideDatabase, color: BLUE }],
    ['PowerShell', { icon: ILucideFileTerminal, color: DEFAULT_ICON_COLOR }],

    // Missing icon for Protobuf
    ['PureScript', { icon: ISimpleIconsPurescript, color: DEFAULT_ICON_COLOR }],
    ['Python', { icon: ISimpleIconsPython, color: BLUE }],
    ['R', { icon: ISimpleIconsR, color: RED }],
    ['Ruby', { icon: ISimpleIconsRuby, color: RED }],
    ['Rust', { icon: ISimpleIconsRust, color: DEFAULT_ICON_COLOR }],
    ['Scala', { icon: ISimpleIconsScala, color: RED }],
    ['Sass', { icon: ISimpleIconsSass, color: PINK }],
    ['SCSS', { icon: ISimpleIconsSass, color: PINK }],
    ['SQL', { icon: ILucideDatabase, color: BLUE }],

    // Missing icon for Starlark
    ['Svelte', { icon: ISimpleIconsSvelte, color: RED }],
    ['SVG', { icon: ISimpleIconsSvg, color: YELLOW }],
    ['Swift', { icon: ISimpleIconsSwift, color: BLUE }],
    ['Terraform', { icon: ISimpleIconsTerraform, color: BLUE }],
    ['TSX', { icon: ISimpleIconsTypescript, color: BLUE }],
    ['TypeScript', { icon: ISimpleIconsTypescript, color: BLUE }],
    ['Text', { icon: ILucideFileText, color: DEFAULT_ICON_COLOR }],

    // Missing icon for Thrift
    ['TOML', { icon: ISimpleIconsToml, color: DEFAULT_ICON_COLOR }],
    ['UnrealScript', { icon: ISimpleIconsUnrealengine, color: DEFAULT_ICON_COLOR }],
    ['VBA', { icon: ISimpleIconsVisualbasic, color: BLUE }],
    ['VBScript', { icon: ISimpleIconsVisualbasic, color: BLUE }],
    ['Vim Script', { icon: ISimpleIconsVim, color: DEFAULT_ICON_COLOR }],
    ['Vue', { icon: ISimpleIconsVuedotjs, color: GREEN }],
    ['WebAssembly', { icon: ISimpleIconsWebassembly, color: BLUE }],
    ['XML', { icon: ILucideSettings, color: DEFAULT_ICON_COLOR }],
    ['YAML', { icon: ILucideSettings, color: DEFAULT_ICON_COLOR }],
    ['Zig', { icon: ISimpleIconsZig, color: YELLOW }],
])

/**
 * DO NOT add any extensions here for which there are multiple different
 * file formats in practice which use the same extensions.
 *
 * For programming languages, update {@link FILE_ICONS_BY_LANGUAGE}.
 */
const BINARY_FILE_ICONS_BY_EXTENSION: Map<string, LanguageIcon> = new Map([
    ['gif', { icon: IPhGifFill, color: DEFAULT_ICON_COLOR }],
    ['giff', { icon: IPhGifFill, color: DEFAULT_ICON_COLOR }],
    ['jpg', { icon: IPhFileJpgLight, color: YELLOW }],
    ['jpeg', { icon: IPhFileJpgLight, color: YELLOW }],
    ['png', { icon: IPhFilePngLight, color: DEFAULT_ICON_COLOR }],
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
export function getFileIconInfo(path: string, language: string): LanguageIcon | undefined {
    const extension = path.split('.').at(-1) ?? ''
    return BINARY_FILE_ICONS_BY_EXTENSION.get(extension) ?? FILE_ICONS_BY_LANGUAGE.get(language)
}

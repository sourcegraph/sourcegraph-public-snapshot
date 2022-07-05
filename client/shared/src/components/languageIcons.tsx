import React from 'react'

import {
    mdiCodeJson,
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
    mdiLanguageLua,
    mdiLanguageMarkdown,
    mdiLanguagePhp,
    mdiLanguagePython,
    mdiLanguageR,
    mdiLanguageRuby,
    mdiLanguageSwift,
    mdiLanguageTypescript,
    mdiPowershell,
    mdiSass,
    mdiWeb,
} from '@mdi/js'

import { Icon } from '@sourcegraph/wildcard'

import { IconProps } from './icons'

interface Props extends IconProps {
    language: string
}

/**
 * Record of known valid language values for the `lang:` filter to their icon in suggestions.
 */
export const languageIcons: Record<string, string | undefined> = {
    __proto__: null as any,

    bash: undefined,
    c: mdiLanguageC,
    cobol: undefined,
    clojure: undefined,
    cpp: mdiLanguageCpp,
    csharp: mdiLanguageCsharp,
    css: mdiLanguageCss3,
    dart: undefined,
    go: mdiLanguageGo,
    graphql: mdiGraphql,
    erlang: undefined,
    elixir: undefined,
    haskell: mdiLanguageHaskell,
    html: mdiLanguageHtml5,
    java: mdiLanguageJava,
    javascript: mdiLanguageJavascript,
    json: mdiCodeJson,
    julia: undefined,
    kotlin: undefined,
    lua: mdiLanguageLua,
    markdown: mdiLanguageMarkdown,
    ocaml: undefined,
    objectivec: undefined,
    php: mdiLanguagePhp,
    protobuf: undefined,
    powershell: mdiPowershell,
    python: mdiLanguagePython,
    r: mdiLanguageR,
    rust: undefined,
    ruby: mdiLanguageRuby,
    sass: mdiSass,
    scala: undefined,
    sql: undefined,
    swift: mdiLanguageSwift,
    typescript: mdiLanguageTypescript,
    webassembly: undefined,
}

export const LanguageIcon: React.FunctionComponent<React.PropsWithChildren<Props>> = ({ language, size }) => {
    const LanguageIconComponent = languageIcons[language] || mdiWeb
    return <Icon svgPath={LanguageIconComponent} inline={false} size={size} aria-hidden={true} />
}

import React from 'react'
import { IconProps } from './icons'
import LanguageCIcon from 'mdi-react/LanguageCIcon'
import LanguageCppIcon from 'mdi-react/LanguageCppIcon'
import LanguageCsharpIcon from 'mdi-react/LanguageCsharpIcon'
import LanguageCss3Icon from 'mdi-react/LanguageCss3Icon'
import LanguageGoIcon from 'mdi-react/LanguageGoIcon'
import LanguageHaskellIcon from 'mdi-react/LanguageHaskellIcon'
import LanguageHtml5Icon from 'mdi-react/LanguageHtml5Icon'
import LanguageJavaIcon from 'mdi-react/LanguageJavaIcon'
import LanguageJavascriptIcon from 'mdi-react/LanguageJavascriptIcon'
import LanguageLuaIcon from 'mdi-react/LanguageLuaIcon'
import LanguagePhpIcon from 'mdi-react/LanguagePhpIcon'
import LanguagePythonIcon from 'mdi-react/LanguagePythonIcon'
import LanguageRIcon from 'mdi-react/LanguageRIcon'
import LanguageSwiftIcon from 'mdi-react/LanguageSwiftIcon'
import SassIcon from 'mdi-react/SassIcon'
import LanguageTypescriptIcon from 'mdi-react/LanguageTypescriptIcon'
import GraphqlIcon from 'mdi-react/GraphqlIcon'
import PowershellIcon from 'mdi-react/PowershellIcon'
import TranslateIcon from 'mdi-react/TranslateIcon'
import MarkdownIcon from 'mdi-react/MarkdownIcon'
import RubyIcon from 'mdi-react/RubyIcon'
import JsonIcon from 'mdi-react/JsonIcon'
import { MdiReactIconComponentType } from 'mdi-react'

interface Props extends IconProps {
    language: string
}

/**
 * Record of known valid language values for the `lang:` filter to their icon in suggestions.
 */
export const languageIcons: Record<string, MdiReactIconComponentType | undefined> = {
    __proto__: null as any,

    c: LanguageCIcon,
    cpp: LanguageCppIcon,
    csharp: LanguageCsharpIcon,
    css: LanguageCss3Icon,
    go: LanguageGoIcon,
    graphql: GraphqlIcon,
    haskell: LanguageHaskellIcon,
    html: LanguageHtml5Icon,
    java: LanguageJavaIcon,
    javascript: LanguageJavascriptIcon,
    json: JsonIcon,
    lua: LanguageLuaIcon,
    markdown: MarkdownIcon,
    php: LanguagePhpIcon,
    powershell: PowershellIcon,
    python: LanguagePythonIcon,
    r: LanguageRIcon,
    ruby: RubyIcon,
    sass: SassIcon,
    swift: LanguageSwiftIcon,
    typescript: LanguageTypescriptIcon,
}

export const LanguageIcon: React.FunctionComponent<Props> = ({ language, size }) => {
    const LanguageIconComponent = languageIcons[language] || TranslateIcon
    return <LanguageIconComponent size={size} />
}

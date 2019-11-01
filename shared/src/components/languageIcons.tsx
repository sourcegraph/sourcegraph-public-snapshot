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
import LanguageTypescriptIcon from 'mdi-react/LanguageTypescriptIcon'
import TranslateIcon from 'mdi-react/TranslateIcon'
import MarkdownIcon from 'mdi-react/MarkdownIcon'
import RubyIcon from 'mdi-react/RubyIcon'

interface Props extends IconProps {
    language: string
}

const languageIconComponents = {
    markdown: MarkdownIcon,
    c: LanguageCIcon,
    cpp: LanguageCppIcon,
    csharp: LanguageCsharpIcon,
    css: LanguageCss3Icon,
    go: LanguageGoIcon,
    haskell: LanguageHaskellIcon,
    html: LanguageHtml5Icon,
    java: LanguageJavaIcon,
    javascript: LanguageJavascriptIcon,
    lua: LanguageLuaIcon,
    php: LanguagePhpIcon,
    python: LanguagePythonIcon,
    r: LanguageRIcon,
    ruby: RubyIcon,
    swift: LanguageSwiftIcon,
    typescript: LanguageTypescriptIcon,
}

const isValidLanguage = (language: string): language is keyof typeof languageIconComponents =>
    Object.prototype.hasOwnProperty.call(languageIconComponents, language)

export const LanguageIcon: React.FunctionComponent<Props> = ({ language, size }) => {
    const LanguageIconComponent = isValidLanguage(language) ? languageIconComponents[language] : TranslateIcon
    return <LanguageIconComponent size={size} />
}

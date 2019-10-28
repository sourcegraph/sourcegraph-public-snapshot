import React from 'react'
import { IconProps } from './icons'
import JavascriptIcon from 'mdi-react/LanguageJavascriptIcon'
import GoIcon from 'mdi-react/LanguageGoIcon'
import MarkdownIcon from 'mdi-react/MarkdownIcon'
import TranslateIcon from 'mdi-react/TranslateIcon'
interface Props extends IconProps {
    language: string
}

export const LanguageIcon: React.FunctionComponent<Props> = ({ language, size }) => {
    switch (language) {
        case 'go':
            return <GoIcon size={size} />
        case 'javascript':
            return <JavascriptIcon size={size} />
        case 'markdown':
            return <MarkdownIcon size={size} />
        default:
            return <TranslateIcon size={size} />
    }
}

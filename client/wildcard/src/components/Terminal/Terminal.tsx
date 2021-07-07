import React from 'react'

import terminalStyles from './Terminal.module.scss'

// 73 '=' characters are the 100% of the progress bar
const CHARACTERS_LENGTH = 73
const CHARACTER = '='

export const Terminal: React.FunctionComponent = ({ children }) => (
    <div className={terminalStyles.mainWrapper}>
        <section className={terminalStyles.terminalWrapper}>
            <code>Cloning Repositories...</code>
            <ul className={terminalStyles.downloadProgressWrapper}>{children}</ul>
        </section>
    </div>
)

export const TerminalTitle: React.FunctionComponent = ({ children }) => (
    <header className={terminalStyles.headerTitle}>
        <code>{children}</code>
    </header>
)

export const TerminalLine: React.FunctionComponent = ({ children }) => (
    <li className={terminalStyles.terminalLine}>{children}</li>
)

export const TerminalDetails: React.FunctionComponent = ({ children }) => <code>{children}</code>

export const TerminalProgress: React.FunctionComponent<{ progress: number }> = ({ progress = 0 }) => {
    const numberOfChars = Math.ceil((progress / 100) * CHARACTERS_LENGTH)

    return (
        <code className={terminalStyles.downloadProgress}>
            {CHARACTER.repeat(numberOfChars)}
            {'>'}
        </code>
    )
}

import React from 'react'

import classNames from 'classnames'

import { Code } from '@sourcegraph/wildcard'

import terminalStyles from './Terminal.module.scss'

interface TerminalLineProps extends React.HTMLAttributes<HTMLLIElement> {
    className?: string
}

// 73 '=' characters are the 100% of the progress bar
const CHARACTERS_LENGTH = 73

export const Terminal: React.FunctionComponent<React.PropsWithChildren<unknown>> = ({ children }) => (
    <section className={terminalStyles.terminalWrapper}>
        <ul className={terminalStyles.downloadProgressWrapper}>{children}</ul>
    </section>
)

export const TerminalTitle: React.FunctionComponent<React.PropsWithChildren<unknown>> = ({ children }) => (
    <header className={terminalStyles.terminalTitle}>
        <Code>{children}</Code>
    </header>
)

export const TerminalLine: React.FunctionComponent<React.PropsWithChildren<TerminalLineProps>> = ({
    children,
    className,
}) => <li className={classNames(terminalStyles.terminalLine, className)}>{children}</li>

export const TerminalDetails: React.FunctionComponent<React.PropsWithChildren<unknown>> = ({ children }) => (
    <div>
        <Code>{children}</Code>
    </div>
)

export const TerminalProgress: React.FunctionComponent<
    React.PropsWithChildren<{ progress: number; character: string }>
> = ({ progress = 0, character = '#' }) => {
    const numberOfChars = Math.ceil((progress / 100) * CHARACTERS_LENGTH)

    return <Code className={terminalStyles.downloadProgress}>{character.repeat(numberOfChars)}</Code>
}

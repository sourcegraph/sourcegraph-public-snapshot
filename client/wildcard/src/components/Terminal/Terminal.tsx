import React, { useRef, useEffect, useState } from 'react'

import terminalStyles from './Terminal.module.scss'

/**
 * 42 '=' characters are the 100% of the progress bar
 * I need to define in a width of 300 how many characters needs to show
 * based on the repository total size.
 *
 * e.g repo 10MB -> 10% of progress = 30px of width and 4.2 characters rounded to 4.
 */

const CHARACTERS_LENGTH = 42
const CHARACTER = '='

export const Terminal: React.FunctionComponent = () => {
    const [progress, setProgress] = useState('>')
    const [status, setStatus] = useState('Downloading...')
    const progressContainerReference = useRef<HTMLSpanElement>(null)
    const progressReference = useRef<HTMLSpanElement>(null)

    const [response, setResponse] = useState({
        repoName: 'sourcegraph',
        id: '12345',
        progress: {
            size: 3.62,
            percentage: 0,
        },
        size: 36.2,
    })

    const [percentage, setPercentage] = useState(0)

    useEffect(() => {
        const interval = setInterval(() => {
            setPercentage(seconds => seconds + 1)
        }, 1000)

        // TODO: not working
        if (status === 'Completed') {
            return () => clearInterval(interval)
        }

        return () => clearInterval(interval)
    }, [status])

    useEffect(() => {
        setResponse({
            repoName: 'sourcegraph',
            id: '12345',
            progress: {
                size: 3.62,
                percentage,
            },
            size: 36.2,
        })
    }, [percentage])

    useEffect(() => {
        const progressBarTracker = () => {
            if (!progressReference.current || !progressContainerReference.current) {
                return null
            }

            const numberOfChars = Math.ceil((response.progress.percentage / CHARACTERS_LENGTH) * 100)

            // if (progressReference.current.clientWidth <= 300 /* progressContainerReference.current.clientWidth*/) {
            if (numberOfChars < 42) {
                setProgress(`${CHARACTER.repeat(numberOfChars)}>`)
            } else {
                setStatus('Completed')
            }
        }

        progressBarTracker()
    }, [response])

    return (
        <div className={terminalStyles.wrapper}>
            <header className={terminalStyles.headerTitle}>
                <h3 className="title">Activity log</h3>
                <div>{percentage}</div>
            </header>
            <section className={terminalStyles.repositoriesInfo}>
                <code>Cloning Repositories...</code>
                <ul className={terminalStyles.downloadProgressWrapper}>
                    <li className={terminalStyles.downloadProgressItem}>
                        <code className={terminalStyles.downloadProgressItemDetails}>
                            <span className={terminalStyles.downloadStatus}>{status}</span>
                            <span ref={progressContainerReference} className={terminalStyles.downloadProgress}>
                                <span ref={progressReference}>{progress}</span>
                            </span>
                            <span>
                                {response.progress.size}MB/{response.size}MB
                            </span>
                        </code>
                    </li>
                    <li className={terminalStyles.downloadProgressItem}>
                        <code className={terminalStyles.downloadProgressItemDetails}>
                            <span className={terminalStyles.downloadStatus}>{status}</span>
                            <span ref={progressContainerReference} className={terminalStyles.downloadProgress}>
                                <span ref={progressReference}>{progress}</span>
                            </span>
                            <span>
                                {response.progress.size}MB/{response.size}MB
                            </span>
                        </code>
                    </li>
                    <li className={terminalStyles.downloadProgressItem}>
                        <code className={terminalStyles.downloadProgressItemDetails}>
                            <span className={terminalStyles.downloadStatus}>{status}</span>
                            <span ref={progressContainerReference} className={terminalStyles.downloadProgress}>
                                <span ref={progressReference}>{progress}</span>
                            </span>
                            <span>
                                {response.progress.size}MB/{response.size}MB
                            </span>
                        </code>
                    </li>
                </ul>
            </section>
        </div>
    )
}

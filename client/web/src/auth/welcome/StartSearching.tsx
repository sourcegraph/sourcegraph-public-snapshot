import React, { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'

import {
    Terminal,
    TerminalTitle,
    TerminalLine,
    TerminalDetails,
    TerminalProgress,
} from '@sourcegraph/wildcard/src/components/Terminal'

import { LogoAscii } from '../LogoAscii'
import { RepoCloningStatus } from '../useRepoCloningStatus'

interface StartSearching {
    cloningStatusLines: RepoCloningStatus['repos']
    cloningStatusLoading: RepoCloningStatus['loading']
    isDoneCloning: RepoCloningStatus['isDoneCloning']
}

export const useShowAlert = (): { showAlert: boolean } => {
    const [showAlert, setShowAlert] = useState(false)

    useEffect(() => {
        const timer = setTimeout(() => setShowAlert(true), 10000)

        return () => {
            clearTimeout(timer)
            setShowAlert(false)
        }
    }, [])

    return { showAlert }
}

export const StartSearching: React.FunctionComponent<StartSearching> = ({
    cloningStatusLines,
    cloningStatusLoading,
    isDoneCloning,
}) => {
    const { showAlert } = useShowAlert()

    return (
        <>
            <h3>Start searching...</h3>
            <p className="text-muted">
                We’re cloning your repos to Sourcegraph. In just a few moments, you can make your first search!
            </p>
            {/* <p>{`cloningStatusLoading: ${cloningStatusLoading}`}</p>
                <p>{`isDoneCloning: ${isDoneCloning}`}</p>
                <p>{`cloningStatusLines count: ${
                    cloningStatusLines ? cloningStatusLines.length : 'undefined'
                }`}</p> */}
            <div className="border overflow-hidden rounded">
                <header>
                    <h3 className="m-0 pl-4 py-3">Activity log</h3>
                </header>
                <Terminal>
                    {cloningStatusLoading && (
                        <TerminalLine>
                            <TerminalTitle>Loading...</TerminalTitle>
                        </TerminalLine>
                    )}
                    {!cloningStatusLoading &&
                        !isDoneCloning &&
                        cloningStatusLines?.map(({ id, title, details, progress }) => (
                            <React.Fragment key={id}>
                                <TerminalLine>
                                    <TerminalTitle>{title}</TerminalTitle>
                                </TerminalLine>
                                <TerminalLine>
                                    <TerminalDetails>{details}</TerminalDetails>
                                </TerminalLine>
                                <TerminalLine>
                                    <TerminalProgress character="#" progress={progress} />
                                </TerminalLine>
                            </React.Fragment>
                        ))}
                    {isDoneCloning && (
                        <>
                            <TerminalLine>
                                <TerminalTitle>Done!</TerminalTitle>
                            </TerminalLine>
                            <TerminalLine>
                                <LogoAscii />
                            </TerminalLine>
                        </>
                    )}
                </Terminal>
            </div>
            {showAlert && (
                <div className="alert alert-warning mt-4">
                    Cloning your repositories is taking a long time. You can wait for cloning to finish, or{' '}
                    <Link to="/search">continue to Sourcegraph now</Link> while cloning continues in the background.
                    Note that you can only search repos that have finished cloning. Check status at any time in Settings
                    → Your repositories.
                </div>
            )}
        </>
    )
}

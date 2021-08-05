import React, { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'

import { AuthenticatedUser } from '../../auth'
import { LogoAscii } from '../LogoAscii'
import { RepoSelectionMode } from '../PostSignUpPage'
import { useSteps } from '../Steps'
import { Terminal, TerminalTitle, TerminalLine, TerminalDetails, TerminalProgress } from '../Terminal'
import { useExternalServices } from '../useExternalServices'
import { useRepoCloningStatus } from '../useRepoCloningStatus'
import { selectedReposVar, useSaveSelectedRepos } from '../useSelectedRepos'

interface StartSearching {
    user: AuthenticatedUser
    repoSelectionMode: RepoSelectionMode
}

const SIXTY_SECONDS = 60000

export const useShowAlert = (isDoneCloning: boolean): { showAlert: boolean } => {
    const [showAlert, setShowAlert] = useState(false)

    useEffect(() => {
        const timer = setTimeout(() => setShowAlert(true), SIXTY_SECONDS)

        if (isDoneCloning) {
            clearTimeout(timer)
            setShowAlert(false)
        }

        return () => {
            clearTimeout(timer)
            setShowAlert(false)
        }
    }, [isDoneCloning])

    return { showAlert }
}

export const StartSearching: React.FunctionComponent<StartSearching> = ({ user, repoSelectionMode }) => {
    const { externalServices } = useExternalServices(user.id)
    const saveSelectedRepos = useSaveSelectedRepos()

    const {
        repos: cloningStatusLines,
        statusSummary,
        loading: cloningStatusLoading,
        isDoneCloning,
    } = useRepoCloningStatus({
        userId: user.id,
        pollInterval: 2000,
        selectedReposVar,
    })

    useEffect(() => {
        const selectedRepos = selectedReposVar()

        if (externalServices && selectedRepos) {
            const codeHostRepoPromises = []

            for (const host of externalServices) {
                const repos: string[] = []
                for (const repo of selectedRepos) {
                    if (repo.externalRepository.id !== host.id) {
                        continue
                    }

                    const nameWithoutService = repo.name.slice(repo.name.indexOf('/') + 1)
                    repos.push(nameWithoutService)
                }

                codeHostRepoPromises.push(
                    saveSelectedRepos({
                        variables: {
                            id: host.id,
                            allRepos: repoSelectionMode === 'all',
                            repos: (repoSelectionMode === 'selected' && repos) || null,
                        },
                    })
                )
            }
        }
    }, [externalServices, saveSelectedRepos, repoSelectionMode])

    const { showAlert } = useShowAlert(isDoneCloning)
    const { currentIndex, setComplete } = useSteps()

    useEffect(() => {
        if (showAlert) {
            setComplete(currentIndex, true)
        } else {
            setComplete(currentIndex, isDoneCloning)
        }
    }, [currentIndex, setComplete, showAlert, isDoneCloning])

    return (
        <div className="mt-5">
            <h3>Start searching...</h3>
            <p className="text-muted mb-4">
                We’re cloning your repos to Sourcegraph. In just a few moments, you can make your first search!
            </p>
            <div className="border overflow-hidden rounded">
                <header>
                    <div className="py-3 px-4">
                        <h3 className="d-inline-block m-0">Activity log</h3>
                        <span className="float-right m-0">{statusSummary}</span>
                    </div>
                </header>
                <Terminal>
                    {!isDoneCloning && (
                        <TerminalLine>
                            <code className="mb-2 post-signup-page__loading">Cloning Repositories</code>
                        </TerminalLine>
                    )}
                    {cloningStatusLoading && (
                        <TerminalLine>
                            <TerminalTitle>
                                <code className="mb-2 post-signup-page__loading">Loading</code>
                            </TerminalTitle>
                        </TerminalLine>
                    )}
                    {!cloningStatusLoading &&
                        !isDoneCloning &&
                        cloningStatusLines?.map(({ id, title, details, progress }) => (
                            <div key={id} className="mb-2">
                                <TerminalLine>
                                    <TerminalTitle>{title}</TerminalTitle>
                                </TerminalLine>
                                <TerminalLine>
                                    <TerminalDetails>{details}</TerminalDetails>
                                </TerminalLine>
                                <TerminalLine>
                                    <TerminalProgress character="#" progress={progress} />
                                </TerminalLine>
                            </div>
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
                    Note that you can only search repos that have finished cloning. Check status at any time in{' '}
                    <Link to="user/settings/repositories">Settings → Repositories</Link>.
                </div>
            )}
        </div>
    )
}

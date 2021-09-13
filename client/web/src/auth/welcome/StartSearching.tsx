import classNames from 'classnames'
import React, { useEffect, useState } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { ErrorLike, isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { AuthenticatedUser } from '../../auth'
import { eventLogger } from '../../tracking/eventLogger'
import { UserExternalServicesOrRepositoriesUpdateProps } from '../../util'
import { LogoAscii } from '../LogoAscii'
import { getPostSignUpEvent, RepoSelectionMode } from '../PostSignUpPage'
import { useSteps } from '../Steps'
import { Terminal, TerminalTitle, TerminalLine, TerminalDetails, TerminalProgress } from '../Terminal'
import { useExternalServices } from '../useExternalServices'
import { useRepoCloningStatus } from '../useRepoCloningStatus'
import { selectedReposVar, useSaveSelectedRepos, MinSelectedRepo } from '../useSelectedRepos'

import styles from './StartSearching.module.scss'

interface StartSearching {
    user: AuthenticatedUser
    repoSelectionMode: RepoSelectionMode
    onUserExternalServicesOrRepositoriesUpdate: UserExternalServicesOrRepositoriesUpdateProps['onUserExternalServicesOrRepositoriesUpdate']
    setSelectedSearchContextSpec: (spec: string) => void
    onError: (error: ErrorLike) => void
}

const SIXTY_SECONDS = 60000

export const useShowAlert = (isDoneCloning: boolean, fetchError: ErrorLike | undefined): { showAlert: boolean } => {
    const [showAlert, setShowAlert] = useState(false)

    useEffect(() => {
        const timer = setTimeout(() => {
            eventLogger.log(getPostSignUpEvent('SlowCloneBanner_Shown'))
            setShowAlert(true)
        }, SIXTY_SECONDS)

        if (isDoneCloning || isErrorLike(fetchError)) {
            clearTimeout(timer)
            setShowAlert(false)
        }

        return () => {
            clearTimeout(timer)
            setShowAlert(false)
        }
    }, [isDoneCloning, fetchError])

    return { showAlert }
}

const trackBannerClick = (): void => {
    eventLogger.log(getPostSignUpEvent('SlowCloneBanner_Clicked'))
}
const getReposForCodeHost = (selectedRepos: MinSelectedRepo[] = [], codeHostId: string): string[] =>
    selectedRepos
        ? selectedRepos.reduce((accumulator, repo) => {
              if ((repo.externalRepository.id = codeHostId)) {
                  const nameWithoutService = repo.name.slice(repo.name.indexOf('/') + 1)
                  accumulator.push(nameWithoutService)
              }

              return accumulator
          }, [] as string[])
        : []

export const StartSearching: React.FunctionComponent<StartSearching> = ({
    user,
    repoSelectionMode,
    setSelectedSearchContextSpec,
    onError,
}) => {
    const { externalServices, errorServices, loadingServices } = useExternalServices(user.id)
    const saveSelectedRepos = useSaveSelectedRepos()

    const {
        repos: cloningStatusLines,
        statusSummary,
        loading: cloningStatusLoading,
        isDoneCloning,
        error: cloningStatusError,
        stopPolling: stopPollingCloningStatus,
    } = useRepoCloningStatus({
        userId: user.id,
        pollInterval: 2000,
        selectedReposVar,
        repoSelectionMode,
    })

    const isLoading = loadingServices || cloningStatusLoading
    const fetchError = cloningStatusError || errorServices

    useEffect(() => {
        if (fetchError) {
            stopPollingCloningStatus()
            onError(fetchError)
        }
    }, [fetchError, onError, stopPollingCloningStatus])

    useEffect(() => {
        if (externalServices) {
            const selectedRepos = selectedReposVar()

            for (const host of externalServices) {
                const areSyncingAllRepos = repoSelectionMode === 'all'
                // when we're in the "sync all" - don't list individual repos
                // set allRepos key to true
                const repos = areSyncingAllRepos ? null : getReposForCodeHost(selectedRepos, host.id)

                saveSelectedRepos({
                    variables: {
                        id: host.id,
                        allRepos: areSyncingAllRepos,
                        repos,
                    },
                })
                    .then(() => {
                        setSelectedSearchContextSpec(`@${user.username}`)
                    })
                    .catch(onError)
            }

            const loggerPayload = {
                userReposSelection: repoSelectionMode ? (repoSelectionMode === 'selected' ? 'specific' : 'all') : null,
            }

            eventLogger.log(getPostSignUpEvent('Repos_Saved'), loggerPayload, loggerPayload)
        }
    }, [externalServices, saveSelectedRepos, repoSelectionMode, onError, setSelectedSearchContextSpec, user.username])

    const { showAlert } = useShowAlert(isDoneCloning, fetchError)
    const { currentIndex, setComplete } = useSteps()

    useEffect(() => {
        if (showAlert) {
            setComplete(currentIndex, true)
        } else {
            setComplete(currentIndex, isDoneCloning || isErrorLike(fetchError))
        }
    }, [currentIndex, setComplete, showAlert, isDoneCloning, fetchError])

    return (
        <div className="mt-5">
            <h3>Fetching repositories...</h3>
            <p className="text-muted mb-4">
                We’re cloning your repos to Sourcegraph. In just a few moments, you can make your first search!
            </p>
            <div className="border overflow-hidden rounded">
                <header>
                    <div className="py-3 px-4 d-flex justify-content-between align-items-center">
                        <h4 className="m-0">Activity log</h4>
                        <small className="m-0 text-muted">{statusSummary}</small>
                    </div>
                </header>
                <Terminal>
                    {!isDoneCloning && (
                        <TerminalLine>
                            <code className={classNames('mb-2', styles.loading)}>Cloning Repositories</code>
                        </TerminalLine>
                    )}
                    {isLoading && (
                        <TerminalLine>
                            <TerminalTitle>
                                <code className={classNames('mb-2', styles.loading)}>Loading</code>
                            </TerminalTitle>
                        </TerminalLine>
                    )}
                    {fetchError && (
                        <TerminalLine>
                            <TerminalTitle>
                                <code className="mb-2">Unexpected error</code>
                            </TerminalTitle>
                        </TerminalLine>
                    )}
                    {!isLoading &&
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
                    <Link to="/search" onClick={trackBannerClick}>
                        continue to Sourcegraph now
                    </Link>{' '}
                    while cloning continues in the background. Note that you can only search repos that have finished
                    cloning. Check status at any time in{' '}
                    <Link to="user/settings/repositories" onClick={trackBannerClick}>
                        Settings → Repositories
                    </Link>
                    .
                </div>
            )}
        </div>
    )
}

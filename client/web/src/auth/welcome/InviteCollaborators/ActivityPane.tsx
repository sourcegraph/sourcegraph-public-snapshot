import React, { useEffect, useState } from 'react'

import { ApolloError } from '@apollo/client'
import classNames from 'classnames'

import { ErrorLike, isErrorLike } from '@sourcegraph/common'
import { Link, Alert, Typography } from '@sourcegraph/wildcard'

import { PageRoutes } from '../../../routes.constants'
import { eventLogger } from '../../../tracking/eventLogger'
import { LogoAscii } from '../../LogoAscii'
import { getPostSignUpEvent } from '../../PostSignUpPage'
import { useSteps } from '../../Steps'
import { Terminal, TerminalTitle, TerminalLine, TerminalDetails, TerminalProgress } from '../../Terminal'
import { RepoLine } from '../../useRepoCloningStatus'

import styles from './InviteCollaborators.module.scss'

const SIXTY_SECONDS = 60000

interface Props {
    className?: string
    statusSummary: string
    isDoneCloning: boolean
    isLoading: boolean
    cloningStatusLines: RepoLine[] | undefined
    fetchError: ApolloError | undefined
}

function useShowAlert(isDoneCloning: boolean, fetchError: ErrorLike | undefined): { showAlert: boolean } {
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

export const ActivityPane: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    className,
    statusSummary,
    isDoneCloning,
    isLoading,
    cloningStatusLines,
    fetchError,
}: Props) => {
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
        <div className={classNames(className, 'mx-2')}>
            <div className={styles.titleDescription}>
                <Typography.H3>Fetching repositories...</Typography.H3>
                <p className="text-muted mb-4">
                    We’re cloning your repos to Sourcegraph. In just a few moments, you can make your first search!
                </p>
            </div>
            <div className="border overflow-hidden rounded">
                <header>
                    <div className="py-4 px-3 d-flex justify-content-between align-items-center">
                        <Typography.H4 className="m-0">Activity log</Typography.H4>
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
                <Alert className="mt-4" variant="warning">
                    Cloning your repositories is taking a long time. You can wait for cloning to finish, or{' '}
                    <Link to={PageRoutes.Search} onClick={trackBannerClick}>
                        continue to Sourcegraph now
                    </Link>{' '}
                    while cloning continues in the background. Note that you can only search repos that have finished
                    cloning. Check status at any time in{' '}
                    <Link to="user/settings/repositories" onClick={trackBannerClick}>
                        Settings → Repositories
                    </Link>
                    .
                </Alert>
            )}
        </div>
    )
}

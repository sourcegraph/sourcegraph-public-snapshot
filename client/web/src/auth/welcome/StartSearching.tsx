import classNames from 'classnames'
import React, { useCallback, useEffect, useMemo, useState, useRef } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { ErrorLike, isErrorLike } from '@sourcegraph/common'
import * as GQL from '@sourcegraph/shared/src/schema'
import { CopyableText } from '@sourcegraph/web/src/components/CopyableText'
import { PageRoutes } from '@sourcegraph/web/src/routes.constants'
import { Link, Alert, LoadingSpinner, Button } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { eventLogger } from '../../tracking/eventLogger'
import { UserAvatar } from '../../user/UserAvatar'
import { UserExternalServicesOrRepositoriesUpdateProps } from '../../util'
import { LogoAscii } from '../LogoAscii'
import { getPostSignUpEvent, RepoSelectionMode, FinishWelcomeFlow } from '../PostSignUpPage'
import { useSteps } from '../Steps'
import { Terminal, TerminalTitle, TerminalLine, TerminalDetails, TerminalProgress } from '../Terminal'
import { useExternalServicesWithCollaborators } from '../useExternalServicesWithCollaborators'
import { useRepoCloningStatus } from '../useRepoCloningStatus'
import { selectedReposVar, useSaveSelectedRepos, MinSelectedRepo } from '../useSelectedRepos'

import styles from './StartSearching.module.scss'

interface StartSearching {
    user: AuthenticatedUser
    repoSelectionMode: RepoSelectionMode
    onUserExternalServicesOrRepositoriesUpdate: UserExternalServicesOrRepositoriesUpdateProps['onUserExternalServicesOrRepositoriesUpdate']
    setSelectedSearchContextSpec: (spec: string) => void
    onError: (error: ErrorLike) => void
    className?: string
    onFinish: FinishWelcomeFlow
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
    className,
    onFinish,
}) => {
    const {
        externalServices,
        errorServices,
        loadingServices,
        refetchExternalServices,
    } = useExternalServicesWithCollaborators(user.id)
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

    // We need to wait for the selected repos to be synced with the backend and
    // the external services to be refetched before we can display collaborators
    const [didRefetchExternalServices, setDidRefetchExternalServices] = useState(false)

    const isLoading = loadingServices || cloningStatusLoading
    const isLoadingCollaborators = loadingServices || !didRefetchExternalServices
    const fetchError = cloningStatusError || errorServices

    useEffect(() => {
        if (fetchError) {
            stopPollingCloningStatus()
            onError(fetchError)
        }
    }, [fetchError, onError, stopPollingCloningStatus])

    // Make sure the below repo saving is only called ones
    const didSaveSelectedReposReference = useRef(false)
    useEffect(() => {
        if (!externalServices || didSaveSelectedReposReference.current) {
            return
        }
        didSaveSelectedReposReference.current = true

        const selectedRepos = selectedReposVar()

        const savePromises: Promise<void>[] = []
        for (const host of externalServices) {
            const areSyncingAllRepos = repoSelectionMode === 'all'
            // when we're in the "sync all" - don't list individual repos
            // set allRepos key to true
            const repos = areSyncingAllRepos ? null : getReposForCodeHost(selectedRepos, host.id)

            savePromises.push(
                saveSelectedRepos({
                    variables: {
                        id: host.id,
                        allRepos: areSyncingAllRepos,
                        repos,
                    },
                })
                    .then(() => setSelectedSearchContextSpec(`@${user.username}`))
                    .catch(onError)
            )
        }

        const loggerPayload = {
            userReposSelection: repoSelectionMode ? (repoSelectionMode === 'selected' ? 'specific' : 'all') : null,
        }

        eventLogger.log(getPostSignUpEvent('Repos_Saved'), loggerPayload, loggerPayload)

        Promise.all(savePromises)
            .then(() => {
                setDidRefetchExternalServices(true)
                return refetchExternalServices({
                    namespace: user.id,
                    first: null,
                    after: null,
                })
            })
            .catch(onError)
    }, [
        externalServices,
        saveSelectedRepos,
        repoSelectionMode,
        onError,
        setSelectedSearchContextSpec,
        user.username,
        user.id,
        refetchExternalServices,
    ])

    const invitableCollaborators = useMemo<GQL.IPerson[]>(() => {
        const invitable = externalServices ? externalServices.flatMap(host => host.invitableCollaborators) : []
        const loggerPayload = {
            discovered: invitable.length,
        }
        eventLogger.log('UserInvitationsDiscoveredCollaborators', loggerPayload, loggerPayload)
        return invitable
    }, [externalServices])
    const preventSubmit = useCallback((event: React.FormEvent<HTMLFormElement>): void => event.preventDefault(), [])
    const [query, setQuery] = useState('')

    const filteredCollaborators: GQL.IPerson[] = useMemo(() => {
        if (query.trim() === '') {
            return invitableCollaborators
        }
        return invitableCollaborators.filter(person => person.name.includes(query) || person.email.includes(query))
    }, [query, invitableCollaborators])

    const [inviteError, setInviteError] = useState<ErrorLike | null>(null)
    const [loadingInvites, setLoadingInvites] = useState<Set<string>>(new Set<string>())

    const invitePerson = useCallback(
        async (person: GQL.IPerson): Promise<void> => {
            if (loadingInvites.has(person.email)) {
                return
            }
            setLoadingInvites(new Set(loadingInvites.add(person.email)))

            try {
                // TODO: actually send GraphQL request to invite via email.

                // dataOrThrowErrors(
                //     await requestGraphQL<ResendVerificationEmailResult, ResendVerificationEmailVariables>(
                //         gql`
                //             mutation ResendVerificationEmail($user: ID!, $email: String!) {
                //                 resendVerificationEmail(user: $user, email: $email) {
                //                     alwaysNil
                //                 }
                //             }
                //         `,
                //         { user, email }
                //     ).toPromise()
                // )

                const removed = new Set(loadingInvites)
                removed.delete(person.email)
                setLoadingInvites(removed)
                eventLogger.log('UserInvitationsSentEmailInvite')
            } catch (error) {
                setInviteError(error)
            }
        },
        [loadingInvites]
    )
    const invitePersonClicked = useCallback(
        (person: GQL.IPerson) => async (): Promise<void> => {
            await invitePerson(person)
        },
        [invitePerson]
    )
    const inviteAllClicked = useCallback(async (): Promise<void> => {
        for (const person of filteredCollaborators) {
            await invitePerson(person)
        }
    }, [invitePerson, filteredCollaborators])

    const { showAlert } = useShowAlert(isDoneCloning, fetchError)
    const { currentIndex, setComplete } = useSteps()

    useEffect(() => {
        if (showAlert) {
            setComplete(currentIndex, true)
        } else {
            setComplete(currentIndex, isDoneCloning || isErrorLike(fetchError))
        }
    }, [currentIndex, setComplete, showAlert, isDoneCloning, fetchError])

    const inviteURL = `${window.context.externalURL}/sign-up?invitedBy=${user.username}`
    return (
        <div className="m5-5 d-flex">
            <div className={classNames(className, 'mx-2')}>
                <div className={styles.titleDescription}>
                    <h3>Introduce friends and colleagues to Sourcegraph</h3>
                    <p className="text-muted mb-4">
                        We’ve selected a few collaborators from your repositories in case you wanted to level them up
                        with Sourcegraph’s powerful code search.
                    </p>
                </div>
                {isErrorLike(inviteError) && <ErrorAlert error={inviteError} />}
                <div className="border overflow-hidden rounded">
                    <header>
                        <div className="py-3 px-3 d-flex justify-content-between align-items-center">
                            <h4 className="flex-1 m-0">Collaborators</h4>
                            <Form
                                onSubmit={preventSubmit}
                                className="flex-1 d-inline-flex justify-content-between flex-row"
                            >
                                <input
                                    className="form-control"
                                    type="search"
                                    placeholder="Filter by email or username"
                                    name="query"
                                    autoComplete="off"
                                    autoCorrect="off"
                                    autoCapitalize="off"
                                    spellCheck={false}
                                    onChange={event => {
                                        setQuery(event.target.value)
                                    }}
                                />
                            </Form>
                        </div>
                    </header>
                    <div className={classNames('mb-3', styles.invitableCollaborators)}>
                        {!isLoadingCollaborators &&
                            filteredCollaborators.map((person, index) => (
                                <div
                                    className={classNames(
                                        'd-flex',
                                        'ml-3',
                                        'align-items-center',
                                        index !== 0 && 'mt-3'
                                    )}
                                    key={person.email}
                                >
                                    <UserAvatar
                                        className={classNames('icon-inline', 'mr-3', styles.avatar)}
                                        user={person}
                                    />
                                    <div>
                                        <strong>{person.displayName}</strong>
                                        <div className="text-muted">{person.email}</div>
                                    </div>
                                    {loadingInvites.has(person.email) ? (
                                        <LoadingSpinner inline={true} className={classNames('ml-auto', 'mr-3')} />
                                    ) : (
                                        <Button
                                            variant="secondary"
                                            outline={true}
                                            size="sm"
                                            className={classNames('ml-auto', 'mr-3', styles.inviteButton)}
                                            onClick={invitePersonClicked(person)}
                                        >
                                            Invite
                                        </Button>
                                    )}
                                </div>
                            ))}
                        {isLoadingCollaborators && (
                            <div className="text-muted d-flex justify-content-center mt-3">
                                <LoadingSpinner inline={false} />
                            </div>
                        )}
                        {!isLoadingCollaborators && filteredCollaborators.length === 0 && (
                            <div className="text-muted d-flex justify-content-center mt-3">
                                No collaborators found. Try sending them a direct link below
                            </div>
                        )}
                    </div>
                    <Button
                        variant="success"
                        className="d-block ml-auto mb-3 mr-3"
                        onClick={inviteAllClicked}
                        disabled={isLoadingCollaborators || filteredCollaborators.length === 0}
                    >
                        Invite all
                    </Button>
                </div>
                <div>
                    <header>
                        <div className="py-3 d-flex justify-content-between align-items-center">
                            <h4 className="m-0">Or invite by sending a link</h4>
                        </div>
                    </header>
                    <CopyableText
                        className="mb-3 flex-1"
                        text={inviteURL}
                        flex={true}
                        size={inviteURL.length}
                        onCopy={() => eventLogger.log('UserInvitationsCopiedInviteLink')}
                    />
                </div>
            </div>
            <div className={classNames(className, 'mx-2')}>
                <div className={styles.titleDescription}>
                    <h3>Fetching repositories...</h3>
                    <p className="text-muted mb-4">
                        We’re cloning your repos to Sourcegraph. In just a few moments, you can make your first search!
                    </p>
                </div>
                <div className="border overflow-hidden rounded">
                    <header>
                        <div className="py-4 px-3 d-flex justify-content-between align-items-center">
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
                    <Alert className="mt-4" variant="warning">
                        Cloning your repositories is taking a long time. You can wait for cloning to finish, or{' '}
                        <Link to={PageRoutes.Search} onClick={trackBannerClick}>
                            continue to Sourcegraph now
                        </Link>{' '}
                        while cloning continues in the background. Note that you can only search repos that have
                        finished cloning. Check status at any time in{' '}
                        <Link to="user/settings/repositories" onClick={trackBannerClick}>
                            Settings → Repositories
                        </Link>
                        .
                    </Alert>
                )}
            </div>
        </div>
    )
}

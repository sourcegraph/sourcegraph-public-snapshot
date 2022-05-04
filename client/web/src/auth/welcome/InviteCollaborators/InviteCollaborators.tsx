import React, { useEffect, useMemo, useState, useRef } from 'react'

import { ErrorLike } from '@sourcegraph/common'
import { Maybe } from '@sourcegraph/shared/src/graphql-operations'

import { AuthenticatedUser } from '../../../auth'
import { eventLogger } from '../../../tracking/eventLogger'
import { UserExternalServicesOrRepositoriesUpdateProps } from '../../../util'
import { getPostSignUpEvent, RepoSelectionMode, FinishWelcomeFlow } from '../../PostSignUpPage'
import { useExternalServicesWithCollaborators } from '../../useExternalServicesWithCollaborators'
import { useRepoCloningStatus } from '../../useRepoCloningStatus'
import { selectedReposVar, useSaveSelectedRepos, MinSelectedRepo } from '../../useSelectedRepos'
import { Footer } from '../Footer'

import { ActivityPane } from './ActivityPane'
import { InvitePane } from './InvitePane'

interface InviteCollaborators {
    user: AuthenticatedUser
    repoSelectionMode: RepoSelectionMode
    onUserExternalServicesOrRepositoriesUpdate: UserExternalServicesOrRepositoriesUpdateProps['onUserExternalServicesOrRepositoriesUpdate']
    setSelectedSearchContextSpec: (spec: string) => void
    onError: (error: ErrorLike) => void
    className?: string
    onFinish: FinishWelcomeFlow
}

export interface InvitableCollaborator {
    email: string
    displayName: string
    name: string
    avatarURL: Maybe<string>
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

export const InviteCollaborators: React.FunctionComponent<React.PropsWithChildren<InviteCollaborators>> = ({
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

    const invitableCollaborators = useMemo((): InvitableCollaborator[] => {
        const invitable: InvitableCollaborator[] = externalServices
            ? externalServices.flatMap(host => host.invitableCollaborators)
            : []
        const loggerPayload = {
            discovered: invitable.length,
        }
        eventLogger.log('UserInvitationsDiscoveredCollaborators', loggerPayload, loggerPayload)
        return invitable
    }, [externalServices])

    return (
        <div>
            <div className="d-flex flex-column flex-xl-row w-100">
                <InvitePane
                    user={user}
                    className={className}
                    invitableCollaborators={invitableCollaborators}
                    isLoadingCollaborators={isLoadingCollaborators}
                />
                <ActivityPane
                    className={className}
                    statusSummary={statusSummary}
                    isDoneCloning={isDoneCloning}
                    isLoading={isLoading}
                    cloningStatusLines={cloningStatusLines}
                    fetchError={fetchError}
                />
            </div>
            <Footer onFinish={onFinish} />
        </div>
    )
}

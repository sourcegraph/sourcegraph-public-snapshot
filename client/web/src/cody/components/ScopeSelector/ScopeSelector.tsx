import React, { useEffect, useMemo, useCallback } from 'react'

import classNames from 'classnames'

import type { Transcript } from '@sourcegraph/cody-shared/dist/chat/transcript'
import type { CodyClientScope } from '@sourcegraph/cody-shared/dist/chat/useClient'
import { useLazyQuery } from '@sourcegraph/http-client'
import { Text } from '@sourcegraph/wildcard'

import type { ReposStatusResult, ReposStatusVariables } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { EventName } from '../../../util/constants'

import { ReposStatusQuery } from './backend'
import { RepositoriesSelectorPopover, getFileName, type IRepo } from './RepositoriesSelectorPopover'

import styles from './ScopeSelector.module.scss'

export interface ScopeSelectorProps {
    scope: CodyClientScope
    setScope: (scope: CodyClientScope) => void
    toggleIncludeInferredRepository: () => void
    toggleIncludeInferredFile: () => void
    fetchRepositoryNames: (count: number) => Promise<string[]>
    isSourcegraphApp?: boolean
    transcript: Transcript | null
    className?: string
    renderHint?: (repos: IRepo[]) => React.ReactNode
}

export const ScopeSelector: React.FC<ScopeSelectorProps> = React.memo(function ScopeSelectorComponent({
    scope,
    setScope,
    toggleIncludeInferredRepository,
    toggleIncludeInferredFile,
    fetchRepositoryNames,
    isSourcegraphApp,
    transcript,
    className,
    renderHint,
}) {
    const [loadReposStatus, { data: newReposStatusData, previousData: previousReposStatusData }] = useLazyQuery<
        ReposStatusResult,
        ReposStatusVariables
    >(ReposStatusQuery, { fetchPolicy: 'cache-first' })

    const reposStatusData = newReposStatusData || previousReposStatusData

    const activeEditor = useMemo(() => scope.editor.getActiveTextEditor(), [scope.editor])

    useEffect(() => {
        const repoNames = [...scope.repositories]

        if (activeEditor?.repoName && !repoNames.includes(activeEditor.repoName)) {
            repoNames.push(activeEditor.repoName)
        }

        if (repoNames.length === 0) {
            return
        }

        loadReposStatus({
            variables: { repoNames, first: repoNames.length, includeJobs: !!window.context.currentUser?.siteAdmin },
            pollInterval: 2000,
        }).catch(() => null)
    }, [activeEditor, scope.repositories, loadReposStatus])

    const allRepositories = useMemo(() => reposStatusData?.repositories.nodes || [], [reposStatusData])

    const inferredRepository = useMemo(() => {
        if (activeEditor?.repoName) {
            return allRepositories.find(repo => repo.name === activeEditor.repoName) || null
        }

        return null
    }, [activeEditor, allRepositories])

    const additionalRepositories: IRepo[] = useMemo(
        () =>
            scope.repositories.reduce((repos, repoName) => {
                const repoStats = allRepositories.find(repo => repo.name === repoName)
                if (repoStats) {
                    repos.push(repoStats)
                }

                return repos
            }, [] as IRepo[]),
        [allRepositories, scope.repositories]
    )

    const addRepository = useCallback(
        (repoName: string) => {
            if (!scope.repositories.includes(repoName)) {
                eventLogger.log(EventName.CODY_CHAT_SCOPE_REPO_ADDED, { chatId: transcript?.id })
                setScope({ ...scope, repositories: [...scope.repositories, repoName] })
            }
        },
        [scope, setScope, transcript?.id]
    )

    const removeRepository = useCallback(
        (repoName: string) => {
            eventLogger.log(EventName.CODY_CHAT_SCOPE_REPO_REMOVED, { chatId: transcript?.id })
            setScope({ ...scope, repositories: scope.repositories.filter(repo => repo !== repoName) })
        },
        [scope, setScope, transcript?.id]
    )

    const resetScope = useCallback(async (): Promise<void> => {
        eventLogger.log(EventName.CODY_CHAT_SCOPE_RESET, { chatId: transcript?.id })
        if (!isSourcegraphApp) {
            return setScope({ ...scope, repositories: [], includeInferredRepository: true, includeInferredFile: true })
        }

        const repositories = await fetchRepositoryNames(10)
        return setScope({ ...scope, repositories, includeInferredRepository: true, includeInferredFile: true })
    }, [scope, setScope, fetchRepositoryNames, isSourcegraphApp, transcript?.id])

    return (
        <>
            <div className={classNames(styles.wrapper, className)}>
                <div className="d-flex text-truncate w-100">
                    <RepositoriesSelectorPopover
                        includeInferredRepository={scope.includeInferredRepository}
                        includeInferredFile={scope.includeInferredFile}
                        inferredRepository={inferredRepository}
                        inferredFilePath={activeEditor?.filePath || null}
                        additionalRepositories={additionalRepositories}
                        addRepository={addRepository}
                        resetScope={!isSourcegraphApp ? resetScope : null}
                        removeRepository={removeRepository}
                        toggleIncludeInferredRepository={toggleIncludeInferredRepository}
                        toggleIncludeInferredFile={toggleIncludeInferredFile}
                    />
                    {scope.includeInferredFile && activeEditor?.filePath && (
                        <Text size="small" className="ml-2 mb-0 align-self-center">
                            {getFileName(activeEditor.filePath)}
                        </Text>
                    )}
                </div>
            </div>
            {renderHint?.(additionalRepositories)}
        </>
    )
})

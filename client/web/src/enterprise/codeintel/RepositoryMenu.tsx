import React from 'react'

import { isErrorLike } from '@sourcegraph/common'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { RepositoryMenuContentProps } from '../../codeintel/RepositoryMenu'

import { useCodeIntelStatus } from './useCodeIntelStatus'

export const RepositoryMenuContent: React.FunctionComponent<RepositoryMenuContentProps> = props => {
    const { data, loading, error } = useCodeIntelStatus({
        variables: {
            repository: props.repoName,
            commit: props.revision,
            path: props.filePath,
        },
    })

    const forNerds =
        !isErrorLike(props.settingsCascade.final) &&
        props.settingsCascade.final?.experimentalFeatures?.codeIntelRepositoryBadge?.forNerds

    return loading ? (
        <LoadingSpinner />
    ) : error ? (
        <span>{error}</span>
    ) : data ? (
        <>
            <div className="px-2 py-1">
                <h2>Unimplemented</h2>

                <p className="text-muted">Unimplemented (enterprise version).</p>
            </div>

            {forNerds && <div className="px-2 py-1">NERD DATA</div>}
        </>
    ) : (
        <></>
    )
}

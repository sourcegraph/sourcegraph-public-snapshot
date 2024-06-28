import React, { type FC } from 'react'

import classNames from 'classnames'

import { CopyPathAction } from '@sourcegraph/branded/src/search-ui/components/CopyPathAction'
import { ResultContainer } from '@sourcegraph/branded/src/search-ui/components/ResultContainer'
import resultStyles from '@sourcegraph/branded/src/search-ui/components/ResultContainer.module.scss'
import { getFileMatchUrl, getRepositoryUrl, getRevision, type PathMatch } from '@sourcegraph/shared/src/search/stream'
import type { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { RepoFileLink } from './RepoFileLink'

export interface FilePathSearchResult extends SettingsCascadeProps {
    result: PathMatch
    repoDisplayName: string
    onSelect: () => void
    containerClassName?: string
    index: number
}

export const FilePathSearchResult: FC<FilePathSearchResult & TelemetryProps & TelemetryV2Props> = ({
    result,
    repoDisplayName,
    onSelect,
    containerClassName,
    index,
    telemetryService,
    telemetryRecorder,
    settingsCascade,
}) => {
    const repoAtRevisionURL = getRepositoryUrl(result.repository, result.branches)
    const revisionDisplayName = getRevision(result.branches, result.commit)

    const title = (
        <span className="d-flex align-items-center">
            <RepoFileLink
                repoName={result.repository}
                repoURL={repoAtRevisionURL}
                filePath={result.path}
                pathMatchRanges={result.pathMatches ?? []}
                fileURL={getFileMatchUrl(result)}
                repoDisplayName={
                    repoDisplayName
                        ? `${repoDisplayName}${revisionDisplayName ? `@${revisionDisplayName}` : ''}`
                        : undefined
                }
                className={resultStyles.titleInner}
                isKeyboardSelectable={true}
            />
            <CopyPathAction
                filePath={result.path}
                className={resultStyles.copyButton}
                telemetryService={telemetryService}
                telemetryRecorder={telemetryRecorder}
            />
        </span>
    )

    return (
        <ResultContainer
            index={index}
            title={title}
            resultType={result.type}
            onResultClicked={onSelect}
            repoName={result.repository}
            repoStars={result.repoStars}
            rankingDebug={result.debug}
            className={classNames(resultStyles.copyButtonContainer, containerClassName)}
            repoLastFetched={result.repoLastFetched}
        />
    )
}

import type { FC } from 'react'

import classNames from 'classnames'

import { getFileMatchUrl, getRepositoryUrl, getRevision, type PathMatch } from '@sourcegraph/shared/src/search/stream'
import type { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { CopyPathAction } from './CopyPathAction'
import { RepoFileLink } from './RepoFileLink'
import { ResultContainer } from './ResultContainer'
import { SearchResultPreviewButton } from './SearchResultPreviewButton'

import resultStyles from './ResultContainer.module.scss'

export interface FilePathSearchResult extends SettingsCascadeProps {
    result: PathMatch
    repoDisplayName: string
    onSelect: () => void
    containerClassName?: string
    index: number
    /**
     * Don't display the file preview button in the VSCode extension.
     * Expose this prop to allow the VSCode extension to hide the button.
     * Name it "hide" in an attempt to communicate that hiding is a special case.
     */
    hideFilePreviewButton?: boolean
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
    hideFilePreviewButton = false, // hiding the file preview button is a special case for the VSCode extension; we normally want it shown.
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
            actions={
                !hideFilePreviewButton ? (
                    <SearchResultPreviewButton
                        result={result}
                        telemetryService={telemetryService}
                        telemetryRecorder={telemetryRecorder}
                    />
                ) : undefined
            }
        />
    )
}

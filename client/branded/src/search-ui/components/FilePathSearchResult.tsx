import { FC } from 'react'

import classNames from 'classnames'

import { getFileMatchUrl, getRepositoryUrl, getRevision, type PathMatch } from '@sourcegraph/shared/src/search/stream'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { CopyPathAction } from './CopyPathAction'
import { RepoFileLink } from './RepoFileLink'
import { ResultContainer } from './ResultContainer'
import { SearchResultPreviewButton } from './SearchResultPreviewButton'

import styles from './SearchResult.module.scss'

export interface FilePathSearchResult extends SettingsCascadeProps {
    result: PathMatch
    repoDisplayName: string
    onSelect: () => void
    containerClassName?: string
    index: number
}

export const FilePathSearchResult: FC<FilePathSearchResult & TelemetryProps> = ({
    result,
    repoDisplayName,
    onSelect,
    containerClassName,
    index,
    telemetryService,
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
                className={styles.titleInner}
                isKeyboardSelectable={true}
            />
            <CopyPathAction filePath={result.path} className={styles.copyButton} telemetryService={telemetryService} />
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
            className={classNames(styles.copyButtonContainer, containerClassName)}
            repoLastFetched={result.repoLastFetched}
            actions={<SearchResultPreviewButton result={result} />}
        >
            <div className={classNames(styles.searchResultMatch, 'p-2')}>
                <small>{result.pathMatches ? 'Path match' : 'File contains matching content'}</small>
            </div>
        </ResultContainer>
    )
}

import { FC, KeyboardEvent, MouseEvent, useState, useCallback } from 'react'
import { createPlatformContext } from '../../../../platform/context'
import { Observable } from 'rxjs'
import classNames from 'classnames'
import { Link } from '@sourcegraph/wildcard'
import { map } from 'rxjs/operators'
import { HighlightResponseFormat } from '@sourcegraph/shared/src/graphql-operations'
import { useNavigate, useLocation } from 'react-router-dom'
import { CodeHostIcon, CodeExcerpt, onClickCodeExcerptHref } from '@sourcegraph/branded'
import { FetchFileParameters, fetchHighlightedFileLineRanges } from '@sourcegraph/shared/src/backend/file'
import { CommitMatch, getCommitMatchUrl, getRepositoryUrl } from '@sourcegraph/shared/src/search/stream'
import { parseSideBlobProps } from '../../../../codeintel/ReferencesPanel'
import { toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'
import { VulnerableCodeProps } from '../VulnerabilityCard/VulnerabilityCard'

import styles from './CodeBlock.module.scss'

interface CodeBlockProps {
    vulnerableCode: VulnerableCodeProps[]
}

export const CodeBlock: FC<CodeBlockProps> = ({ vulnerableCode }) => {
    const navigate = useNavigate()
    const navigateToUrl = (url: string): void => {
        navigate(url)
    }
    const [platformContext] = useState(() => createPlatformContext())
    const fetchHighlightedFileRangeLines = useCallback(
        (startLine: number, endLine: number): Observable<string[]> =>
            fetchHighlightedFileLineRanges(
                {
                    repoName: 'repo',
                    commitID: 'commitID',
                    filePath: 'file',
                    disableTimeout: false,
                    format: HighlightResponseFormat.HTML_HIGHLIGHT,
                    ranges: [
                        {
                            startLine: 1,
                            endLine: 1,
                        },
                    ],
                    platformContext,
                },
                false
            ).pipe(
                map(
                    lines =>
                        lines[
                            [
                                {
                                    startLine: 1,
                                    endLine: 1,
                                },
                            ].findIndex(group => group.startLine === startLine && group.endLine === endLine + 1)
                        ]
                )
            ),
        [fetchHighlightedFileLineRanges]
    )
    return (
        <div>
            <span
                className={
                    classNames()
                    // styles.titleInner,
                    // styles.mutedRepoFileLink
                }
            >
                <Link className={styles.urlLink} to={getRepositoryUrl(vulnerableCode[0].repository)}>
                    {getRepoNameWithIcon(vulnerableCode[0].repository)}
                </Link>
                <span aria-hidden={true}> ›</span>{' '}
                <Link className={styles.urlLink} to={getCommitMatchUrl(vulnerableCode[0])}>
                    {vulnerableCode[0].fileName}
                </Link>
            </span>
            <div className={styles.locationContainer}>
                <ul className="list-unstyled mb-0">
                    {vulnerableCode[0].group[0].locations.map((reference, index) => {
                        // const isActive = isActiveLocation(reference)
                        // const isFirstInActive =
                        //     isActive && !(index > 0 && isActiveLocation(group[0].locations[index - 1]))
                        // const locationActive = isActive ? styles.locationActive : ''
                        const selectReference = (event: KeyboardEvent<HTMLElement> | MouseEvent<HTMLElement>): void => {
                            onClickCodeExcerptHref(event, () => {
                                // if (isActive && activeURL) {
                                //     navigateToUrl(activeURL)
                                // } else {
                                //     setActiveLocation(reference)
                                // }

                                // setActiveLocation(reference)
                                // navigateToUrl(activeURL)
                                console.log('click click click')
                            })
                        }

                        return (
                            <li key={reference.url} className={classNames('border-0 rounded-0 mb-0', styles.location)}>
                                <div
                                    role="link"
                                    data-testid={`reference-item-${vulnerableCode[0].group[0].path}-${index}`}
                                    tabIndex={0}
                                    onClick={selectReference}
                                    onKeyDown={selectReference}
                                    data-href={reference.url}
                                    className={classNames(styles.locationLink)}
                                >
                                    <CodeExcerpt
                                        className={styles.locationLinkCodeExcerpt}
                                        commitID={reference.commitID}
                                        filePath={reference.file}
                                        repoName={reference.repo}
                                        highlightRanges={[
                                            {
                                                startLine: reference.range?.start.line ?? 0,
                                                startCharacter: reference.range?.start.character ?? 0,
                                                endLine: reference.range?.end.line ?? 0,
                                                endCharacter: reference.range?.end.character ?? 0,
                                            },
                                        ]}
                                        startLine={reference.range?.start.line ?? 0}
                                        endLine={reference.range?.end.line ?? 0}
                                        fetchHighlightedFileRangeLines={fetchHighlightedFileRangeLines}
                                        visibilityOffset={{ bottom: 0 }}
                                        // fetchPlainTextFileRangeLines={(): Observable<string[]> =>
                                        //     fetchPlainTextFileRangeLines(reference)
                                        // }
                                    />
                                </div>
                            </li>
                        )
                    })}
                </ul>
            </div>
        </div>
    )
}

export function displayRepoName(repoName: string): string {
    let parts = repoName.split('/')
    if (parts.length >= 3 && parts[0].includes('.')) {
        parts = parts.slice(1)
    }
    return parts.join('/')
}

function getRepoNameWithIcon(repoName: string): JSX.Element {
    const Icon = CodeHostIcon({
        repoName: repoName,
        className: styles.sidebarSectionIcon,
    })

    return (
        <span className={styles.sidebarSectionListItemBreakWords}>
            {Icon ? (
                <>
                    {Icon}
                    {displayRepoName(repoName)}
                </>
            ) : (
                repoName
            )}
        </span>
    )
}

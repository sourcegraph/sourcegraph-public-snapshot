import formatDistance from 'date-fns/formatDistance'
import * as H from 'history'
import SourceCommitIcon from 'mdi-react/SourceCommitIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../backend/graphqlschema'
import { DecoratedTextLines } from '../../components/DecoratedTextLines'
import { ResultContainer } from '../../components/ResultContainer'
import { AbsoluteRepoFilePosition, RepoSpec } from '../../repo'
import { GitRefTag } from '../../repo/GitRefTag'
import { RepoLink } from '../../repo/RepoLink'
import { eventLogger } from '../../tracking/eventLogger'
import { UserAvatar } from '../../user/UserAvatar'
import { toPrettyBlobURL } from '../../util/url'

interface Props {
    location: H.Location

    /**
     * The commit search result.
     */
    result: GQL.ICommitSearchResult

    /**
     * Called when the search result is selected.
     */
    onSelect: () => void

    /**
     * Whether this diff should be rendered as expanded.
     */
    expanded: boolean

    allExpanded?: boolean
}

export const CommitSearchResult: React.StatelessComponent<Props> = (props: Props) => {
    const telemetryData: { [key: string]: any } = {
        preview_type: props.result.diffPreview ? 'diff' : 'message',
    }
    const logClickOnPerson = () =>
        eventLogger.log('CommitSearchResultClicked', { commit_search_result: { ...telemetryData, target: 'person' } })
    const logClickOnMessage = () =>
        eventLogger.log('CommitSearchResultClicked', {
            commit_search_result: { ...telemetryData, target: 'message' },
        })
    const logClickOnTag = () =>
        eventLogger.log('CommitSearchResultClicked', {
            commit_search_result: { ...telemetryData, target: 'tag' },
        })
    const logClickOnCommitID = () =>
        eventLogger.log('CommitSearchResultClicked', {
            commit_search_result: { ...telemetryData, target: 'commit-id' },
        })
    const logClickOnTimestamp = () =>
        eventLogger.log('CommitSearchResultClicked', {
            commit_search_result: { ...telemetryData, target: 'timestamp' },
        })
    const logClickOnText = () =>
        eventLogger.log('CommitSearchResultClicked', {
            commit_search_result: { ...telemetryData, target: 'text' },
        })

    const title: React.ReactChild = (
        <div className="commit-search-result__title">
            <RepoLink
                repoPath={props.result.commit.repository.name}
                to={
                    props.result.commit.tree
                        ? props.result.commit.tree.canonicalURL
                        : props.result.commit.repository.url
                }
            />
            <Link
                to={props.result.commit.url}
                className="commit-search-result__title-person"
                onClick={stopPropagationToCollapseOrExpand}
                onMouseDown={logClickOnPerson}
            >
                <UserAvatar user={props.result.commit.author.person!} size={32} />{' '}
                {props.result.commit.author.person!.displayName}
            </Link>
            <Link
                to={props.result.commit.url}
                className="commit-search-result__title-message"
                onClick={stopPropagationToCollapseOrExpand}
                onMouseDown={logClickOnMessage}
            >
                {commitMessageSubject(props.result.commit.message) || '(empty commit message)'}
            </Link>
            <span className="commit-search-result__title-signature">
                {uniqueRefs([...props.result.refs, ...props.result.sourceRefs]).map((ref, i) => (
                    <GitRefTag key={i} gitRef={ref} onMouseDown={logClickOnTag} />
                ))}
                <code>
                    <Link
                        to={props.result.commit.url}
                        onClick={stopPropagationToCollapseOrExpand}
                        onMouseDown={logClickOnCommitID}
                    >
                        {props.result.commit.abbreviatedOID}
                    </Link>
                </code>{' '}
                <Link
                    to={props.result.commit.url}
                    onClick={stopPropagationToCollapseOrExpand}
                    onMouseDown={logClickOnTimestamp}
                >
                    {formatDistance(props.result.commit.author.date, new Date(), {
                        addSuffix: true,
                    })}
                </Link>
            </span>
        </div>
    )

    const expandedChildren: JSX.Element[] = []

    if (props.result.messagePreview) {
        expandedChildren.push(
            <DecoratedTextLines
                key="messagePreview"
                className="file-match__item commit-search-result__item commit-search-result__item-message-preview"
                value={props.result.messagePreview.value.split('\n')}
                highlights={props.result.messagePreview.highlights}
                lineClasses={[{ line: 1, className: 'strong' }]}
                onMouseDown={logClickOnText}
            />
        )
    }

    if (props.result.diffPreview) {
        const commonCtx: RepoSpec = {
            repoPath: props.result.commit.repository.name,
        }
        // lhsCtx and rhsCtx need the cast because their values at const init time lack
        // the filePath field, which is assigned as we iterate over the lines below.
        const lhsCtx = {
            ...commonCtx,
            commitID: props.result.commit.oid + '~',
            rev: props.result.commit.oid + '~',
        } as AbsoluteRepoFilePosition
        const rhsCtx = {
            ...commonCtx,
            commitID: props.result.commit.oid,
            rev: props.result.commit.oid,
        } as AbsoluteRepoFilePosition

        // Omit "index ", "--- file", and "+++ file" lines.
        const lines = props.result.diffPreview.value.split('\n')
        const lineClasses: { line: number; className: string; url?: string }[] = []
        let ignoreUntilAtAt = false
        for (const [i, line] of lines.entries()) {
            if (ignoreUntilAtAt && !line.startsWith('@@')) {
                lineClasses.push({ line: i + 1, className: 'hidden' })
                continue
            } else {
                ignoreUntilAtAt = false
            }
            if (line.startsWith('diff ')) {
                ignoreUntilAtAt = true

                // Simplify from "diff --git origname newname".
                const [origName, newName] = line.replace(/^diff --git /, '').split(' ')
                let simplerLine: string
                if (origName === newName) {
                    simplerLine = origName
                } else {
                    simplerLine = `${origName} -> ${newName} (renamed)`
                }
                lines[i] = simplerLine

                lhsCtx.filePath = origName
                rhsCtx.filePath = newName
                lineClasses.push({ line: i + 1, className: 'file-header', url: toPrettyBlobURL(rhsCtx) })
            } else if (line.startsWith('@@')) {
                // TODO(sqs): a bit hacky getting the position
                try {
                    const m = line.match(/^@@ -(\d+),.*\+(\d+)/)
                    if (m) {
                        if (m[1]) {
                            const lhsLine = parseInt(m[1], 10)
                            lhsCtx.position = { line: lhsLine - 1, character: 0 }
                        }
                        if (m[2]) {
                            const rhsLine = parseInt(m[2], 10)
                            rhsCtx.position = { line: rhsLine - 1, character: 0 }
                        }
                    }
                } catch (err) {
                    // TODO(sqs)
                    console.error(err)
                }

                lineClasses.push({ line: i + 1, className: 'hunk-header', url: toPrettyBlobURL(rhsCtx) })
            } else {
                if (rhsCtx.position && rhsCtx.position.line) {
                    if (!line.startsWith('+')) {
                        lhsCtx.position.line++
                    }
                    if (!line.startsWith('-')) {
                        rhsCtx.position.line++
                    }
                }

                if (line.startsWith('+')) {
                    lineClasses.push({ line: i + 1, className: 'added', url: toPrettyBlobURL(rhsCtx) })
                } else if (line.startsWith('-')) {
                    lineClasses.push({
                        line: i + 1,
                        className: 'deleted',
                        url: toPrettyBlobURL(lhsCtx),
                    })
                }
            }
        }

        expandedChildren.push(
            <DecoratedTextLines
                key="diffPreview"
                className="file-match__item commit-search-result__item commit-search-result__item-diff-preview"
                value={lines}
                highlights={props.result.diffPreview.highlights}
                lineClasses={lineClasses}
                onMouseDown={logClickOnText}
            />
        )
    }

    return (
        <ResultContainer
            collapsible={true}
            defaultExpanded={true}
            icon={SourceCommitIcon}
            title={title}
            expandedChildren={expandedChildren}
        />
    )
}

function commitMessageSubject(message: string): string {
    const eol = message.indexOf('\n')
    return eol === -1 ? message : message.slice(0, eol)
}

function stopPropagationToCollapseOrExpand(e: React.MouseEvent<HTMLElement>): void {
    e.stopPropagation()
}

function uniqueRefs(refs: GQL.IGitRef[]): GQL.IGitRef[] {
    const seenName = new Set<string>()
    const uniq: GQL.IGitRef[] = []
    for (const ref of refs) {
        if (!seenName.has(ref.name)) {
            uniq.push(ref)
            seenName.add(ref.name)
        }
    }
    return uniq
}

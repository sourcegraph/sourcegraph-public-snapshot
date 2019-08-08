import H from 'history'
import { isEqual } from 'lodash'
import React, { useCallback, useEffect, useState } from 'react'
import { Subscription } from 'rxjs'
import { catchError, map, startWith } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { Action } from '../../../../../../shared/src/api/types/action'
import { fromDiagnostic, toDiagnostic } from '../../../../../../shared/src/api/types/diagnostic'
import { CodeExcerpt } from '../../../../../../shared/src/components/CodeExcerpt'
import { LinkOrSpan } from '../../../../../../shared/src/components/LinkOrSpan'
import { displayRepoName } from '../../../../../../shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { propertyIsDefined } from '../../../../../../shared/src/util/types'
import { parseRepoURI } from '../../../../../../shared/src/util/url'
import { DiagnosticSeverityIcon } from '../../../../diagnostics/components/DiagnosticSeverityIcon'
import { fetchHighlightedFileLines } from '../../../../repo/backend'
import { ThemeProps } from '../../../../theme'
import { ActionsWithPreview } from '../../../actions/ActionsWithPreview'
import { ChangesetPlanOperation } from '../../../changesetsOLD/plan/plan'
import { ADD_DIAGNOSTIC_TO_THREAD_COMMAND } from '../../../threads/contributions/AddDiagnosticToThreadAction'
import { DiagnosticInfo, getCodeActions } from '../../../threadsOLD/detail/backend'

const LOADING: 'loading' = 'loading'

interface Props extends ExtensionsControllerProps, PlatformContextProps, ThemeProps {
    diagnostic: DiagnosticInfo
    selectedAction: Pick<ChangesetPlanOperation, 'editCommand'> | null
    onActionSelect: (diagnostic: DiagnosticInfo, action: Action | null) => void

    className?: string
    headerClassName?: string
    headerStyle?: React.CSSProperties
    history: H.History
    location: H.Location
}

/**
 * An item in a task list.
 */
export const DiagnosticsListItem: React.FunctionComponent<Props> = ({
    diagnostic,
    selectedAction,
    onActionSelect: onDiagnosticActionSelect,
    className = '',
    headerClassName = '',
    headerStyle,
    isLightTheme,
    extensionsController,
    ...props
}) => {
    const [actionsOrError, setActionsOrError] = useState<typeof LOADING | Action[] | ErrorLike>(LOADING)

    // Reduce recomputation of code actions when the diagnostic object reference changes but it
    // contains the same data.
    const diagnosticData = JSON.stringify(fromDiagnostic(diagnostic))
    useEffect(() => {
        const diagnostic2: DiagnosticInfo = {
            ...toDiagnostic(JSON.parse(diagnosticData)),
            type: diagnostic.type,
            entry: diagnostic.entry,
        }
        const subscriptions = new Subscription()
        subscriptions.add(
            getCodeActions({ diagnostic: diagnostic2, extensionsController })
                .pipe(
                    map(actions => [
                        ...actions,
                        // Common actions
                        {
                            title: 'Add to thread',
                            command: { ...ADD_DIAGNOSTIC_TO_THREAD_COMMAND, arguments: [diagnostic2] },
                        },
                    ]),
                    catchError(err => [asError(err)]),
                    startWith(LOADING)
                )
                .subscribe(setActionsOrError)
        )
        return () => subscriptions.unsubscribe()
    }, [diagnosticData, diagnostic.entry, extensionsController, diagnostic.type])

    const onActionSelect = useCallback(
        (action: Action | null) => {
            onDiagnosticActionSelect(diagnostic, action)
        },
        [diagnostic, onDiagnosticActionSelect]
    )

    return (
        <ActionsWithPreview
            {...props}
            actionsOrError={actionsOrError}
            selectedAction={
                (selectedAction &&
                    actionsOrError !== LOADING &&
                    !isErrorLike(actionsOrError) &&
                    actionsOrError
                        .filter(propertyIsDefined('computeEdit'))
                        .find(a => commandIsEqual(a.computeEdit, selectedAction.editCommand))) ||
                null
            }
            onActionSelect={onActionSelect}
            diagnostic={fromDiagnostic(diagnostic)}
            extensionsController={extensionsController}
            defaultPreview={
                diagnostic.range && (
                    <>
                        <CodeExcerpt
                            repoName={diagnostic.entry.repository.name}
                            commitID={diagnostic.entry.commit.oid}
                            filePath={diagnostic.entry.path}
                            context={4}
                            highlightRanges={[diagnostic.range].map(r => ({
                                line: r.start.line,
                                character: r.start.character,
                                highlightLength: r.end.character - r.start.character,
                            }))}
                            className="d-block w-100 h-100 overflow-auto p-2"
                            isLightTheme={isLightTheme}
                            fetchHighlightedFileLines={fetchHighlightedFileLines}
                        />
                        {diagnostic.relatedInformation &&
                            diagnostic.relatedInformation.map((info, i) => (
                                <div key={i} className="mt-4">
                                    <h6 className="d-flex align-items-center font-weight-normal mb-0 ml-2">
                                        <DiagnosticSeverityIcon severity={diagnostic.severity} className="small mr-2" />{' '}
                                        {info.message} ({parseRepoURI(info.location.uri.toString()).filePath})
                                    </h6>
                                    <CodeExcerpt
                                        repoName={diagnostic.entry.repository.name} // TODO!(sqs): not always the same
                                        commitID={diagnostic.entry.commit.oid} // TODO!(sqs): not always the same
                                        filePath={info.location.uri.toString().replace(/.*#/, '')} // TODO!(sqs): hack
                                        context={Math.ceil(
                                            (info.location.range!.end.line - info.location.range!.start.line) / 2
                                        )}
                                        highlightRanges={[
                                            {
                                                // TODO!(sqs): remove '!' non-null assertions
                                                line: Math.ceil(
                                                    (info.location.range!.start.line + info.location.range!.end.line) /
                                                        2
                                                ),
                                                character: info.location.range!.start.character,
                                                highlightLength:
                                                    info.location.range!.end.character -
                                                    info.location.range!.start.character,
                                            },
                                        ]}
                                        className="w-100 h-100 overflow-auto p-2 d-block"
                                        isLightTheme={isLightTheme}
                                        fetchHighlightedFileLines={fetchHighlightedFileLines}
                                    />
                                </div>
                            ))}
                    </>
                )
            }
        >
            {({ actions, preview }) => (
                <div className={`d-flex flex-wrap align-items-stretch ${className}`}>
                    <div style={{ flex: '1 1 40%', minWidth: '400px' }} className="pr-5">
                        <header className={`d-flex align-items-start ${headerClassName}`} style={headerStyle}>
                            <div className={`flex-1 d-flex align-items-center`}>
                                <h3 className="mb-0 small">
                                    <LinkOrSpan to={diagnostic.entry.url} className="d-block">
                                        {diagnostic.entry.path ? (
                                            <>
                                                <span className="font-weight-normal">
                                                    {displayRepoName(diagnostic.entry.repository.name)}
                                                </span>{' '}
                                                › {diagnostic.entry.path}
                                            </>
                                        ) : (
                                            displayRepoName(diagnostic.entry.repository.name)
                                        )}
                                    </LinkOrSpan>
                                </h3>
                            </div>
                        </header>
                        <div className={`d-flex align-items-start mt-2 mb-1`}>
                            <DiagnosticSeverityIcon severity={diagnostic.severity} className="icon-inline mr-2" />
                            <span>{diagnostic.message}</span>
                        </div>
                        {actions}
                    </div>
                    {preview && (
                        <aside
                            className="d-flex flex-column justify-content-between"
                            style={{
                                flex: '2 0 60%',
                                minWidth: '600px',
                                margin: '-0.5rem -1rem -0.5rem 0',
                            }}
                        >
                            {preview}
                        </aside>
                    )}
                </div>
            )}
        </ActionsWithPreview>
    )
}

function commandIsEqual(a: sourcegraph.Command, b: sourcegraph.Command): boolean {
    return a.command === b.command && isEqual(a.arguments || [], b.arguments || [])
}

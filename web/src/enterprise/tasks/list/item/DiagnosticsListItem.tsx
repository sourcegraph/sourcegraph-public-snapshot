import H from 'history'
import { isEqual } from 'lodash'
import React, { useCallback, useEffect, useState } from 'react'
import { Subscription } from 'rxjs'
import { catchError, startWith } from 'rxjs/operators'
import { Action } from '../../../../../../shared/src/api/types/action'
import { fromDiagnostic, toDiagnostic } from '../../../../../../shared/src/api/types/diagnostic'
import { CodeExcerpt } from '../../../../../../shared/src/components/CodeExcerpt'
import { LinkOrSpan } from '../../../../../../shared/src/components/LinkOrSpan'
import { displayRepoName } from '../../../../../../shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { DiagnosticSeverityIcon } from '../../../../diagnostics/components/DiagnosticSeverityIcon'
import { fetchHighlightedFileLines } from '../../../../repo/backend'
import { ActionsWithPreview } from '../../../actions/ActionsWithPreview'
import { DiagnosticInfo, getCodeActions } from '../../../threads/detail/backend'

const LOADING: 'loading' = 'loading'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    diagnostic: DiagnosticInfo
    selectedAction: Action | null // TODO!(sqs): isnt reference-equal to the action in actionsOrError
    onActionSelect: (diagnostic: DiagnosticInfo, action: Action | null) => void

    className?: string
    headerClassName?: string
    headerStyle?: React.CSSProperties
    isLightTheme: boolean
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
                (actionsOrError !== LOADING &&
                    !isErrorLike(actionsOrError) &&
                    actionsOrError.find(a => isEqual(a, selectedAction))) ||
                null
            }
            onActionSelect={onActionSelect}
            extensionsController={extensionsController}
            defaultPreview={
                diagnostic.range && (
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
                        className="w-100 h-100 overflow-auto p-2"
                        isLightTheme={isLightTheme}
                        fetchHighlightedFileLines={fetchHighlightedFileLines}
                    />
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

import AJV from 'ajv'
import addFormats from 'ajv-formats'
import { load as loadYAML } from 'js-yaml'
import CloseIcon from 'mdi-react/CloseIcon'
import WarningIcon from 'mdi-react/WarningIcon'
import React, { useState, useCallback, useEffect, useMemo } from 'react'
import { asyncScheduler, concat, Observable, of, OperatorFunction, SchedulerLike, Subject } from 'rxjs'
import {
    catchError,
    debounceTime,
    delay,
    distinctUntilChanged,
    map,
    mergeMap,
    publish,
    repeatWhen,
    startWith,
    switchMap,
    take,
    takeWhile,
    tap,
} from 'rxjs/operators'

import { isErrorLike } from '@sourcegraph/codeintellify/lib/errors'
import { BatchSpecWorkspaceResolutionState } from '@sourcegraph/shared/src/graphql-operations'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { asError, ErrorLike } from '@sourcegraph/shared/src/util/errors'
import { pluralize } from '@sourcegraph/shared/src/util/strings'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Container, LoadingSpinner } from '@sourcegraph/wildcard'

import batchSpecSchemaJSON from '../../../../../../../schema/batch_spec.schema.json'
import { ErrorAlert } from '../../../../components/alerts'
import { BatchSpecWorkspacesFields, Scalars } from '../../../../graphql-operations'
import { BatchSpec } from '../../../../schema/batch_spec.schema'
import { BatchSpecDownloadLink, getFileName } from '../../BatchSpec'
import { excludeRepo } from '../yaml-util'

import { createBatchSpecFromRaw, fetchBatchSpec, replaceBatchSpecInput } from './backend'
import { MonacoSettingsEditor } from './MonacoBatchSpecEditor'

export interface Spec {
    fileName: string
    code: string
}

interface ExampleTabsProps extends ThemeProps {
    updateSpec: (spec: Spec) => void
    setPreviewID: (id: Scalars['ID']) => void
}

export const ExampleTabs: React.FunctionComponent<ExampleTabsProps> = ({ isLightTheme, updateSpec, setPreviewID }) => (
    <ExampleTabPanel isLightTheme={isLightTheme} updateSpec={updateSpec} setPreviewID={setPreviewID} />
)

interface ExampleTabPanelProps extends ThemeProps {
    updateSpec: (spec: Spec) => void
    setPreviewID: (id: Scalars['ID']) => void
}

const ExampleTabPanel: React.FunctionComponent<ExampleTabPanelProps> = ({ isLightTheme, updateSpec, setPreviewID }) => {
    const [code, setCode] = useState<string>('name:')
    const [codeUpdateError, setCodeUpdateError] = useState<string>()

    // Updates the batch spec code when the user wants to exclude a repo resolved in the
    // workspaces preview.
    const excludeRepoFromSpec = useCallback(
        (repo: string, branch: string) => {
            setCodeUpdateError(undefined)

            const result = excludeRepo(code, repo, branch)

            if (result.success) {
                setCode(result.spec)
            } else {
                setCodeUpdateError(
                    'Unable to update batch spec. Double-check to make sure there are no syntax errors, then try again.'
                )
            }
        },
        [code]
    )

    const codeUpdates = useMemo(() => new Subject<string>(), [])

    useEffect(() => {
        codeUpdates.next(code)
    }, [codeUpdates, code])

    const [previewStale, setPreviewStale] = useState<boolean>(true)

    const specValidator = useMemo(() => {
        const ajv = new AJV()
        addFormats(ajv)
        return ajv.compile<BatchSpec>(batchSpecSchemaJSON)
    }, [])

    const [invalid, setInvalid] = useState<boolean>(false)

    const preview = useObservable(
        useMemo(
            () => {
                const initialFetchRunning = false
                let initialFetchCompleted = false
                return codeUpdates.pipe(
                    startWith(code),
                    distinctUntilChanged(),
                    tap(() => {
                        setPreviewStale(true)
                    }),
                    debounceTimeAfterFirst(250),
                    map(code => {
                        try {
                            const parsedDocument = loadYAML(code)
                            const valid = specValidator(parsedDocument)
                            setInvalid(!valid)
                        } catch {
                            setInvalid(true)
                        }
                        return code
                    }),
                    switchMap(code => {
                        let specCreator: Observable<BatchSpecWorkspacesFields>
                        if (preview !== undefined && !isErrorLike(preview)) {
                            specCreator = replaceBatchSpecInput(preview.id, code)
                        } else {
                            specCreator = createBatchSpecFromRaw(code).pipe(
                                tap(() => {
                                    initialFetchCompleted = true
                                })
                            )
                        }
                        return specCreator.pipe(
                            switchMap(spec =>
                                concat(
                                    of(spec),
                                    fetchBatchSpec(spec.id).pipe(
                                        // Poll the batch spec until resolution is complete or failed.
                                        repeatWhen(completed => completed.pipe(delay(500))),
                                        takeWhile(
                                            response =>
                                                [
                                                    BatchSpecWorkspaceResolutionState.PROCESSING,
                                                    BatchSpecWorkspaceResolutionState.QUEUED,
                                                ].includes(response.workspaceResolution?.state as any),
                                            true
                                        )
                                    )
                                )
                            ),
                            catchError(error => [asError(error)])
                        )
                    }),
                    tap(() => {
                        setPreviewStale(false)
                    }),
                    catchError(error => [asError(error)])
                )
            },
            // Don't want to trigger on changes to code, it's just the initial value.
            // eslint-disable-next-line react-hooks/exhaustive-deps
            [codeUpdates]
        )
    )

    useEffect(() => {
        if (preview && !isErrorLike(preview)) {
            setPreviewID(preview.id)
        }
    }, [preview, setPreviewID])

    // Update the spec in parent state whenever the code changes
    useEffect(() => {
        updateSpec({ code, fileName: getFileName('whatever') })
    }, [code, updateSpec])

    console.log(specValidator.errors)

    return (
        <>
            <div className="d-flex justify-content-end align-items-center mb-2">
                <BatchSpecDownloadLink name="whatever" originalInput={code} />
            </div>
            <Container className="mb-3">
                <MonacoSettingsEditor
                    isLightTheme={isLightTheme}
                    language="yaml"
                    value={code}
                    jsonSchema={batchSpecSchemaJSON}
                    onChange={setCode}
                />
            </Container>
            <Container>
                {codeUpdateError && <ErrorAlert error={codeUpdateError} />}
                {invalid && specValidator.errors && (
                    <ErrorAlert error={`The entered spec is invalid ${specValidator.errors}`} />
                )}
                <PreviewWorkspaces excludeRepo={excludeRepoFromSpec} preview={preview} previewStale={previewStale} />
            </Container>
        </>
    )
}

interface PreviewWorkspacesProps {
    excludeRepo: (repo: string, branch: string) => void
    preview: BatchSpecWorkspacesFields | Error | undefined
    previewStale: boolean
}

const PreviewWorkspaces: React.FunctionComponent<PreviewWorkspacesProps> = ({ excludeRepo, preview, previewStale }) => {
    if (!preview || previewStale) {
        return <LoadingSpinner />
    }
    if (isErrorLike(preview)) {
        return <ErrorAlert error={preview} className="mb-0" />
    }
    if (!preview.workspaceResolution) {
        throw new Error('Expected workspace resolution to exist.')
    }
    return (
        <>
            <h3>
                Preview workspaces ({preview.workspaceResolution.workspaces.nodes.length})
                {preview.workspaceResolution.state === BatchSpecWorkspaceResolutionState.QUEUED && (
                    <LoadingSpinner className="icon-inline" />
                )}
                {preview.workspaceResolution.state === BatchSpecWorkspaceResolutionState.PROCESSING && (
                    <LoadingSpinner className="icon-inline" />
                )}
                {preview.workspaceResolution.state === BatchSpecWorkspaceResolutionState.ERRORED && (
                    <WarningIcon className="text-danger icon-inline" />
                )}
                {preview.workspaceResolution.state === BatchSpecWorkspaceResolutionState.FAILED && (
                    <WarningIcon className="text-danger icon-inline" />
                )}
            </h3>
            {preview.workspaceResolution.failureMessage !== null && (
                <ErrorAlert error={preview.workspaceResolution.failureMessage} />
            )}
            <p className="text-monospace">
                allowUnsupported: {JSON.stringify(preview.allowUnsupported)}
                <br />
                allowIgnored: {JSON.stringify(preview.allowIgnored)}
            </p>
            <ul className="list-group p-1 mb-0">
                {preview.workspaceResolution.workspaces.nodes.map(item => (
                    <li
                        className="d-flex border-bottom mb-3"
                        key={`${item.repository.id}_${item.branch.target.oid}_${item.path || '/'}`}
                    >
                        <button
                            className="btn align-self-start p-0 m-0 mr-3"
                            data-tooltip="Omit this repository from batch spec file"
                            type="button"
                            // TODO: Alert that for monorepos, we will exclude all paths
                            onClick={() => excludeRepo(item.repository.name, item.branch.displayName)}
                        >
                            <CloseIcon className="icon-inline" />
                        </button>
                        <div className="mb-2 flex-1">
                            <p>
                                {item.repository.name}:{item.branch.abbrevName} Path: {item.path || '/'}
                            </p>
                            <p>
                                {item.searchResultPaths.length} {pluralize('result', item.searchResultPaths.length)}
                            </p>
                        </div>
                    </li>
                ))}
            </ul>
            {preview.workspaceResolution.workspaces.nodes.length === 0 && (
                <span className="text-muted">No workspaces found</span>
            )}
        </>
    )
}

export function debounceTimeAfter<T>(
    amount: number,
    dueTime: number,
    scheduler: SchedulerLike = asyncScheduler
): OperatorFunction<T, T> {
    return publish(value => concat(value.pipe(take(amount)), value.pipe(debounceTime(dueTime, scheduler))))
}

export function debounceTimeAfterFirst<T>(
    dueTime: number,
    scheduler: SchedulerLike = asyncScheduler
): OperatorFunction<T, T> {
    return debounceTimeAfter(1, dueTime, scheduler)
}

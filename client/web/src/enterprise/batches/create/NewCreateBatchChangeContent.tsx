import AJV from 'ajv'
import addFormats from 'ajv-formats'
import classNames from 'classnames'
import { load as loadYAML } from 'js-yaml'
import CloseIcon from 'mdi-react/CloseIcon'
import WarningIcon from 'mdi-react/WarningIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useHistory } from 'react-router'
import { asyncScheduler, concat, Observable, of, OperatorFunction, SchedulerLike, Subject } from 'rxjs'
import {
    catchError,
    debounceTime,
    delay,
    distinctUntilChanged,
    map,
    publish,
    repeatWhen,
    startWith,
    switchMap,
    take,
    takeWhile,
    tap,
} from 'rxjs/operators'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { BatchSpecWorkspaceResolutionState, Scalars } from '@sourcegraph/shared/src/graphql-operations'
import {
    SettingsCascadeOrError,
    SettingsCascadeProps,
    SettingsOrgSubject,
    SettingsSubject,
    SettingsUserSubject,
} from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { pluralize } from '@sourcegraph/shared/src/util/strings'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Container, LoadingSpinner } from '@sourcegraph/wildcard'

import batchSpecSchemaJSON from '../../../../../../schema/batch_spec.schema.json'
import { ErrorAlert } from '../../../components/alerts'
import { BatchSpecWorkspacesFields } from '../../../graphql-operations'
import { BatchSpec } from '../../../schema/batch_spec.schema'
import { Settings } from '../../../schema/settings.schema'

import { createBatchSpecFromRaw, executeBatchSpec, fetchBatchSpec, replaceBatchSpecInput } from './backend'
import { MonacoBatchSpecEditor } from './editor/MonacoBatchSpecEditor'
import styles from './NewCreateBatchChangeContent.module.scss'
import { excludeRepo } from './yaml-util'

interface CreateBatchChangePageProps extends ThemeProps, SettingsCascadeProps<Settings> {}

export const NewCreateBatchChangeContent: React.FunctionComponent<CreateBatchChangePageProps> = ({
    isLightTheme,
    settingsCascade,
}) => {
    const history = useHistory()

    const [isLoading, setIsLoading] = useState<boolean | Error>(false)
    const [selectedNamespace, setSelectedNamespace] = useState<string>('')
    const [previewID, setPreviewID] = useState<Scalars['ID']>()
    const [code, setCode] = useState<string>('name: ')

    const submitBatchSpec = useCallback<React.MouseEventHandler>(async () => {
        if (!previewID) {
            return
        }
        setIsLoading(true)
        try {
            const execution = await executeBatchSpec(previewID)
            history.push(`${execution.namespace.url}/batch-changes/executions/${execution.id}`)
        } catch (error) {
            setIsLoading(error)
        }
    }, [previewID, history])

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
                    'Unable to update batch spec. Double-check to make sure there are no syntax errors, then try again.' +
                        result.error
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
    const [invalid, setInvalid] = useState<boolean>(false)

    const specValidator = useMemo(() => {
        const ajv = new AJV()
        addFormats(ajv)
        return ajv.compile<BatchSpec>(batchSpecSchemaJSON)
    }, [])

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
                                                ].includes(response.workspaceResolution?.state),
                                            true
                                        )
                                    )
                                )
                            ),
                            catchError(error => [asError(error)])
                        )
                    }),
                    tap(preview => {
                        setPreviewStale(false)
                        if (!isErrorLike(preview)) {
                            setPreviewID(preview.id)
                        } else {
                            setPreviewID(undefined)
                        }
                    }),
                    catchError(error => [asError(error)])
                )
            },
            // Don't want to trigger on changes to code, it's just the initial value.
            // eslint-disable-next-line react-hooks/exhaustive-deps
            [codeUpdates]
        )
    )

    return (
        <>
            <h3>1. Select a namespace</h3>
            <NamespaceSelector
                settingsCascade={settingsCascade}
                selectedNamespace={selectedNamespace}
                onSelect={setSelectedNamespace}
            />
            <h3>2. Write a batch spec</h3>
            <p>
                The batch spec describes what a batch change should do. Choose an example template and edit it here. See
                the{' '}
                <a
                    href="https://docs.sourcegraph.com/batch_changes/references/batch_spec_yaml_reference"
                    rel="noopener noreferrer"
                    target="_blank"
                >
                    syntax reference
                </a>{' '}
                for more options.
            </p>
            <div className="row">
                <div className="col-8">
                    <EditSpecSection isLightTheme={isLightTheme} code={code} setCode={setCode} />
                </div>
                <div className="col-4">
                    <Container>
                        {codeUpdateError && <ErrorAlert error={codeUpdateError} />}
                        {invalid && specValidator.errors && (
                            <ErrorAlert error={`The entered spec is invalid ${specValidator.errors}`} />
                        )}
                        <PreviewWorkspaces
                            excludeRepo={excludeRepoFromSpec}
                            preview={preview}
                            previewStale={previewStale}
                        />
                    </Container>
                </div>
            </div>
            <h3 className="mt-4">
                3. Execute the batch spec <span className="badge badge-info text-uppercase ml-1">Experimental</span>
            </h3>
            <p>Execute the batch spec to get a preview of your batch change before publishing the results.</p>
            <p>
                Use{' '}
                <a
                    href="https://docs.sourcegraph.com/@batch-executor-mvp-docs/batch_changes/explanations/executors"
                    rel="noopener noreferrer"
                    target="_blank"
                >
                    Sourcegraph executors
                </a>{' '}
                to run the spec and preview your batch change.
            </p>
            <button
                type="button"
                className="btn btn-primary mb-3"
                onClick={submitBatchSpec}
                disabled={isLoading === true}
            >
                Run batch spec
            </button>
            {isErrorLike(isLoading) && <ErrorAlert error={isLoading} />}
        </>
    )
}

const NAMESPACE_SELECTOR_ID = 'batch-spec-execution-namespace-selector'

interface NamespaceSelectorProps extends SettingsCascadeProps<Settings> {
    selectedNamespace: string
    onSelect: (namespace: string) => void
}

const NamespaceSelector: React.FunctionComponent<NamespaceSelectorProps> = ({
    onSelect,
    selectedNamespace,
    settingsCascade,
}) => {
    const namespaces: SettingsSubject[] = useMemo(() => namespacesFromSettings(settingsCascade), [settingsCascade])

    const userNamespace = useMemo(
        () => namespaces.find((namespace): namespace is SettingsUserSubject => namespace.__typename === 'User'),
        [namespaces]
    )

    if (!userNamespace) {
        throw new Error('No user namespace found')
    }

    const organizationNamespaces = useMemo(
        () => namespaces.filter((namespace): namespace is SettingsOrgSubject => namespace.__typename === 'Org'),
        [namespaces]
    )

    // Set the initially-selected namespace to the user's namespace
    useEffect(() => onSelect(userNamespace.id), [onSelect, userNamespace])

    const onSelectNamespace = useCallback<React.ChangeEventHandler<HTMLSelectElement>>(
        event => {
            onSelect(event.target.value)
        },
        [onSelect]
    )

    return (
        <div className="form-group d-flex flex-column justify-content-start">
            <label className="text-nowrap mr-2 mb-0" htmlFor={NAMESPACE_SELECTOR_ID}>
                <strong>Select namespace:</strong>
            </label>
            <select
                className={classNames(styles.namespaceSelector, 'form-control')}
                id={NAMESPACE_SELECTOR_ID}
                value={selectedNamespace}
                onChange={onSelectNamespace}
            >
                {/* Put the user namespace first. */}
                <option value={userNamespace.id}>{userNamespace.displayName ?? userNamespace.username}</option>
                {organizationNamespaces.map(namespace => (
                    <option key={namespace.id} value={namespace.id}>
                        {namespace.displayName ?? namespace.name}
                    </option>
                ))}
            </select>
        </div>
    )
}

function namespacesFromSettings(settingsCascade: SettingsCascadeOrError<Settings>): SettingsSubject[] {
    return (
        (settingsCascade !== null &&
            !isErrorLike(settingsCascade) &&
            settingsCascade.subjects !== null &&
            settingsCascade.subjects.map(({ subject }) => subject).filter(subject => !isErrorLike(subject))) ||
        []
    )
}

interface EditSpecSectionProps extends ThemeProps {
    code: string
    setCode: (spec: string) => void
}

const EditSpecSection: React.FunctionComponent<EditSpecSectionProps> = ({ isLightTheme, code, setCode }) => (
    <MonacoBatchSpecEditor isLightTheme={isLightTheme} value={code} onChange={setCode} />
)

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
            <h3>Workspaces preview</h3>
            {preview.workspaceResolution.failureMessage !== null && (
                <ErrorAlert error={preview.workspaceResolution.failureMessage} />
            )}
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
            {preview.importingChangesets?.totalCount > 0 && (
                <>
                    <h3>Importing changesets</h3>
                    <ul>
                        {preview.importingChangesets?.nodes.map(node => (
                            <li key={node.id}>
                                <Link
                                    to={
                                        node.__typename === 'VisibleChangesetSpec' &&
                                        node.description.__typename === 'ExistingChangesetReference' &&
                                        node.description.baseRepository.url
                                    }
                                >
                                    {node.__typename === 'VisibleChangesetSpec' &&
                                        node.description.__typename === 'ExistingChangesetReference' &&
                                        node.description.baseRepository.name}
                                </Link>{' '}
                                #
                                {node.__typename === 'VisibleChangesetSpec' &&
                                    node.description.__typename === 'ExistingChangesetReference' &&
                                    node.description.externalID}
                            </li>
                        ))}
                    </ul>
                </>
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

import AJV from 'ajv'
import addFormats from 'ajv-formats'
import classNames from 'classnames'
import { load as loadYAML } from 'js-yaml'
import CloseIcon from 'mdi-react/CloseIcon'
import ContentSaveIcon from 'mdi-react/ContentSaveIcon'
import WarningIcon from 'mdi-react/WarningIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useHistory, useLocation } from 'react-router'
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

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
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
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { Container, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import batchSpecSchemaJSON from '../../../../../../schema/batch_spec.schema.json'
import { BatchChangesIcon } from '../../../batches/icons'
import { BatchSpecWithWorkspacesFields } from '../../../graphql-operations'
import { BatchSpec } from '../../../schema/batch_spec.schema'
import { Settings } from '../../../schema/settings.schema'
import { BatchSpecDownloadLink } from '../BatchSpec'

import {
    createBatchSpecFromRaw as _createBatchSpecFromRaw,
    executeBatchSpec,
    fetchBatchSpec,
    replaceBatchSpecInput,
} from './backend'
import { MonacoBatchSpecEditor } from './editor/MonacoBatchSpecEditor'
import helloWorldSample from './examples/hello-world.batch.yaml'
import styles from './NewCreateBatchChangePage.module.scss'
import { excludeRepo } from './yaml-util'

const getNamespaceDisplayName = (namespace: SettingsUserSubject | SettingsOrgSubject): string => {
    switch (namespace.__typename) {
        case 'User':
            return namespace.displayName ?? namespace.username
        case 'Org':
            return namespace.displayName ?? namespace.name
    }
}

/** TODO: This duplicates the URL field from the org/user resolvers on the backend, but we
 * don't have access to that from the settings cascade presently. Can we get it included
 * in the cascade instead somehow? */
const getNamespaceBatchChangesURL = (namespace: SettingsUserSubject | SettingsOrgSubject): string => {
    switch (namespace.__typename) {
        case 'User':
            return '/users/' + namespace.username + '/batch-changes'
        case 'Org':
            return '/organizations/' + namespace.name + '/batch-changes'
    }
}

interface CreateBatchChangePageProps extends ThemeProps, SettingsCascadeProps<Settings> {
    /* For testing only. */
    createBatchSpecFromRaw?: typeof _createBatchSpecFromRaw
}

export const NewCreateBatchChangePage: React.FunctionComponent<CreateBatchChangePageProps> = ({
    isLightTheme,
    settingsCascade,
    createBatchSpecFromRaw = _createBatchSpecFromRaw,
}) => {
    const history = useHistory()
    const location = useLocation()

    // Gather all the available namespaces from user settings
    const rawNamespaces: SettingsSubject[] = useMemo(() => namespacesFromSettings(settingsCascade), [settingsCascade])

    const userNamespace = useMemo(
        () => rawNamespaces.find((namespace): namespace is SettingsUserSubject => namespace.__typename === 'User'),
        [rawNamespaces]
    )

    if (!userNamespace) {
        throw new Error('No user namespace found')
    }

    const organizationNamespaces = useMemo(
        () => rawNamespaces.filter((namespace): namespace is SettingsOrgSubject => namespace.__typename === 'Org'),
        [rawNamespaces]
    )

    const namespaces: (SettingsUserSubject | SettingsOrgSubject)[] = useMemo(
        () => [userNamespace, ...organizationNamespaces],
        [userNamespace, organizationNamespaces]
    )

    // Check if there's a namespace parameter in the URL
    const defaultNamespace = new URLSearchParams(location.search).get('namespace')

    // The default namespace selected from the dropdown should match whatever was in the
    // URL parameter, or else default to the user's namespace
    const defaultSelectedNamespace = useMemo(() => {
        if (defaultNamespace) {
            const lowerCaseDefaultNamespace = defaultNamespace.toLowerCase()
            return (
                namespaces.find(
                    namespace =>
                        namespace.displayName?.toLowerCase() === lowerCaseDefaultNamespace ||
                        (namespace.__typename === 'User' &&
                            namespace.username.toLowerCase() === lowerCaseDefaultNamespace) ||
                        (namespace.__typename === 'Org' && namespace.name.toLowerCase() === lowerCaseDefaultNamespace)
                ) || userNamespace
            )
        }
        return userNamespace
    }, [namespaces, defaultNamespace, userNamespace])

    const [selectedNamespace, setSelectedNamespace] = useState<SettingsUserSubject | SettingsOrgSubject>(
        defaultSelectedNamespace
    )

    const [isLoading, setIsLoading] = useState<boolean | Error>(false)
    const [previewID, setPreviewID] = useState<Scalars['ID']>()
    const [code, setCode] = useState<string>(helloWorldSample)

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
            () =>
                codeUpdates.pipe(
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
                        let specCreator: Observable<BatchSpecWithWorkspacesFields>
                        if (preview !== undefined && !isErrorLike(preview)) {
                            specCreator = replaceBatchSpecInput(preview.id, code)
                        } else {
                            specCreator = createBatchSpecFromRaw(code, selectedNamespace.id)
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
                                                !!response.workspaceResolution &&
                                                (response.workspaceResolution.state ===
                                                    BatchSpecWorkspaceResolutionState.QUEUED ||
                                                    response.workspaceResolution.state ===
                                                        BatchSpecWorkspaceResolutionState.PROCESSING),
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
                ),
            // Don't want to trigger on changes to code, it's just the initial value.
            // eslint-disable-next-line react-hooks/exhaustive-deps
            [codeUpdates]
        )
    )

    return (
        <div className="d-flex flex-column p-4 w-100 h-100">
            <div className="d-flex flex-0 justify-content-between">
                <div className="flex-1">
                    <PageHeader
                        path={[
                            { icon: BatchChangesIcon },
                            {
                                to: getNamespaceBatchChangesURL(selectedNamespace),
                                text: getNamespaceDisplayName(selectedNamespace),
                            },
                            { text: 'Create batch change' },
                        ]}
                        className="flex-1 pb-2"
                        description="Run custom code over hundreds of repositories and manage the resulting changesets."
                    />

                    <NamespaceSelector
                        namespaces={namespaces}
                        selectedNamespace={selectedNamespace.id}
                        onSelect={setSelectedNamespace}
                    />
                </div>
                <div className="d-flex flex-column flex-0 align-items-center justify-content-center">
                    <button
                        type="button"
                        className="btn btn-primary mb-2"
                        onClick={submitBatchSpec}
                        disabled={isLoading === true}
                    >
                        Run batch spec
                    </button>
                    <BatchSpecDownloadLink name="new-batch-spec" originalInput={code}>
                        or download for src-cli
                    </BatchSpecDownloadLink>
                </div>
            </div>
            <div className="d-flex flex-1">
                <div className={styles.editorContainer}>
                    <MonacoBatchSpecEditor isLightTheme={isLightTheme} value={code} onChange={setCode} />
                </div>
                <Container className={styles.workspacesPreviewContainer}>
                    {codeUpdateError && <ErrorAlert error={codeUpdateError} />}
                    {invalid && specValidator.errors && (
                        <ErrorAlert
                            error={`The entered spec is invalid ${specValidator.errors
                                .map(error => error.message)
                                .join('\n')}`}
                        />
                    )}
                    <PreviewWorkspaces
                        excludeRepo={excludeRepoFromSpec}
                        preview={preview}
                        previewStale={previewStale}
                    />
                </Container>
            </div>
        </div>
    )
}

const NAMESPACE_SELECTOR_ID = 'batch-spec-execution-namespace-selector'

interface NamespaceSelectorProps {
    namespaces: (SettingsUserSubject | SettingsOrgSubject)[]
    selectedNamespace: string
    onSelect: (namespace: SettingsUserSubject | SettingsOrgSubject) => void
}

const NamespaceSelector: React.FunctionComponent<NamespaceSelectorProps> = ({
    namespaces,
    selectedNamespace,
    onSelect,
}) => {
    const onSelectNamespace = useCallback<React.ChangeEventHandler<HTMLSelectElement>>(
        event => {
            const selectedNamespace = namespaces.find(
                (namespace): namespace is SettingsUserSubject | SettingsOrgSubject =>
                    namespace.id === event.target.value
            )
            onSelect(selectedNamespace || namespaces[0])
        },
        [onSelect, namespaces]
    )

    return (
        <div className="form-group d-flex align-items-center">
            <label className="text-nowrap mr-2 mb-0" htmlFor={NAMESPACE_SELECTOR_ID}>
                <strong>Change namespace:</strong>
            </label>
            <select
                className={classNames(styles.namespaceSelector, 'form-control')}
                id={NAMESPACE_SELECTOR_ID}
                value={selectedNamespace}
                onChange={onSelectNamespace}
            >
                {namespaces.map(namespace => (
                    <option key={namespace.id} value={namespace.id}>
                        {getNamespaceDisplayName(namespace)}
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

interface PreviewWorkspacesProps {
    excludeRepo: (repo: string, branch: string) => void
    preview: BatchSpecWithWorkspacesFields | Error | undefined
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
                        {item.cachedResultFound && <ContentSaveIcon className="icon-inline" />}
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
            {preview.importingChangesets && preview.importingChangesets.totalCount > 0 && (
                <>
                    <h3>Importing changesets</h3>
                    <ul>
                        {preview.importingChangesets?.nodes.map(node => (
                            <li key={node.id}>
                                <LinkOrSpan
                                    to={
                                        node.__typename === 'VisibleChangesetSpec' &&
                                        node.description.__typename === 'ExistingChangesetReference'
                                            ? node.description.baseRepository.url
                                            : undefined
                                    }
                                >
                                    {node.__typename === 'VisibleChangesetSpec' &&
                                        node.description.__typename === 'ExistingChangesetReference' &&
                                        node.description.baseRepository.name}
                                </LinkOrSpan>{' '}
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

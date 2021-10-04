import classNames from 'classnames'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useHistory } from 'react-router'

import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import {
    SettingsCascadeProps,
    SettingsOrgSubject,
    SettingsSubject,
    SettingsUserSubject,
} from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ErrorAlert } from '../../../components/alerts'
import { Settings } from '../../../schema/settings.schema'

import { createBatchSpec } from './backend'
import { ExampleTabs } from './examples/ExampleTabs'
import styles from './NewCreateBatchChangeContent.module.scss'

interface CreateBatchChangePageProps extends ThemeProps, SettingsCascadeProps<Settings> {}

export const NewCreateBatchChangeContent: React.FunctionComponent<CreateBatchChangePageProps> = ({
    isLightTheme,
    settingsCascade,
}) => {
    const history = useHistory()

    const [spec, setSpec] = useState<{ fileName: string; code: string }>({ fileName: '', code: '' })
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)
    const [selectedNamespace, setSelectedNamespace] = useState<string>('')
    const [previewID, setPreviewID] = useState<Scalars['ID']>()

    const submitBatchSpec = useCallback<React.MouseEventHandler>(async () => {
        if (!previewID) {
            return
        }
        setIsLoading(true)
        try {
            const execution = await createBatchSpec(previewID)
            history.push(`${execution.namespace.url}/batch-changes/executions/${execution.id}`)
        } catch (error) {
            setIsLoading(error)
        }
    }, [previewID, history])

    return (
        <>
            <h3>1. Write a batch spec</h3>
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
            <ExampleTabs isLightTheme={isLightTheme} updateSpec={setSpec} setPreviewID={setPreviewID} />
            <h3 className="mt-4">
                2. Execute the batch spec <span className="badge badge-info text-uppercase ml-1">Experimental</span>
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
            <NamespaceSelector
                settingsCascade={settingsCascade}
                selectedNamespace={selectedNamespace}
                onSelect={setSelectedNamespace}
            />
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
    const namespaces: SettingsSubject[] = useMemo(
        () =>
            (settingsCascade !== null &&
                !isErrorLike(settingsCascade) &&
                settingsCascade.subjects !== null &&
                settingsCascade.subjects.map(({ subject }) => subject).filter(subject => !isErrorLike(subject))) ||
            [],
        [settingsCascade]
    )

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

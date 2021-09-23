import classNames from 'classnames'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useHistory } from 'react-router'

import { CodeSnippet } from '@sourcegraph/branded/src/components/CodeSnippet'
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

    const submitBatchSpec = useCallback<React.MouseEventHandler>(async () => {
        setIsLoading(true)
        try {
            const execution = await createBatchSpec(spec.code)
            history.push(`${execution.namespace.url}/batch-changes/executions/${execution.id}`)
        } catch (error) {
            setIsLoading(error)
        }
    }, [spec.code, history])

    return (
        <>
            <h2>1. Write a batch spec</h2>
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
            <ExampleTabs isLightTheme={isLightTheme} updateSpec={setSpec} />
            <h2 className="mt-4">2. Execute the batch spec</h2>
            <p>
                Execute the batch spec to preview your batch change before publishing the results. There are two ways to
                execute your batch spec:
            </p>
            <div className="d-flex">
                <div className="w-50 pr-3">
                    <h3>
                        Locally (with{' '}
                        <a href="https://github.com/sourcegraph/src-cli" rel="noopener noreferrer" target="_blank">
                            <span className="text-monospace">src-cli</span>
                        </a>
                        )
                    </h3>
                    <CodeSnippet code={`src batch preview -f ${spec.fileName}`} language="bash" className="mb-3" />
                    <p>Follow the URL printed in your terminal to preview your batch change.</p>
                    <hr className="mb-3" />
                    <p>
                        This is the <strong>classic, existing</strong> way to execute a batch spec. Choose this option
                        if:
                    </p>
                    <ul>
                        <li>You prefer to execute the spec on your personal system</li>
                        <li>You want better debugging capabilities</li>
                        <li>
                            You enjoy reminiscing about the "good old days" of Visual Basic, dial-up internet, and
                            burning CDs, you still use Yahoo as your search engine of choice, or you're not yet sold on
                            this "newfangled MySpace thing" all the kids are using these days
                        </li>
                        <li>
                            You strongly relate to{' '}
                            <a href="https://i.imgur.com/91sn32Q.jpeg" rel="noopener noreferrer" target="_blank">
                                this image
                            </a>
                        </li>
                    </ul>
                </div>
                <div className="w-50 pl-3">
                    <h3 className="d-flex align-items-center">
                        On Sourcegraph <span className="badge badge-info text-uppercase ml-1">Experimental</span>
                    </h3>
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
                    <hr className="mb-3" />
                    <p>
                        This is the <strong>recommended, new</strong> way to execute a batch spec. Choose this option
                        if:
                    </p>
                    <ul>
                        <li>You're trying out batch changes for the first time</li>
                        <li>
                            You don't have <span className="text-monospace">src-cli</span> set up
                        </li>
                        <li>Your batch spec takes a lot of time or resources to execute</li>
                        <li>You swoon anytime you hear the phrase "cloud infrastructure" ☁️☁️☁️</li>
                        <li>You get nervous when you hear the fans of your computer turn on</li>
                    </ul>
                </div>
            </div>
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

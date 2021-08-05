import classNames from 'classnames'
import React, { useCallback, useMemo, useState } from 'react'

import { CodeSnippet } from '@sourcegraph/branded/src/components/CodeSnippet'
import { Link } from '@sourcegraph/shared/src/components/Link'
import {
    SettingsCascadeProps,
    SettingsOrgSubject,
    SettingsSubject,
    SettingsUserSubject,
} from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { Container } from '@sourcegraph/wildcard'

import { ErrorAlert } from '../../../components/alerts'
import { BatchSpecExecutionCreateFields } from '../../../graphql-operations'
import { Settings } from '../../../schema/settings.schema'

import { createBatchSpecExecution } from './backend'
import { ExampleTabs } from './examples/ExampleTabs'
import styles from './NewCreateBatchChangeContent.module.scss'

interface CreateBatchChangePageProps extends ThemeProps, SettingsCascadeProps<Settings> {}

export const NewCreateBatchChangeContent: React.FunctionComponent<CreateBatchChangePageProps> = ({
    isLightTheme,
    settingsCascade,
}) => {
    const [spec, setSpec] = useState<{ fileName: string; code: string }>({ fileName: '', code: '' })
    const [batchSpecExecution, setBatchSpecExecution] = useState<BatchSpecExecutionCreateFields>()
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)
    const [selectedNamespace, setSelectedNamespace] = useState<string>('')

    const submitBatchSpec = useCallback<React.MouseEventHandler>(async () => {
        setBatchSpecExecution(undefined)
        setIsLoading(true)
        try {
            const exec = await createBatchSpecExecution(spec.code, selectedNamespace)
            setBatchSpecExecution(exec)
            setIsLoading(false)
        } catch (error) {
            setIsLoading(error)
        }
    }, [spec.code, selectedNamespace])

    return (
        <>
            <h2>1. Write a batch spec</h2>
            <Container className="mb-3">
                <p className="mb-0">
                    The batch spec YAML file (
                    <a
                        href="https://docs.sourcegraph.com/batch_changes/references/batch_spec_yaml_reference"
                        rel="noopener noreferrer"
                        target="_blank"
                    >
                        syntax reference
                    </a>
                    ) describes what the batch change does. You'll need it to create or update batch changes. We
                    recommend downloading and committing it to source control.
                </p>
            </Container>
            <ExampleTabs isLightTheme={isLightTheme} updateSpec={setSpec} />
            <h2 className="mt-4">2. Execute the spec</h2>
            <Container className="mb-3">
                <p className="mb-0">Execute the spec to preview your batch change, before publishing the results.</p>
            </Container>
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
                        This is the <strong>classic</strong> way to execute a batch spec. Choose this option if:
                    </p>
                    <ul>
                        <li>You prefer to execute the spec on your personal system</li>
                        <li>
                            You enjoy reminiscing about the "good old days" of Visual Basic, dial-up internet, and
                            burning CDs, you still use Yahoo as your search engine of choice, or you're still not sold
                            on this "newfangled MySpace thing" all the kids are using these days
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
                    {batchSpecExecution && (
                        <div className="mt-3 mb-0 alert alert-success">
                            Running batch spec.{' '}
                            <Link
                                to={`${batchSpecExecution.namespace.url}/batch-changes/executions/${batchSpecExecution.id}`}
                            >
                                Check it out here.
                            </Link>
                        </div>
                    )}
                    <hr className="mb-3" />
                    <p>
                        This is the <strong>recommended</strong> way to execute a batch spec. Choose this option if:
                    </p>
                    <ul>
                        <li>You're trying out batch changes for the first time</li>
                        <li>
                            You don't have <span className="text-monospace">src-cli</span> set up
                        </li>
                        <li>Your batch spec takes a lot of time or resources to execute</li>
                        <li>You swoon anytime you hear the phrase "cloud infrastructure" ☁️☁️☁️</li>
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
                value={selectedNamespace || userNamespace.id}
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

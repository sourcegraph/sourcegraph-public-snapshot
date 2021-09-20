import { Tab, TabList, TabPanel, TabPanels, Tabs, useTabsContext } from '@reach/tabs'
import classNames from 'classnames'
import React, { useState, useCallback, useEffect, useMemo } from 'react'
import { Subject } from 'rxjs'
import { catchError, debounceTime, startWith, switchMap } from 'rxjs/operators'

import { isErrorLike } from '@sourcegraph/codeintellify/lib/errors'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { asError } from '@sourcegraph/shared/src/util/errors'
import { pluralize } from '@sourcegraph/shared/src/util/strings'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Container, LoadingSpinner } from '@sourcegraph/wildcard'

import batchSpecSchemaJSON from '../../../../../../../schema/batch_spec.schema.json'
import { ErrorAlert } from '../../../../components/alerts'
import { SidebarGroup, SidebarGroupHeader } from '../../../../components/Sidebar'
import { BatchSpecWorkspacesFields } from '../../../../graphql-operations'
import { MonacoSettingsEditor } from '../../../../settings/MonacoSettingsEditor'
import { BatchSpecDownloadLink, getFileName } from '../../BatchSpec'

import { resolveWorkspacesForBatchSpec } from './backend'
import combySample from './comby.batch.yaml'
import helloWorldSample from './empty.batch.yaml'
import styles from './ExampleTabs.module.scss'
import goImportsSample from './go-imports.batch.yaml'
import minimalSample from './minimal.batch.yaml'

interface Example {
    name: string
    code: string
}

export interface Spec {
    fileName: string
    code: string
}

const EXAMPLES: [Example, Example, Example, Example] = [
    { name: 'Hello world', code: helloWorldSample },
    { name: 'Modify with comby', code: combySample },
    { name: 'Update go imports', code: goImportsSample },
    { name: 'Minimal', code: minimalSample },
]

interface ExampleTabsProps extends ThemeProps {
    updateSpec: (spec: Spec) => void
}

export const ExampleTabs: React.FunctionComponent<ExampleTabsProps> = ({ isLightTheme, updateSpec }) => (
    <Tabs className={styles.exampleTabs}>
        <TabList className="d-flex flex-column flex-shrink-0">
            <SidebarGroup>
                <SidebarGroupHeader label="Examples" />
                {EXAMPLES.map((example, index) => (
                    <ExampleTab key={example.name} index={index}>
                        {example.name}
                    </ExampleTab>
                ))}
            </SidebarGroup>
        </TabList>

        <div className="ml-3 flex-grow-1">
            <TabPanels>
                {EXAMPLES.map((example, index) => (
                    <ExampleTabPanel
                        key={example.name}
                        example={example}
                        isLightTheme={isLightTheme}
                        index={index}
                        updateSpec={updateSpec}
                    />
                ))}
            </TabPanels>
        </div>
    </Tabs>
)

const ExampleTab: React.FunctionComponent<{ index: number }> = ({ children, index }) => {
    const { selectedIndex } = useTabsContext()

    return (
        <Tab>
            <button
                type="button"
                className={classNames(
                    'btn text-left sidebar__link--inactive d-flex w-100',
                    index === selectedIndex && 'btn-primary'
                )}
            >
                {children}
            </button>
        </Tab>
    )
}

interface ExampleTabPanelProps extends ThemeProps {
    example: Example
    updateSpec: (spec: Spec) => void
    index: number
}

const ExampleTabPanel: React.FunctionComponent<ExampleTabPanelProps> = ({
    example,
    isLightTheme,
    index,
    updateSpec,
    ...props
}) => {
    const [code, setCode] = useState<string>(example.code)

    const codeUpdates = useMemo(() => new Subject<string>(), [])

    useEffect(() => {
        codeUpdates.next(code)
    }, [codeUpdates, code])

    const preview = useObservable(
        useMemo(
            () =>
                codeUpdates.pipe(
                    startWith(code),
                    debounceTime(5000),
                    switchMap(code => resolveWorkspacesForBatchSpec(code)),
                    catchError(error => [asError(error)])
                ),
            // Don't want to trigger on changes to code, it's just the initial value.
            // eslint-disable-next-line react-hooks/exhaustive-deps
            [codeUpdates]
        )
    )

    const { selectedIndex } = useTabsContext()

    // Update the spec in parent state whenever the code changes
    useEffect(() => {
        if (selectedIndex === index) {
            updateSpec({ code, fileName: getFileName(example.name) })
        }
    }, [code, example.name, updateSpec, selectedIndex, index])

    const reset = useCallback(() => setCode(example.code), [example.code])

    return (
        <TabPanel {...props}>
            <div className="d-flex justify-content-end align-items-center mb-2">
                {/* TODO: Confirmation before discarding changes */}
                <button className="text-right btn btn-outline-secondary text-nowrap mr-2" type="button" onClick={reset}>
                    Reset
                </button>
                <BatchSpecDownloadLink name={example.name} originalInput={code} />
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
                <h3>Preview workspaces</h3>
                <PreviewWorkspaces preview={preview} />
            </Container>
        </TabPanel>
    )
}

const PreviewWorkspaces: React.FunctionComponent<{ preview: BatchSpecWorkspacesFields | Error | undefined }> = ({
    preview,
}) => {
    if (isErrorLike(preview)) {
        return <ErrorAlert error={preview} />
    }
    if (!preview) {
        return <LoadingSpinner />
    }
    return (
        <>
            <p className="text-monospace">
                allowUnsupported: {JSON.stringify(preview.allowUnsupported)}
                <br />
                allowIgnored: {JSON.stringify(preview.allowIgnored)}
            </p>
            <ul className="list-group p-1 mb-0">
                {preview.workspaces.map(item => (
                    <li
                        className="list-group-item"
                        key={`${item.repository.id}_${item.branch.target.oid}_${item.path || '/'}`}
                    >
                        <p>
                            {item.repository.name}:{item.branch.abbrevName}@{item.branch.target.oid} Path:{' '}
                            {item.path || '/'}
                        </p>
                        <p>{item.searchResultPaths.join(', ')}</p>
                        <ul>
                            {item.steps.map((step, index) => (
                                <li key={index}>
                                    <span className="text-monospace">{step.command}</span>
                                    <br />
                                    <span className="text-muted">{step.container}</span>
                                </li>
                            ))}
                        </ul>
                    </li>
                ))}
            </ul>
            {preview.workspaces.length === 0 && <span className="text-muted">No workspaces found</span>}
            <hr />
            {preview.ignored.length > 0 && (
                <>
                    <p>
                        {preview.ignored.length} {pluralize('repo is', preview.ignored.length, 'repos are')} ignored
                        {preview.allowIgnored && (
                            <>
                                , but {pluralize('it has', preview.ignored.length, 'they have')} been included, based on
                                settings
                            </>
                        )}
                        .
                    </p>
                    <ul>
                        {preview.ignored.map(repo => (
                            <li key={repo.id}>{repo.name}</li>
                        ))}
                    </ul>
                </>
            )}
            {preview.unsupported.length > 0 && (
                <>
                    <p>
                        {preview.unsupported.length} {pluralize('repo is', preview.unsupported.length, 'repos are')}{' '}
                        unsupported
                        {preview.allowUnsupported && (
                            <>
                                , but {pluralize('it has', preview.unsupported.length, 'they have')} been included,
                                based on settings
                            </>
                        )}
                        .
                    </p>
                    <ul>
                        {preview.unsupported.map(repo => (
                            <li key={repo.id}>{repo.name}</li>
                        ))}
                    </ul>
                </>
            )}
        </>
    )
}

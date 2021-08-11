import { Tab, TabList, TabPanel, TabPanels, Tabs, useTabsContext } from '@reach/tabs'
import classNames from 'classnames'
import React, { useState, useCallback, useEffect } from 'react'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Container } from '@sourcegraph/wildcard'

import batchSpecSchemaJSON from '../../../../../../../schema/batch_spec.schema.json'
import { SidebarGroup, SidebarGroupHeader, SidebarGroupItems } from '../../../../components/Sidebar'
import { MonacoSettingsEditor } from '../../../../settings/MonacoSettingsEditor'
import { BatchSpecDownloadLink, getFileName } from '../../BatchSpec'

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
                <SidebarGroupItems>
                    <SidebarGroupHeader label="Examples" />
                    {EXAMPLES.map((example, index) => (
                        <ExampleTab key={example.name} index={index}>
                            {example.name}
                        </ExampleTab>
                    ))}
                </SidebarGroupItems>
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
                    'btn text-left sidebar__link--inactive d-flex sidebar-nav-link w-100',
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
            <Container>
                <MonacoSettingsEditor
                    isLightTheme={isLightTheme}
                    language="yaml"
                    value={code}
                    jsonSchema={batchSpecSchemaJSON}
                    onChange={setCode}
                />
            </Container>
        </TabPanel>
    )
}

import { Tab, TabList, TabPanel, TabPanels, Tabs, useTabsContext } from '@reach/tabs'
import classNames from 'classnames'
import React, { useState, useCallback, useEffect } from 'react'

import { ThemeProps } from '@sourcegraph/shared/src/theme'

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

export const ExampleTabs: React.FunctionComponent<ExampleTabsProps> = ({ isLightTheme, updateSpec }) => {
    const [activeIndex, setActiveIndex] = useState<number>(0)

    // Update the spec whenever the active changes, before the user makes any edits to it
    useEffect(() => {
        updateSpec({ code: EXAMPLES[activeIndex].code, fileName: getFileName(EXAMPLES[activeIndex].name) })
    }, [updateSpec, activeIndex])

    return (
        <Tabs className={styles.exampleTabs} onChange={setActiveIndex}>
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
                    {EXAMPLES.map(example => (
                        <ExampleTabPanel
                            key={example.name}
                            example={example}
                            isLightTheme={isLightTheme}
                            updateSpec={updateSpec}
                        />
                    ))}
                </TabPanels>
            </div>
        </Tabs>
    )
}

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
}

const ExampleTabPanel: React.FunctionComponent<ExampleTabPanelProps> = ({
    example,
    isLightTheme,
    updateSpec,
    ...props
}) => {
    const [code, setCode] = useState<string>(example.code)

    const onChange = useCallback(
        (newCode: string) => {
            setCode(newCode)
            updateSpec({ code: newCode, fileName: getFileName(example.name) })
        },
        [updateSpec, example.name]
    )
    const reset = useCallback(() => setCode(example.code), [example.code])

    return (
        <TabPanel {...props}>
            <div className="d-flex justify-content-between align-items-center mb-2">
                Choose an example template, then edit your spec here.
                <div className="d-flex">
                    {/* TODO: Confirmation before discarding changes */}
                    <button
                        className="text-right btn btn-outline-secondary text-nowrap mr-2"
                        type="button"
                        onClick={reset}
                    >
                        Reset
                    </button>
                    <BatchSpecDownloadLink name={example.name} originalInput={code} />
                </div>
            </div>
            <MonacoSettingsEditor
                isLightTheme={isLightTheme}
                language="yaml"
                value={code}
                jsonSchema={batchSpecSchemaJSON}
                onChange={onChange}
            />
        </TabPanel>
    )
}

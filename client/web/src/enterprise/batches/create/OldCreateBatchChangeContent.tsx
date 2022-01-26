import classNames from 'classnames'
import React, { useCallback, useState } from 'react'

import { CodeSnippet } from '@sourcegraph/branded/src/components/CodeSnippet'
import { Container, Button } from '@sourcegraph/wildcard'

import { SidebarGroup, SidebarGroupHeader } from '../../../components/Sidebar'

import combySample from './library/comby.batch.yaml'
import goImportsSample from './library/go-imports.batch.yaml'
import helloWorldSample from './library/hello-world.batch.yaml'
import minimalSample from './library/minimal.batch.yaml'

// SampleTabHeader is superseded by ExampleTabs and can be removed when SSBC is rolled out
// at the same time as this exported component from this file is removed
interface SampleTabHeaderProps {
    sample: Sample
    active: boolean
    setSelectedSample: (sample: Sample) => void
}

const SampleTabHeader: React.FunctionComponent<SampleTabHeaderProps> = ({ sample, active, setSelectedSample }) => {
    const onClick = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setSelectedSample(sample)
        },
        [setSelectedSample, sample]
    )
    return (
        <Button
            onClick={onClick}
            className={classNames('text-left sidebar__link--inactive d-flex w-100', active && 'btn-primary')}
        >
            {sample.name}
        </Button>
    )
}

interface Sample {
    name: string
    file: string
}

const samples: Sample[] = [
    { name: 'Hello world', file: helloWorldSample },
    { name: 'Modify with comby', file: combySample },
    { name: 'Update go imports', file: goImportsSample },
    { name: 'Minimal', file: minimalSample },
]

export const OldBatchChangePageContent: React.FunctionComponent<{}> = () => {
    const [selectedSample, setSelectedSample] = useState<Sample>(samples[0])

    return (
        <>
            <h2>1. Write a batch spec YAML file</h2>
            <Container className="mb-3">
                <p className="mb-0">
                    The batch spec (
                    <a
                        href="https://docs.sourcegraph.com/batch_changes/references/batch_spec_yaml_reference"
                        rel="noopener noreferrer"
                        target="_blank"
                    >
                        syntax reference
                    </a>
                    ) describes what the batch change does. You'll provide it when previewing, creating, and updating
                    batch changes. We recommend committing it to source control.
                </p>
            </Container>
            <div className="d-flex mb-3">
                <div className="flex-shrink-0">
                    <SidebarGroup>
                        <SidebarGroupHeader label="Examples" />
                        {samples.map(sample => (
                            <SampleTabHeader
                                key={sample.name}
                                sample={sample}
                                active={selectedSample.name === sample.name}
                                setSelectedSample={setSelectedSample}
                            />
                        ))}
                    </SidebarGroup>
                </div>
                <Container className="ml-3 flex-grow-1">
                    <CodeSnippet code={selectedSample.file} language="yaml" className="mb-0" />
                </Container>
            </div>
            <h2>2. Preview the batch change with Sourcegraph CLI</h2>
            <Container className="mb-3">
                <p>
                    Use the{' '}
                    <a href="https://github.com/sourcegraph/src-cli" rel="noopener noreferrer" target="_blank">
                        Sourcegraph CLI (src)
                    </a>{' '}
                    to preview the commits and changesets that your batch change will make:
                </p>
                <CodeSnippet code={`src batch preview -f ${selectedSample.name}`} language="bash" className="mb-3" />
                <p className="mb-0">
                    Follow the URL printed in your terminal to see the preview and (when you're ready) create the batch
                    change.
                </p>
            </Container>
        </>
    )
}

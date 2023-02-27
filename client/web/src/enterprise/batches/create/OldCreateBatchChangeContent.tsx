import React, { useCallback, useState } from 'react'

import { CodeSnippet } from '@sourcegraph/branded/src/components/CodeSnippet'
import { Container, Button, Link, H2, Text } from '@sourcegraph/wildcard'

import { SidebarGroup, SidebarGroupHeader } from '../../../components/Sidebar'
import combySample from '../batch-spec/edit/library/comby.batch.yaml'
import goImportsSample from '../batch-spec/edit/library/go-imports.batch.yaml'
import helloWorldSample from '../batch-spec/edit/library/hello-world.batch.yaml'
import minimalSample from '../batch-spec/edit/library/minimal.batch.yaml'
import { getFileName } from '../BatchSpec'

// SampleTabHeader is superseded by ExampleTabs and can be removed when SSBC is rolled out
// at the same time as this exported component from this file is removed
interface SampleTabHeaderProps {
    sample: Sample
    active: boolean
    setSelectedSample: (sample: Sample) => void
}

const SampleTabHeader: React.FunctionComponent<React.PropsWithChildren<SampleTabHeaderProps>> = ({
    sample,
    active,
    setSelectedSample,
}) => {
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
            className="text-left sidebar__link--inactive d-flex w-100"
            variant={active ? 'primary' : undefined}
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

export const OldBatchChangePageContent: React.FunctionComponent<React.PropsWithChildren<{}>> = () => {
    const [selectedSample, setSelectedSample] = useState<Sample>(samples[0])

    return (
        <>
            <H2 data-testid="batch-spec-yaml-file">1. Write a batch spec YAML file</H2>
            <Container className="mb-3">
                <Text className="mb-0">
                    The batch spec (
                    <Link
                        to="/help/batch_changes/references/batch_spec_yaml_reference"
                        rel="noopener noreferrer"
                        target="_blank"
                    >
                        syntax reference
                    </Link>
                    ) describes what the batch change does. You'll provide it when previewing, creating, and updating
                    batch changes. We recommend committing it to source control.
                </Text>
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
                <Container className="ml-3 flex-grow-1 overflow-auto">
                    <CodeSnippet code={selectedSample.file} language="yaml" className="mb-0" />
                </Container>
            </div>
            <H2>2. Preview the batch change with Sourcegraph CLI</H2>
            <Container className="mb-3">
                <Text>
                    Use the{' '}
                    <Link to="https://github.com/sourcegraph/src-cli" rel="noopener noreferrer" target="_blank">
                        Sourcegraph CLI (src)
                    </Link>{' '}
                    to preview the commits and changesets that your batch change will make:
                </Text>
                <CodeSnippet
                    code={`src batch preview -f ${getFileName(selectedSample.name)}`}
                    language="bash"
                    className="mb-3"
                />
                <Text className="mb-0">
                    Follow the URL printed in your terminal to see the preview and (when you're ready) create the batch
                    change.
                </Text>
            </Container>
        </>
    )
}

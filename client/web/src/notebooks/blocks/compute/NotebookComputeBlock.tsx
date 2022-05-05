import React from 'react'

import ElmComponent from 'react-elm-components'

import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { BlockInput, BlockProps, ComputeBlock } from '../..'
import { useCommonBlockMenuActions } from '../menu/useCommonBlockMenuActions'
import { NotebookBlock } from '../NotebookBlock'

import { Elm } from './component/src/Main.elm'

import styles from './NotebookComputeBlock.module.scss'

interface ComputeBlockProps extends BlockProps<ComputeBlock>, ThemeProps {
    platformContext: Pick<PlatformContext, 'sourcegraphURL'>
}

interface ElmEvent {
    data: string
    eventType?: string
    id?: string
}

interface ExperimentalOptions {}

interface ComputeInput {
    computeQueries: string[]
    experimentalOptions: ExperimentalOptions
}

interface Ports {
    receiveEvent: { send: (event: ElmEvent) => void }
    openStream: { subscribe: (callback: (args: string[]) => void) => void }
    emitInput: { subscribe: (callback: (input: ComputeInput) => void) => void }
}

const updateBlockInput = (id: string, onBlockInputChange: (id: string, blockInput: BlockInput) => void) => (
    blockInput: BlockInput
): void => {
    onBlockInputChange(id, blockInput)
}

const setupPorts = (updateBlockInputWithID: (blockInput: BlockInput) => void) => (ports: Ports): void => {
    const sources: { [key: string]: EventSource } = {}

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    function sendEventToElm(event: any): void {
        // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-member-access
        const elmEvent = { data: event.data, eventType: event.type || null, id: event.id || null }
        ports.receiveEvent.send(elmEvent)
    }

    function newEventSource(address: string): EventSource {
        sources[address] = new EventSource(address)
        return sources[address]
    }

    function deleteAllEventSources(): void {
        for (const [key] of Object.entries(sources)) {
            deleteEventSource(key)
        }
    }

    function deleteEventSource(address: string): void {
        sources[address].close()
        delete sources[address]
    }

    ports.openStream.subscribe((args: string[]) => {
        deleteAllEventSources() // Close any open streams if we receive a request to open a new stream before seeing 'done'.
        console.log(`stream: ${args[0]}`)
        const address = args[0]

        const eventSource = newEventSource(address)
        eventSource.addEventListener('error', () => {
            console.log('EventSource failed')
        })
        eventSource.addEventListener('results', sendEventToElm)
        eventSource.addEventListener('alert', sendEventToElm)
        eventSource.addEventListener('error', sendEventToElm)
        eventSource.addEventListener('done', () => {
            deleteEventSource(address)
            // Note: 'done:true' is sent in progress too. But we want a 'done' for the entire stream in case we don't see it.
            sendEventToElm({ type: 'done', data: '' })
        })
    })

    ports.emitInput.subscribe((computeInput: ComputeInput) => {
        updateBlockInputWithID({ type: 'compute', input: JSON.stringify(computeInput) })
    })
}

export const NotebookComputeBlock: React.FunctionComponent<React.PropsWithChildren<ComputeBlockProps>> = React.memo(
    ({
        id,
        input,
        output,
        isSelected,
        isLightTheme,
        platformContext,
        isReadOnly,
        onBlockInputChange,
        onRunBlock,
        ...props
    }) => {
        const commonMenuActions = useCommonBlockMenuActions({ id, isReadOnly, ...props })
        return (
            <NotebookBlock
                className={styles.input}
                id={id}
                aria-label="Notebook compute block"
                isSelected={isSelected}
                isReadOnly={isReadOnly}
                actions={isSelected ? commonMenuActions : []}
                {...props}
            >
                <div className="elm">
                    <ElmComponent
                        src={Elm.Main}
                        ports={setupPorts(updateBlockInput(id, onBlockInputChange))}
                        flags={{
                            sourcegraphURL: platformContext.sourcegraphURL,
                            isLightTheme,
                            computeInput: input === '' ? null : (JSON.parse(input) as ComputeInput),
                        }}
                    />
                </div>
            </NotebookBlock>
        )
    }
)

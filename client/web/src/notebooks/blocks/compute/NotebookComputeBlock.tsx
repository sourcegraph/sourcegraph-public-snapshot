import React from 'react'

import { fetchEventSource } from '@microsoft/fetch-event-source'
import ElmComponent from 'react-elm-components'

import { logger } from '@sourcegraph/common'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { BlockInput, BlockProps, ComputeBlock } from '../..'
import { Elm } from '../../../search/results/components/compute/src/Main.elm'
import { useCommonBlockMenuActions } from '../menu/useCommonBlockMenuActions'
import { NotebookBlock } from '../NotebookBlock'

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

const setupPorts = (sourcegraphURL: string, updateBlockInputWithID: (blockInput: BlockInput) => void) => (
    ports: Ports
): void => {
    const openRequests: AbortController[] = []

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    function sendEventToElm(event: any): void {
        // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-member-access
        const elmEvent = { data: event.data, eventType: event.event || null, id: event.id || null }
        ports.receiveEvent.send(elmEvent)
    }

    ports.openStream.subscribe((args: string[]) => {
        logger.log(`stream: ${args[0]}`)
        const address = args[0]

        // Close any open streams if we receive a request to open a new stream before seeing 'done'.
        for (const request of openRequests) {
            request.abort()
        }

        const ctrl = new AbortController()
        openRequests.push(ctrl)
        async function fetch(): Promise<void> {
            await fetchEventSource(address, {
                method: 'GET',
                headers: {
                    'X-Requested-With': 'Sourcegraph',
                },
                signal: ctrl.signal,
                onerror(error) {
                    logger.error(`Compute EventSource error: ${JSON.stringify(error)}`)
                },
                onclose() {
                    // Note: 'done:true' is sent in progress too. But we want a 'done' for the entire stream in case we don't see it.
                    sendEventToElm({ type: 'done', data: '' })
                    openRequests.splice(0)
                },
                onmessage(event) {
                    sendEventToElm(event)
                },
            })
        }

        fetch().catch(error => {
            logger.error(`Compute fetch error: ${JSON.stringify(error)}`)
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
                        ports={setupPorts(platformContext.sourcegraphURL, updateBlockInput(id, onBlockInputChange))}
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

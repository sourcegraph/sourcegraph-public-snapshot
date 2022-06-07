import React from 'react'

import ElmComponent from 'react-elm-components'
import { Subscription } from 'rxjs'

import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { streamComputeQuery } from '@sourcegraph/shared/src/search/stream'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { BlockInput, BlockProps, ComputeBlock } from '../..'
import { Elm } from '../../../search/results/components/compute/src/Main.elm'
import { useCommonBlockMenuActions } from '../menu/useCommonBlockMenuActions'
import { NotebookBlock } from '../NotebookBlock'

import styles from './NotebookComputeBlock.module.scss'

interface ComputeBlockProps extends BlockProps<ComputeBlock>, ThemeProps {
    platformContext: Pick<PlatformContext, 'sourcegraphURL'>
}

const updateBlockInput = (id: string, onBlockInputChange: (id: string, blockInput: BlockInput) => void) => (
    blockInput: BlockInput
): void => {
    onBlockInputChange(id, blockInput)
}

const setupPorts = (updateBlockInputWithID: (blockInput: BlockInput) => void) => (ports: Ports): void => {
    const openRequests: AbortController[] = []
    let eventSubscription: Subscription | null = null

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    function sendEventToElm(event: any): void {
        // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-member-access
        const elmEvent = { data: event.data, eventType: event.event || null, id: event.id || null }
        ports.receiveEvent.send(elmEvent)
    }

    ports.openStream.subscribe((query: string) => {
        console.log(`stream: ${query}}`)

        // Close any open streams if we receive a request to open a new stream before seeing 'done'.
        for (const request of openRequests) {
            request.abort()
        }
        eventSubscription?.unsubscribe()

        const ctrl = new AbortController()
        openRequests.push(ctrl)

        eventSubscription = streamComputeQuery(query, ctrl.signal).subscribe(
            event => sendEventToElm(event),
            error => console.log(`Compute EventSource error: ${JSON.stringify(error)}`),
            () => {
                // Note: 'done:true' is sent in progress too. But we want a 'done' for the entire stream in case we don't see it.
                sendEventToElm({ type: 'done', data: '' })
                openRequests.splice(0)
            }
        )
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
                            isLightTheme,
                            computeInput: input === '' ? null : (JSON.parse(input) as ComputeInput),
                        }}
                    />
                </div>
            </NotebookBlock>
        )
    }
)

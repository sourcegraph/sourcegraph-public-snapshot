import classNames from 'classnames'
import React, { useRef } from 'react'
import ElmComponent from 'react-elm-components'

import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { BlockProps, ComputeBlock } from '../..'
import { NotebookBlockMenu } from '../menu/NotebookBlockMenu'
import { useCommonBlockMenuActions } from '../menu/useCommonBlockMenuActions'
import blockStyles from '../NotebookBlock.module.scss'
import { useBlockSelection } from '../useBlockSelection'
import { useBlockShortcuts } from '../useBlockShortcuts'

import { Elm } from './component/src/Main.elm'
import styles from './NotebookComputeBlock.module.scss'

interface ComputeBlockProps extends BlockProps, ComputeBlock, ThemeProps {
    isMacPlatform: boolean
}

interface ElmEvent {
    data: string
    eventType?: string
    id?: string
}

function setupPorts(ports: {
    receiveEvent: { send: (event: ElmEvent) => void }
    openStream: { subscribe: (callback: (args: string[]) => void) => void }
}): void {
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
}

export const NotebookComputeBlock: React.FunctionComponent<ComputeBlockProps> = ({
    id,
    input,
    output,
    isSelected,
    isLightTheme,
    isMacPlatform,
    isReadOnly,
    onRunBlock,
    onSelectBlock,
    ...props
}) => {
    const isInputFocused = false
    const blockElement = useRef<HTMLDivElement>(null)

    const { onSelect } = useBlockSelection({
        id,
        blockElement: blockElement.current,
        isSelected,
        isInputFocused,
        onSelectBlock,
        ...props,
    })

    const { onKeyDown } = useBlockShortcuts({
        id,
        isMacPlatform,
        onEnterBlock: () => {},
        ...props,
        onRunBlock: () => {},
    })

    const modifierKeyLabel = isMacPlatform ? '⌘' : 'Ctrl'
    const commonMenuActions = useCommonBlockMenuActions({
        modifierKeyLabel,
        isInputFocused,
        isReadOnly,
        isMacPlatform,
        ...props,
    })

    const blockMenu = isSelected && !isReadOnly && <NotebookBlockMenu id={id} actions={commonMenuActions} />

    return (
        <div className={classNames('block-wrapper', blockStyles.blockWrapper)} data-block-id={id}>
            {/* See the explanation for the disable above. */}
            {/* eslint-disable-next-line jsx-a11y/no-static-element-interactions */}
            <div
                className={classNames(
                    blockStyles.block,
                    styles.input,
                    (isInputFocused || isSelected) && blockStyles.selected
                )}
                onClick={onSelect}
                onFocus={onSelect}
                onKeyDown={onKeyDown}
                // A tabIndex is necessary to make the block focusable.
                // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex
                tabIndex={0}
                aria-label="Notebook compute block"
                ref={blockElement}
            >
                <div className="elm">
                    {/* eslint-disable-next-line @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-member-access */}
                    <ElmComponent src={Elm.Main} ports={setupPorts} flags={null} />
                </div>
            </div>
            {blockMenu}
        </div>
    )
}

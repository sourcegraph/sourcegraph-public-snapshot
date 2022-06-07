import React from 'react'

import { fetchEventSource } from '@microsoft/fetch-event-source'
import ElmComponent from 'react-elm-components'

import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { useNavbarQueryState } from '../../stores'

import { Elm } from './components/compute/src/Main.elm'

interface ComputeSearchResultsProps extends ThemeProps {
    platformContext: Pick<PlatformContext, 'sourcegraphURL'>
}

const setupPorts = (sourcegraphURL: string) => (ports: Ports): void => {
    const openRequests: AbortController[] = []

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    function sendEventToElm(event: any): void {
        // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-member-access
        const elmEvent = { data: event.data, eventType: event.event || null, id: event.id || null }
        ports.receiveEvent.send(elmEvent)
    }

    ports.openStream.subscribe((args: string[]) => {
        console.log(`stream: ${args[0]}`)
        const address = args[0]

        // Close any open streams if we receive a request to open a new stream before seeing 'done'.
        for (const request of openRequests) {
            request.abort()
        }

        const ctrl = new AbortController()
        openRequests.push(ctrl)
        async function fetch(): Promise<void> {
            await fetchEventSource(address, {
                method: 'POST',
                headers: {
                    origin: sourcegraphURL,
                },
                signal: ctrl.signal,
                onerror(error) {
                    console.log(`Compute EventSource error: ${JSON.stringify(error)}`)
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
            console.log(`Compute fetch error: ${JSON.stringify(error)}`)
        })
    })
}

export const ComputeSearchResults: React.FunctionComponent<ComputeSearchResultsProps> = ({
    platformContext,
    isLightTheme,
}) => {
    const query = useNavbarQueryState(state => state.searchQueryFromURL)

    return (
        <ElmComponent
            src={Elm.Main}
            ports={setupPorts(platformContext.sourcegraphURL)}
            flags={{
                sourcegraphURL: platformContext.sourcegraphURL,
                isLightTheme,
                computeInput: { computeQueries: [query] },
            }}
        />
    )
}

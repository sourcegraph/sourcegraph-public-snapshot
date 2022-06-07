import React from 'react'

import ElmComponent from 'react-elm-components'
import { Subscription } from 'rxjs'

import { streamComputeQuery } from '@sourcegraph/shared/src/search/stream'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { useNavbarQueryState } from '../../stores'

import { Elm } from './components/compute/src/Main.elm'

interface ComputeSearchResultsProps extends ThemeProps {}

const setupPorts = () => (ports: Ports): void => {
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
}

export const ComputeSearchResults: React.FunctionComponent<ComputeSearchResultsProps> = ({ isLightTheme }) => {
    const query = useNavbarQueryState(state => state.searchQueryFromURL)

    return (
        <ElmComponent
            src={Elm.Main}
            ports={setupPorts()}
            flags={{
                isLightTheme,
                computeInput: { computeQueries: [query] },
            }}
        />
    )
}

import H from 'history'
import React from 'react'
import { ChangesAreaContext } from '../global/ChangesArea'
import { ChangesList } from './ChangesList'

interface Props extends ChangesAreaContext {
    history: H.History
    location: H.Location
}

/**
 * The changes list page.
 */
export const ChangesListPage: React.FunctionComponent<Props> = props => {
    const query = new URLSearchParams(location.search).get('q') || ''
    const onQueryChange = (query: string) => {
        const params = new URLSearchParams(location.search)
        params.set('q', query)
        props.history.push({ search: `${params}` })
    }

    return (
        <div className="changes-overview-page mt-3 container">
            <h1 className="h4">
                Changes <span className="text-muted">to...</span>
            </h1>
            <ChangesList {...props} query={query} onQueryChange={onQueryChange} />
        </div>
    )
}

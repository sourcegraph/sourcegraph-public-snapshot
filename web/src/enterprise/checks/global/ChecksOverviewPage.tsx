import H from 'history'
import * as React from 'react'
import { ChecksAreaContext } from './ChecksArea'
import { ChecksList } from './ChecksList'

interface Props extends ChecksAreaContext {
    location: H.Location
}

/**
 * The checks overview page.
 */
export const ChecksOverviewPage: React.FunctionComponent<Props> = ({ location }) => (
    <div className="checks-overview-page mt-3 container">
        <ChecksList location={location} />
    </div>
)

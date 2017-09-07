import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { SearchBox } from 'sourcegraph/search/SearchBox'
import { sourcegraphContext } from 'sourcegraph/util/sourcegraphContext'

/**
 * The landing page
 */
export class Home extends React.Component<RouteComponentProps<void>, {}> {
    public render(): JSX.Element | null {
        return (
            <div className='home'>
                <img className='home__logo' src={`${sourcegraphContext.assetsRoot}/img/ui2/sourcegraph-head-logo.svg`} />
                <div className='home__search-box-container'>
                    <SearchBox {...this.props} />
                </div>
            </div>
        )
    }
}

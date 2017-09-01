import * as React from 'react';
import { RouteComponentProps } from 'react-router';
import { SearchBox } from 'sourcegraph/search/SearchBox';
import { sourcegraphContext } from 'sourcegraph/util/sourcegraphContext';

export class Home extends React.Component<RouteComponentProps<any>, {}> {
    public render(): JSX.Element | null {
        return <div className='home'>
            <img className='header' src={`${sourcegraphContext.assetsRoot}/img/ui2/sourcegraph-head-logo.svg`} />
            <div id='search-widget'>
                <SearchBox history={this.props.history} />
            </div>
            <div className='footer'>
            </div>
        </div>;
    }
}

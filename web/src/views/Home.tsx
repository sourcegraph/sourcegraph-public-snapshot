import * as React from 'react';
import { RouteComponentProps } from 'react-router';
import { SearchBox } from 'sourcegraph/search/SearchBox';

export class Home extends React.Component<RouteComponentProps<any>, {}> {
    constructor(props: any) {
        super(props);
    }

    public render(): JSX.Element | null {
        return <SearchBox />;
    }
}

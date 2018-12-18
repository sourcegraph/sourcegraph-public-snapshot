import * as React from 'react'

// tslint:disable-next-line:ban-imports
import { MemoryRouter, Route, RouteComponentProps } from 'react-router'

export interface WithRouterProviderProps<T> {
    children(routerProps: RouteComponentProps<T>): React.ReactNode
}

export class WithRouterProvider<T = undefined> extends React.Component<WithRouterProviderProps<T>, {}> {
    public render(): React.ReactNode {
        return (
            <MemoryRouter>
                <Route path="/" render={this.renderComposedComponent} />
            </MemoryRouter>
        )
    }

    private renderComposedComponent = (routerProps: RouteComponentProps<T>): React.ReactNode =>
        this.props.children(routerProps)
}

import * as React from 'react'
import './App.scss'
import { CriticalConfigEditor } from './CriticalConfigEditor'
import { ProductSubscriptionStatus } from './ProductSubscriptionStatus'

export class App extends React.Component<{}, {}> {
    public render(): JSX.Element | null {
        return (
            <div>
                <h1 className="app__title">Sourcegraph management console</h1>
                <p className="app__subtitle">
                    View and edit critical Sourcegraph configuration. See{' '}
                    <a target="_blank" href="https://docs.sourcegraph.com/admin/management_console">
                        documentation
                    </a>{' '}
                    for more information.
                </p>
                <ProductSubscriptionStatus className="product-certificate-parcel" />
                <CriticalConfigEditor />
            </div>
        )
    }
}

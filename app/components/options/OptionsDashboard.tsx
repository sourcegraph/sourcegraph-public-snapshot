import * as React from 'react'
import { Redirect, Route, Switch } from 'react-router'
import { HashRouter } from 'react-router-dom'
import { sourcegraphUrl } from '../../util/context'
import { ExtensionRegistry } from './ExtensionRegistry'
import { OptionsConfiguration } from './OptionsConfiguration'
import { OptionsPageSidebar } from './OptionsPageSidebar'

const EXTENSION_ROUTE = '/extensions/:url/:author/:extension?'

export class OptionsDashboard extends React.Component<any, {}> {
    /**
     * extensionRedirectComponent redirects to the extension page on Sourcegraph instead of showing
     * the extension in the options page and keeps the user on the extension registry page.
     */
    private extensionRedirectComponent = (): JSX.Element | null => {
        const extensionUrl = window.location.hash.replace('#', sourcegraphUrl)
        window.open(extensionUrl, '_blank')
        return <Redirect from={EXTENSION_ROUTE} to="/extensions" />
    }

    public render(): JSX.Element {
        return (
            <HashRouter>
                <div className="site-admin-area area">
                    <OptionsPageSidebar className="area__sidebar" />
                    <div className="area__content">
                        <Switch>
                            <Route path="/" component={OptionsConfiguration} exact={true} />
                            <Route path="/extensions" component={ExtensionRegistry} exact={true} />
                            <Route path={EXTENSION_ROUTE} component={this.extensionRedirectComponent} />
                        </Switch>
                    </div>
                </div>
            </HashRouter>
        )
    }
}

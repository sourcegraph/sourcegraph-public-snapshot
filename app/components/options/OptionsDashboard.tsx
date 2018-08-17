import * as React from 'react'
import { Route, Switch } from 'react-router'
import { HashRouter } from 'react-router-dom'
import { CXPExtensionRegistry } from './CXPExtensionRegistry'
import { OptionsConfiguration } from './OptionsConfiguration'
import { OptionsPageSidebar } from './OptionsPageSidebar'

export class OptionsDashboard extends React.Component<any, {}> {
    public render(): JSX.Element {
        return (
            <HashRouter>
                <div className="site-admin-area area">
                    <OptionsPageSidebar className="area__sidebar" />
                    <div className="area__content">
                        <Switch>
                            <Route path="/" component={OptionsConfiguration} exact={true} />
                            <Route path="/extensions" component={CXPExtensionRegistry} exact={true} />
                        </Switch>
                    </div>
                </div>
            </HashRouter>
        )
    }
}

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { ThemeProps } from '../../theme'
import { OrgSavedSearchListPage } from '../saved-searches/OrgSavedSearchListPage'
import { OrgAreaPageProps } from './OrgArea'

interface Props extends OrgAreaPageProps, RouteComponentProps<{}>, ThemeProps {
    location: H.Location
}

interface State {
    isCreating: boolean
}

export class OrgSavedSearchesPage extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        this.state = {
            isCreating: false,
        }
    }

    public render(): JSX.Element {
        const transferProps: OrgAreaPageProps = {
            authenticatedUser: this.props.authenticatedUser,
            org: this.props.org,
            onOrganizationUpdate: this.props.onOrganizationUpdate,
            platformContext: this.props.platformContext,
            settingsCascade: this.props.settingsCascade,
        }

        return (
            <>
                <React.Suspense fallback={<LoadingSpinner className="icon-inline m-2" />}>
                    <Switch>
                        <Route
                            path={this.props.match.path}
                            key="hardcoded-key"
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <OrgSavedSearchListPage {...routeComponentProps} {...transferProps} />
                            )}
                        />
                    </Switch>
                </React.Suspense>
            </>
        )
    }
}

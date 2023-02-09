import { Route, Switch } from 'react-router'

import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import { AuthenticatedUser } from '../../auth'
import { SecurityView } from './SecurityView'

export interface SecurityAppRouter {
    authenticatedUser: AuthenticatedUser
}

export const SecurityAppRouter = withAuthenticatedUser<SecurityAppRouter>(props => {
    return (
        <Switch>
            <Route path="/security" component={SecurityView} />
        </Switch>
    )
})

import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { HeroPage } from '../../../components/HeroPage'
import { ThreadsAreaContext } from '../../threads/global/ThreadsArea'
import { ChangesetArea } from '../detail/ChangesetArea'
import { NewChangesetPage } from '../new/NewChangesetPage'
import { ChangesetsOverviewPage } from '../overview/ChangesetsOverviewPage'
import { ChangesetPreviewPage } from '../preview/ChangesetPreviewPage'

const NotFoundPage: React.FunctionComponent = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle={`Sorry, the requested page was not found.`} />
)

/**
 * Properties passed to all page components in the changesets area.
 */
export interface ChangesetsAreaContext extends ThreadsAreaContext {}

interface ChangesetsAreaProps
    extends Pick<ChangesetsAreaContext, Exclude<keyof ChangesetsAreaContext, 'type'>>,
        RouteComponentProps<{}>,
        ExtensionsControllerProps {
    /** The base URL of this area. */
    areaURL?: string
}

/**
 * The changesets area.
 */
export const ChangesetsArea: React.FunctionComponent<ChangesetsAreaProps> = ({ match, ...props }) => {
    const context: ChangesetsAreaContext = {
        ...props,
        type: GQL.ThreadType.CHANGESET,
    }

    return (
        <Switch>
            <Route
                path={match.url}
                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                exact={true}
                // tslint:disable-next-line:jsx-no-lambda
                render={routeComponentProps => <ChangesetsOverviewPage {...routeComponentProps} {...context} />}
            />
            <Route
                path={`${match.url}/new`}
                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                // tslint:disable-next-line:jsx-no-lambda
                render={routeComponentProps => <NewChangesetPage {...routeComponentProps} {...context} />}
            />
            <Route
                path={`${match.url}/preview/:threadID`}
                // tslint:disable-next-line:jsx-no-lambda
                render={routeComponentProps => <ChangesetPreviewPage {...routeComponentProps} {...context} />}
            />
            <Route
                path={`${match.url}/:threadID`}
                // tslint:disable-next-line:jsx-no-lambda
                render={routeComponentProps => <ChangesetArea {...routeComponentProps} {...context} />}
            />
            <Route key="hardcoded-key" component={NotFoundPage} />
        </Switch>
    )
}

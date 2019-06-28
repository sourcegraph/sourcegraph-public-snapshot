import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import * as sourcegraph from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { ErrorBoundary } from '../../../components/ErrorBoundary'
import { HeroPage } from '../../../components/HeroPage'
import { Checklist } from '../checklist'
import { useChecklist } from '../util/WithChecklistQueryResults'
import { ChecklistItemsList } from './itemsList/ChecklistItemsList'

const NotFoundPage = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested page was not found." />
)

interface Props
    extends Pick<ChecklistAreaContext, Exclude<keyof ChecklistAreaContext, 'checklist'>>,
        RouteComponentProps<{}> {
    scope: sourcegraph.ChecklistScope | sourcegraph.WorkspaceRoot
}

export interface ChecklistAreaContext extends ExtensionsControllerProps, PlatformContextProps {
    /** The checklist. */
    checklist: Checklist

    authenticatedUser: GQL.IUser | null
    isLightTheme: boolean
}

const LOADING: 'loading' = 'loading'

/**
 * The area for a single checklist.
 */
export const ChecklistArea: React.FunctionComponent<Props> = ({ scope, ...props }) => {
    const checklistOrError = useChecklist(props.extensionsController, scope)
    if (checklistOrError === LOADING) {
        return null // loading
    }
    if (isErrorLike(checklistOrError)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={checklistOrError.message} />
    }

    const context: ChecklistAreaContext & {
        areaURL: string
    } = {
        ...props,
        checklist: checklistOrError,
        areaURL: props.match.url,
    }

    return (
        <div className="checklist-area">
            <ErrorBoundary location={props.location}>
                <Switch>
                    <Route
                        path={props.match.url}
                        exact={true}
                        // tslint:disable-next-line:jsx-no-lambda
                        render={routeComponentProps => <ChecklistItemsList {...context} {...routeComponentProps} />}
                    />
                    <Route key="hardcoded-key" component={NotFoundPage} />
                </Switch>
            </ErrorBoundary>
        </div>
    )
}

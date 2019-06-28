import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useState } from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { filter, first, map, switchMap } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { ErrorBoundary } from '../../../components/ErrorBoundary'
import { HeroPage } from '../../../components/HeroPage'
import { useEffectAsync } from '../../../util/useEffectAsync'
import { getCodeActions, getDiagnosticInfos } from '../../threads/detail/backend'
import { Checklist } from '../checklist'
import { ChecklistItemsList } from './itemsList/ChecklistItemsList'

const NotFoundPage = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested page was not found." />
)

interface Props
    extends Pick<ChecklistAreaContext, Exclude<keyof ChecklistAreaContext, 'checklist'>>,
        RouteComponentProps<{}> {}

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
export const ChecklistArea: React.FunctionComponent<Props> = props => {
    const [checklistOrError, setChecklistOrError] = useState<typeof LOADING | Checklist | ErrorLike>(LOADING)

    useEffectAsync(async () => {
        try {
            // TODO!(sqs)
            setChecklistOrError(
                await getDiagnosticInfos(props.extensionsController)
                    .pipe(
                        filter(diagnostics => diagnostics.length > 0),
                        first(),
                        map(diagnostics => diagnostics[0]),
                        switchMap(diagnostic =>
                            getCodeActions({ diagnostic, extensionsController: props.extensionsController }).pipe(
                                filter(codeActions => codeActions.length > 0),
                                first(),
                                map(codeActions => ({ diagnostic, codeActions }))
                            )
                        )
                    )
                    .toPromise()
            )
        } catch (err) {
            setChecklistOrError(asError(err))
        }
    }, [props.extensionsController])
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

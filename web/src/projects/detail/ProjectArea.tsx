import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useMemo, useState } from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { map } from 'rxjs/operators'
import { Resizable } from '../../../../shared/src/components/Resizable'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { queryGraphQL } from '../../backend/graphql'
import { ErrorBoundary } from '../../components/ErrorBoundary'
import { HeroPage } from '../../components/HeroPage'
import { ChecksArea } from '../../enterprise/checks/global/ChecksArea'
import { ThreadsArea } from '../../enterprise/threads/global/ThreadsArea'
import { ProjectLabelsPage } from './labels/ProjectLabelsPage'
import { ProjectSettingsPage } from './settings/ProjectSettingsPage'
import { ProjectAreaSidebar } from './sidebar/ProjectAreaSidebar'

const getProject = (idWithoutKind: GQL.IProjectOnQueryArguments['idWithoutKind']): Promise<GQL.IProject> =>
    queryGraphQL(
        gql`
            query Project($idWithoutKind: String!) {
                project(idWithoutKind: $idWithoutKind) {
                    id
                    name
                    url
                }
            }
        `,
        { idWithoutKind }
    )
        .pipe(
            map(({ data, errors }) => {
                if (!data || !data.project || (errors && errors.length > 0)) {
                    throw createAggregateError(errors)
                }
                return data.project
            })
        )
        .toPromise()

const LOADING: 'loading' = 'loading'

const NotFoundPage = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested page was not found." />
)

interface Props extends RouteComponentProps<{ idWithoutKind: string }>, ExtensionsControllerProps {}

export interface ProjectAreaContext extends ExtensionsControllerProps {
    /** The project. */
    project: GQL.IProject

    /** Called to update the project. */
    onProjectUpdate: (project: GQL.IProject) => void
}

/**
 * The area for a single project.
 */
export const ProjectArea: React.FunctionComponent<Props> = props => {
    const [projectOrError, setProjectOrError] = useState<typeof LOADING | GQL.IProject | ErrorLike>(LOADING)

    // tslint:disable-next-line: no-floating-promises beacuse fetchDiscussionProjectAndComments never throws
    useMemo(async () => {
        try {
            setProjectOrError(await getProject(props.match.params.idWithoutKind))
        } catch (err) {
            setProjectOrError(asError(err))
        }
    }, [props.match.params.idWithoutKind])

    if (projectOrError === LOADING) {
        return null // loading
    }
    if (isErrorLike(projectOrError)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={projectOrError.message} />
    }

    const context: ProjectAreaContext & {
        areaURL: string
    } = {
        ...props,
        project: projectOrError,
        onProjectUpdate: setProjectOrError,
        areaURL: props.match.url,
    }

    return (
        <div className="project-area flex-1 d-flex overflow-hidden">
            <ProjectAreaSidebar {...context} className="project-area__sidebar flex-0 border-right" />
            <div className="flex-1 overflow-auto d-flex flex-column">
                <ErrorBoundary location={props.location}>
                    <Switch>
                        <Route
                            path={props.match.url}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={() => <p>Overview!</p>}
                        />
                        <Route
                            path={`${props.match.url}/checks`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => <ChecksArea {...context} {...routeComponentProps} />}
                        />
                        <Route
                            path={`${props.match.url}/threads`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => <ThreadsArea {...context} {...routeComponentProps} />}
                        />
                        <Route
                            path={`${props.match.url}/labels`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => <ProjectLabelsPage {...context} {...routeComponentProps} />}
                        />
                        <Route
                            path={`${props.match.url}/settings`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <ProjectSettingsPage {...context} {...routeComponentProps} />
                            )}
                        />
                        <Route key="hardcoded-key" component={NotFoundPage} />
                    </Switch>
                </ErrorBoundary>
            </div>
        </div>
    )
}

import H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useMemo, useState } from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { isDefined } from '../../../../../shared/src/util/types'
import { BreadcrumbItem, Breadcrumbs } from '../../../components/breadcrumbs/Breadcrumbs'
import { HeroPage } from '../../../components/HeroPage'
import { NamespaceAreaContext } from '../../../namespaces/NamespaceArea'
import { ThemeProps } from '../../../theme'
import { ThreadArea } from '../detail/ThreadArea'
import { NamespaceThreadsListPage } from './list/NamespaceThreadsListPage'
import { ThreadsNewPage } from './new/ThreadsNewPage'

export interface RepositoryThreadsAreaContext extends ExtensionsControllerProps, ThemeProps {
    repository: Pick<GQL.IRepository, 'id' | 'url'>

    /** The URL to the repository threads area. */
    threadsURL: string

    setBreadcrumbItem?: (breadcrumbItem: BreadcrumbItem | undefined) => void

    location: H.Location
    authenticatedUser: GQL.IUser | null
}

interface Props
    extends Pick<
        RepositoryThreadsAreaContext,
        Exclude<keyof RepositoryThreadsAreaContext, 'threadsURL' | 'setBreadcrumbItem'>
    > {}

/**
 * The threads area for a repository.
 */
export const RepositoryThreadsArea: React.FunctionComponent<Props> = ({ ...props }) => {
    const [breadcrumbItem, setBreadcrumbItem] = useState<BreadcrumbItem>()

    const context: RepositoryThreadsAreaContext = {
        ...props,
        threadsURL: `${props.repository.url}/-/threads`,
        setBreadcrumbItem,
    }
    const newThreadURL = `${context.threadsURL}/new`

    const breadcrumbItems: BreadcrumbItem[] = useMemo(
        () =>
            [
                { text: props.namespace.namespaceName, to: props.namespace.url },
                { text: 'Threads', to: context.threadsURL },
                breadcrumbItem,
            ].filter(isDefined),
        [breadcrumbItem, context.threadsURL, props.namespace.namespaceName, props.namespace.url]
    )

    const breadcrumbs = <Breadcrumbs items={breadcrumbItems} className="my-4" />

    return (
        <>
            <Switch>
                <Route path={context.threadsURL} exact={true}>
                    breadcrumbs
                    <NamespaceThreadsListPage {...context} newThreadURL={newThreadURL} />
                </Route>
                <Route path={newThreadURL} exact={true}>
                    breadcrumbs
                    <ThreadsNewPage {...context} />
                </Route>
                <Route
                    path={`${context.threadsURL}/:threadID`}
                    // tslint:disable-next-line:jsx-no-lambda
                    render={(routeComponentProps: RouteComponentProps<{ threadID: string }>) => (
                        <ThreadArea
                            {...context}
                            threadID={routeComponentProps.match.params.threadID}
                            header={breadcrumbs}
                        />
                    )}
                />
                <Route>
                    <HeroPage
                        icon={MapSearchIcon}
                        title="404: Not Found"
                        subtitle="Sorry, the requested page was not found."
                    />
                </Route>
            </Switch>
        </>
    )
}

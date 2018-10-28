import ConsoleIcon from 'mdi-react/ConsoleIcon'
import * as React from 'react'
import { Link, NavLink, RouteComponentProps } from 'react-router-dom'
import * as GQL from '../../../../src/backend/graphqlschema'
import {
    SIDEBAR_BUTTON_CLASS,
    SIDEBAR_CARD_CLASS,
    SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS,
} from '../../../../src/components/Sidebar'

interface Props extends RouteComponentProps<any> {
    className: string
    repo?: GQL.IRepository
    routePrefix: string
}

/**
 * Sidebar for repository graph pages.
 */
export const RepositoryGraphSidebar: React.SFC<Props> = (props: Props) =>
    props.repo ? (
        <div className={`repository-graph-sidebar ${props.className}`}>
            <div className={SIDEBAR_CARD_CLASS}>
                <div className="card-header">Repository graph</div>
                <div className="list-group list-group-flush">
                    <NavLink
                        to={`${props.routePrefix}/-/graph`}
                        exact={true}
                        className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}
                    >
                        Overview
                    </NavLink>
                    <NavLink
                        to={`${props.routePrefix}/-/graph/packages`}
                        exact={true}
                        className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}
                    >
                        Packages
                    </NavLink>
                    <NavLink
                        to={`${props.routePrefix}/-/graph/dependencies`}
                        exact={true}
                        className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}
                    >
                        Dependencies
                    </NavLink>
                </div>
            </div>
            <Link to="/api/console" className={SIDEBAR_BUTTON_CLASS}>
                <ConsoleIcon className="icon-inline" />
                API console
            </Link>
        </div>
    ) : (
        <div className={`repository-graph-sidebar ${props.className}`} />
    )

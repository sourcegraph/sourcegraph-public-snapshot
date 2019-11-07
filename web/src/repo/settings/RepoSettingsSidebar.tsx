import ConsoleIcon from 'mdi-react/ConsoleIcon'
import * as React from 'react'
import { Link, NavLink, RouteComponentProps } from 'react-router-dom'
import * as GQL from '../../../../shared/src/graphql/schema'
import {
    SIDEBAR_BUTTON_CLASS,
    SIDEBAR_CARD_CLASS,
    SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS,
} from '../../components/Sidebar'

interface Props extends RouteComponentProps<any> {
    className?: string
    repo?: GQL.IRepository
}

/**
 * Sidebar for repository settings pages.
 */
export const RepoSettingsSidebar: React.FunctionComponent<Props> = (props: Props) =>
    props.repo ? (
        <div className={`repo-settings-sidebar ${props.className || ''}`}>
            <div className={SIDEBAR_CARD_CLASS}>
                <div className="card-header">Settings</div>
                <div className="list-group list-group-flush">
                    <NavLink
                        to={`/${props.repo.name}/-/settings`}
                        exact={true}
                        className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}
                    >
                        Options
                    </NavLink>
                    <NavLink
                        to={`/${props.repo.name}/-/settings/index`}
                        exact={true}
                        className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}
                    >
                        Indexing
                    </NavLink>
                    <NavLink
                        to={`/${props.repo.name}/-/settings/code-intelligence`}
                        exact={true}
                        className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}
                    >
                        Code intelligence
                    </NavLink>
                    <NavLink
                        to={`/${props.repo.name}/-/settings/mirror`}
                        exact={true}
                        className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}
                    >
                        Mirroring
                    </NavLink>
                </div>
            </div>
            <Link to="/api/console" className={SIDEBAR_BUTTON_CLASS}>
                <ConsoleIcon className="icon-inline" />
                API console
            </Link>
        </div>
    ) : (
        <div className={`repo-settings-sidebar ${props.className || ''}`} />
    )

import ConsoleIcon from 'mdi-react/ConsoleIcon'
import * as React from 'react'
import { Link, NavLink, RouteComponentProps } from 'react-router-dom'
import * as GQL from '../../../../shared/src/graphql/schema'
import {
    SIDEBAR_BUTTON_CLASS,
    SIDEBAR_CARD_CLASS,
    SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS,
} from '../../components/Sidebar'
import { NavItemDescriptor } from '../../util/contributions'

export interface RepoSettingsSideBarItem extends NavItemDescriptor {}

export type RepoSettingsSideBarItems = readonly RepoSettingsSideBarItem[]

interface Props extends RouteComponentProps<{}> {
    repoSettingsSidebarItems: RepoSettingsSideBarItems
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
                    {props.repoSettingsSidebarItems.map(
                        ({ label, to, exact, condition = () => true }) =>
                            condition({}) && (
                                <NavLink
                                    to={`/${props.repo && props.repo.name}/-/settings${to}`}
                                    exact={exact}
                                    key={label}
                                    className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}
                                >
                                    {label}
                                </NavLink>
                            )
                    )}
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

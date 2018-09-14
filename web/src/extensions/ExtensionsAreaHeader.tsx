import ArrowLeftIcon from 'mdi-react/ArrowLeftIcon'
import * as React from 'react'
import { Link, RouteComponentProps } from 'react-router-dom'
import { ActionButtonDescriptor } from '../util/contributions'
import { ExtensionsAreaRouteContext } from './ExtensionsArea'

export interface ExtensionsAreaHeaderProps extends ExtensionsAreaRouteContext, RouteComponentProps<{}> {
    isPrimaryHeader: boolean
    actionButtons: ReadonlyArray<ExtensionsAreaHeaderActionButton>
}

export interface ExtensionAreaHeaderContext {
    isPrimaryHeader: boolean
}

export interface ExtensionsAreaHeaderActionButton extends ActionButtonDescriptor<ExtensionAreaHeaderContext> {}

/**
 * Header for the extensions area.
 */
export const ExtensionsAreaHeader: React.SFC<ExtensionsAreaHeaderProps> = (props: ExtensionsAreaHeaderProps) => (
    <div className="border-bottom simple-area-header">
        <div className="navbar navbar-expand py-2">
            <div className="container">
                {props.isPrimaryHeader && (
                    <h3 className="mb-0">
                        <Link className="nav-brand mr-2" to="/extensions">
                            <strong>Extensions</strong>
                        </Link>
                    </h3>
                )}
                <ul className="navbar-nav nav">
                    {!props.isPrimaryHeader && (
                        <li className="nav-item">
                            <Link to="/extensions" className="nav-link">
                                <ArrowLeftIcon className="icon-inline" /> All extensions
                            </Link>
                        </li>
                    )}
                </ul>
                <div className="spacer" />
                <ul className="navbar-nav nav">
                    {props.actionButtons.map(
                        ({ condition = () => true, to, icon: Icon, label, tooltip }) =>
                            condition(props) && (
                                <li className="nav-item" key={label}>
                                    <Link className="btn ml-2 btn-secondary" to={to(props)} data-tooltip={tooltip}>
                                        {Icon && <Icon className="icon-inline" />} {label}
                                    </Link>
                                </li>
                            )
                    )}
                </ul>
            </div>
        </div>
    </div>
)

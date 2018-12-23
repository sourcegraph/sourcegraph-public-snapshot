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
export const ExtensionsAreaHeader: React.FunctionComponent<ExtensionsAreaHeaderProps> = (
    props: ExtensionsAreaHeaderProps
) => (
    <div>
        {props.isPrimaryHeader && (
            <div className="navbar navbar-expand">
                <div className="container">
                    <h2 className="mb-0">Extensions</h2>
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
        )}
    </div>
)

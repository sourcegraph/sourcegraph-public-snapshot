import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'
import * as React from 'react'
import { Link, RouteComponentProps } from 'react-router-dom'

import { PageHeader } from '../components/PageHeader'
import { ActionButtonDescriptor } from '../util/contributions'

import { ExtensionsAreaRouteContext } from './ExtensionsArea'

export interface ExtensionsAreaHeaderProps extends ExtensionsAreaRouteContext, RouteComponentProps<{}> {
    isPrimaryHeader: boolean
    actionButtons: readonly ExtensionsAreaHeaderActionButton[]
}

export interface ExtensionAreaHeaderContext {
    isPrimaryHeader: boolean
}

export interface ExtensionsAreaHeaderActionButton extends ActionButtonDescriptor<ExtensionAreaHeaderContext> {}

/**
 * Header for the extensions area.
 */
export const ExtensionsAreaHeader: React.FunctionComponent<ExtensionsAreaHeaderProps> = props => (
    <div className="container">
        {props.isPrimaryHeader && (
            <PageHeader
                path={[{ icon: PuzzleOutlineIcon, text: 'Extensions' }]}
                actions={props.actionButtons.map(
                    ({ condition = () => true, to, icon: Icon, label, tooltip }) =>
                        condition(props) && (
                            <Link className="btn ml-2 btn-secondary" to={to(props)} data-tooltip={tooltip} key={label}>
                                {Icon && <Icon className="icon-inline" />} {label}
                            </Link>
                        )
                )}
                byline="Improve your workflow with code intelligence, test coverage, and other useful information."
            />
        )}
    </div>
)

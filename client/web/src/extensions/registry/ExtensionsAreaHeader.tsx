import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'
import React from 'react'
import { RouteComponentProps } from 'react-router-dom'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { PageHeader } from '@sourcegraph/wildcard'

import { ActionButtonDescriptor } from '../../util/contributions'

import { ExtensionsAreaRouteContext } from './ExtensionsArea'

export interface ExtensionsAreaHeaderProps extends ExtensionsAreaRouteContext, RouteComponentProps<{}> {
    isPrimaryHeader: boolean
    actionButtons: readonly ExtensionsAreaHeaderActionButton[]
}

export interface ExtensionAreaHeaderContext {
    isPrimaryHeader: boolean
}

export interface ExtensionsAreaHeaderActionButton extends ActionButtonDescriptor<ExtensionAreaHeaderContext> {}

export const extensionsAreaHeaderActionButtons: ExtensionsAreaHeaderActionButton[] = []

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
            />
        )}
    </div>
)

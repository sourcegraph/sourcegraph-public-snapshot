import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'
import * as React from 'react'
import { Link, RouteComponentProps } from 'react-router-dom'

import { PageHeader, Button } from '@sourcegraph/wildcard'

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
                            <Button
                                className="ml-2"
                                to={to(props)}
                                data-tooltip={tooltip}
                                key={label}
                                variant="secondary"
                                as={Link}
                            >
                                {Icon && <Icon className="icon-inline" />} {label}
                            </Button>
                        )
                )}
            />
        )}
    </div>
)

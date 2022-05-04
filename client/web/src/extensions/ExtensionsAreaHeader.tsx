import * as React from 'react'

import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'
import { RouteComponentProps } from 'react-router-dom'

import { PageHeader, Button, Link, Icon } from '@sourcegraph/wildcard'

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
export const ExtensionsAreaHeader: React.FunctionComponent<
    React.PropsWithChildren<ExtensionsAreaHeaderProps>
> = props => (
    <div className="container">
        {props.isPrimaryHeader && (
            <PageHeader
                path={[{ icon: PuzzleOutlineIcon, text: 'Extensions' }]}
                actions={props.actionButtons.map(
                    ({ condition = () => true, to, icon: ButtonIcon, label, tooltip }) =>
                        condition(props) && (
                            <Button
                                className="ml-2"
                                to={to(props)}
                                data-tooltip={tooltip}
                                key={label}
                                variant="secondary"
                                as={Link}
                            >
                                {ButtonIcon && <Icon as={ButtonIcon} />} {label}
                            </Button>
                        )
                )}
            />
        )}
    </div>
)

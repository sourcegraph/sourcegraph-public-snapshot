import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'

/**
 * A container for multiple OverviewItem components.
 */
export const OverviewList: React.FunctionComponent<{ children: React.ReactNode | React.ReactNode[] }> = ({
    children,
}) => <ul className="overview-list">{children}</ul>

export interface Props {
    link?: string

    /**
     * Content in the overview item's always-visible, single-line title bar.
     */
    title: string

    /**
     * Optional children that appear below the item's title bar that can be expanded/collapsed.
     * If present, a "View more" button that expands or collapses the children will be added to the item's actions.
     */
    children?: React.ReactNode | React.ReactNode[]

    actions?: string

    /**
     * Whether the item's children are expanded and visible by default.
     */
    defaultExpanded?: boolean
}

export interface State {
    expanded: boolean
}

/**
 * A row item used for an overview page, with an icon, linked elements, and right-hand actions.
 */
export class OverviewItem extends React.Component<Props, State> {
    public state: State = { expanded: this.props.defaultExpanded || false }

    public render(): JSX.Element | null {
        let e: React.ReactFragment = <>{this.props.title}</>
        let actions
        if (this.props.actions !== undefined) {
            actions = (
                <div className="overview-item__actions" data-tooltip={this.props.actions}>
                    <span className="overview-item__button">
                        <OpenInNewIcon className="icon-inline" />
                    </span>
                </div>
            )
        }
        if (this.props.children !== undefined) {
            e = (
                <div className="overview-item__header-link" onClick={this.toggleExpand}>
                    {e}
                </div>
            )
            actions = (
                <div className="overview-item__actions">
                    <span className="icon-click-area" onClick={this.toggleExpand} />
                    <div className="overview-item__toggle-icon">
                        {this.state.expanded ? (
                            <ChevronUpIcon className="icon-inline" aria-label="Close section" />
                        ) : (
                            <ChevronDownIcon className="icon-inline" aria-label="Expand section" />
                        )}
                    </div>
                    {actions}
                </div>
            )
        } else if (this.props.link !== undefined) {
            e = (
                <Link to={this.props.link} className="overview-item__header-link">
                    {e}
                </Link>
            )
        }

        if (this.props.link !== undefined && this.props.children === undefined) {
            return (
                <Link to={this.props.link} className="overview-item">
                    <div className="overview-item__header">{this.props.title}</div>
                    {actions}
                    {this.props.children && this.state.expanded && (
                        <div className="overview-item__children">{this.props.children}</div>
                    )}
                </Link>
            )
        } else {
            return (
                <div className="overview-item">
                    <div className="overview-item__header">{e}</div>
                    {actions}
                    {this.props.children && this.state.expanded && (
                        <div className="overview-item__children">{this.props.children}</div>
                    )}
                </div>
            )
        }
    }

    private toggleExpand = () => this.setState(prevState => ({ expanded: !prevState.expanded }))
}

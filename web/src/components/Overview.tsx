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

    actions?: React.ReactFragment
    icon?: React.ComponentType<{ className?: string }>

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
        let e: React.ReactFragment = (
            <>
                {this.props.icon && <this.props.icon className="icon-inline overview-item__header-icon" />}
                {this.props.title}
            </>
        )
        let actions = this.props.actions
        if (this.props.children !== undefined) {
            e = (
                <div className="overview-item__header-link" onClick={this.toggleExpand}>
                    {e}
                </div>
            )
            actions = (
                <>
                    <button className="btn btn-secondary btn-sm" onClick={this.toggleExpand}>
                        {this.state.expanded ? 'Hide' : 'View more'}
                    </button>
                    {actions && actions}
                </>
            )
        } else if (this.props.link !== undefined) {
            e = (
                <Link to={this.props.link} className="overview-item__header-link">
                    {e}
                </Link>
            )
        }

        return (
            <div className="overview-item">
                <div className="overview-item__header">{e}</div>
                {actions && <div className="overview-item__actions">{actions}</div>}
                {this.props.children && this.state.expanded && (
                    <div className="overview-item__children mt-4 mb-2">{this.props.children}</div>
                )}
            </div>
        )
    }

    private toggleExpand = () => this.setState(prevState => ({ expanded: !prevState.expanded }))
}

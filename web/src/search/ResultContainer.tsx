import ChevronDownIcon from '@sourcegraph/icons/lib/ChevronDown'
import ChevronRightIcon from '@sourcegraph/icons/lib/ChevronRight'
import ChevronUpIcon from '@sourcegraph/icons/lib/ChevronUp'
import * as React from 'react'

interface Props {
    /**
     * Whether the result container's children are visible by default.
     * The header is always visible even when the component is not expanded.
     */
    defaultExpanded?: boolean

    /**
     * Whether the result container can be collapsed. If false, its children
     * are always displayed, and no expand/collapse actions are shown.
     */
    collapsible?: boolean

    /**
     * The icon to show left to the title.
     */
    icon: React.ComponentType<{ className: string }>

    /**
     * The title component.
     */
    title: React.ReactFragment

    /**
     * The content of the result displayed underneath the result container's
     * header when collapsed.
     */
    collapsedChildren?: React.ReactFragment

    /**
     * The content of the result displayed underneath the result container's
     * header when expanded.
     */
    expandedChildren?: React.ReactFragment

    /**
     * The label to display next to the collapse button
     */
    collapseLabel?: string

    /**
     * The label to display next to the expand button
     */
    expandLabel?: string

    /**
     * This component does not accept children.
     */
    children?: never
}

interface State {
    /**
     * Whether this result container is currently expanded.
     */
    expanded?: boolean
}

/**
 * The container component for a result in the SearchResults component.
 */
export class ResultContainer extends React.PureComponent<Props, State> {
    constructor(props: Props) {
        super(props)
        this.state = { expanded: this.props.defaultExpanded }
    }

    public render(): JSX.Element | null {
        const Icon = this.props.icon
        return (
            <div className="result-container">
                <div
                    className={
                        'result-container__header' +
                        (this.props.collapsible ? ' result-container__header--collapsible' : '')
                    }
                    onClick={this.toggle}
                >
                    <Icon className="icon-inline" />
                    <div className="result-container__header-title">{this.props.title}</div>
                    {this.props.collapsible &&
                        (this.state.expanded ? (
                            <span className="result-container__toggle-matches-container">
                                {this.props.collapseLabel}
                                {this.props.collapseLabel && <ChevronUpIcon className="icon-inline" />}
                                {!this.props.collapseLabel && <ChevronDownIcon className="icon-inline" />}
                            </span>
                        ) : (
                            <span className="result-container__toggle-matches-container">
                                {this.props.expandLabel}
                                {this.props.expandLabel && <ChevronDownIcon className="icon-inline" />}
                                {!this.props.expandLabel && <ChevronRightIcon className="icon-inline" />}
                            </span>
                        ))}
                </div>
                {!this.state.expanded && this.props.collapsedChildren}
                {this.state.expanded && this.props.expandedChildren}
            </div>
        )
    }

    private toggle = () => {
        this.setState({ expanded: !this.state.expanded })
    }
}

import ChevronDownIcon from '@sourcegraph/icons/lib/ChevronDown'
import ChevronRightIcon from '@sourcegraph/icons/lib/ChevronRight'
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
    title: React.ReactChild

    /**
     * The main content of the result, displayed underneath the result
     * container's header.
     */
    children?: React.ReactChild
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
                            <ChevronDownIcon className="icon-inline" />
                        ) : (
                            <ChevronRightIcon className="icon-inline" />
                        ))}
                </div>
                {this.state.expanded && this.props.children}
            </div>
        )
    }

    private toggle = () => {
        this.setState({ expanded: !this.state.expanded })
    }
}

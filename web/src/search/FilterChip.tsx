import truncate from 'lodash/truncate'
import * as React from 'react'
import { queryIndexOfScope } from './helpers'

interface Props {
    name?: string
    value: string
    query: string
    onFilterChosen: (filter: string) => void
}

export class FilterChip extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        const truncatedValue = truncate(this.props.value, { length: 50 })

        return (
            <button
                className={
                    'btn btn-sm filter-chip' +
                    (this.isScopeSelected(this.props.query, this.props.value) ? ' filter-chip--selected' : '')
                }
                value={this.props.value}
                // only display tooltip if chip shows truncated value or scope name
                data-tooltip={
                    this.isScopeSelected(this.props.query, this.props.value)
                        ? 'Already added to query'
                        : this.props.value !== truncatedValue || this.props.name ? this.props.value : null
                }
                onMouseDown={this.onMouseDown}
                onClick={this.onClick}
            >
                {this.props.name || truncatedValue}
            </button>
        )
    }

    private isScopeSelected(query: string, scope: string): boolean {
        return queryIndexOfScope(query, scope) !== -1
    }

    private onMouseDown: React.MouseEventHandler<HTMLButtonElement> = event => {
        // prevent clicking on chips from taking focus away from the search input.
        event.preventDefault()
    }

    private onClick: React.MouseEventHandler<HTMLButtonElement> = event => {
        event.preventDefault()
        this.props.onFilterChosen(event.currentTarget.value)
    }
}

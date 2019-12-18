import React from 'react'
import * as H from 'history'
import RegexIcon from 'mdi-react/RegexIcon'
import { submitSearch } from '../helpers'
import { SearchPatternType } from '../../../../shared/src/graphql/schema'
import { Subscription, fromEvent } from 'rxjs'
import { filter } from 'rxjs/operators'
import { PatternTypeProps } from '..'
import { FiltersToTypeAndValue } from '../../../../shared/src/search/interactive/util'

interface RegexpToggleProps extends PatternTypeProps {
    navbarSearchQuery: string
    history: H.History
    filtersInQuery?: FiltersToTypeAndValue
    hasGlobalQueryBehavior?: boolean
}

export class RegexpToggle extends React.Component<RegexpToggleProps> {
    private subscriptions = new Subscription()
    private toggleCheckbox = React.createRef<HTMLDivElement>()

    public componentDidMount(): void {
        this.subscriptions.add(
            fromEvent<KeyboardEvent>(window, 'keydown')
                .pipe(
                    filter(
                        event =>
                            document.activeElement === this.toggleCheckbox.current &&
                            (event.keyCode === 13 || event.keyCode === 32)
                    )
                )
                .subscribe(event => {
                    event.preventDefault()
                    this.toggle()
                })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const isRegexp = this.props.patternType === SearchPatternType.regexp
        return (
            <div
                ref={this.toggleCheckbox}
                onClick={this.toggle}
                className="btn btn-icon icon-inline regexp-toggle e2e-regexp-toggle"
                role="checkbox"
                aria-checked={isRegexp}
                aria-label="Regular expression toggle"
                tabIndex={0}
                data-tooltip={`${isRegexp ? 'Disable' : 'Enable'} regular expressions`}
            >
                <span className={isRegexp ? 'regexp-toggle--active e2e-regexp-toggle--active' : ''}>
                    <RegexIcon />
                </span>
            </div>
        )
    }

    private toggle = (): void => {
        const newPatternType =
            this.props.patternType !== SearchPatternType.regexp ? SearchPatternType.regexp : SearchPatternType.literal

        this.props.setPatternType(newPatternType)
        if (this.props.hasGlobalQueryBehavior) {
            // We only want the regexp toggle to submit searches if the query input it is in
            // has global behavior (i.e. query inputs on the main search page or global navbar). Non-global inputs
            // don't have the canonical query, and are dependent on the page it's on for context, which makes the
            // submit on-toggle behavior undesirable.
            submitSearch(
                this.props.history,
                this.props.navbarSearchQuery,
                'filter',
                newPatternType,
                undefined,
                this.props.filtersInQuery
            )
        }
    }
}

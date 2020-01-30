import React from 'react'
import * as H from 'history'
import FormatLetterCaseIcon from 'mdi-react/FormatLetterCaseIcon'
import { submitSearch } from '../helpers'
import { Subscription, fromEvent } from 'rxjs'
import { filter } from 'rxjs/operators'
import { PatternTypeProps, CaseSensitivityProps, parseSearchURLQuery } from '..'
import { FiltersToTypeAndValue } from '../../../../shared/src/search/interactive/util'
import { isEmpty } from 'lodash'

interface Props extends PatternTypeProps, CaseSensitivityProps {
    navbarSearchQuery: string
    history: H.History
    location: H.Location
    filtersInQuery?: FiltersToTypeAndValue
    hasGlobalQueryBehavior?: boolean
}

export class CaseSensitivityToggle extends React.Component<Props> {
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
        const isCaseSensitive = this.props.caseSensitive
        return (
            <div
                ref={this.toggleCheckbox}
                onClick={this.toggle}
                className="btn btn-icon icon-inline query-input2__toggle e2e-case-sensitivity-toggle"
                role="checkbox"
                aria-checked={isCaseSensitive}
                aria-label="Case sensitivity toggle"
                tabIndex={0}
                data-tooltip={`${isCaseSensitive ? 'Disable' : 'Enable'} case sensitivity`}
            >
                <span
                    className={`query-input__toggle-icon ${isCaseSensitive ? 'query-input2__toggle-icon--active' : ''}`}
                >
                    <FormatLetterCaseIcon />
                </span>
            </div>
        )
    }

    private toggle = (): void => {
        const isSearchHomepage =
            this.props.location.pathname === '/search' && !parseSearchURLQuery(this.props.location.search)

        const shouldSubmitSearchOnHomepage =
            this.props.navbarSearchQuery !== '' || (this.props.filtersInQuery && !isEmpty(this.props.filtersInQuery))
        const newCaseSensitivity = !this.props.caseSensitive
        this.props.setCaseSensitivity(newCaseSensitivity)

        const shouldSubmitSearch =
            this.props.hasGlobalQueryBehavior &&
            ((isSearchHomepage && shouldSubmitSearchOnHomepage) || !isSearchHomepage)
        if (shouldSubmitSearch) {
            // We only want the toggle to submit searches if the query input it is in
            // has global behavior (i.e. query inputs on the main search page or global navbar). Non-global inputs
            // don't have the canonical query, and are dependent on the page it's on for context, which makes the
            // submit-on-toggle behavior undesirable.
            //
            // Also, we only want to submit a search when toggling on the search homepage when the query is non-empty.
            submitSearch(
                this.props.history,
                this.props.navbarSearchQuery,
                'filter',
                this.props.patternType,
                newCaseSensitivity,
                undefined,
                this.props.filtersInQuery
            )
        }
    }
}

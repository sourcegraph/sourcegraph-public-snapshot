import React from 'react'
import * as H from 'history'
import RegexIcon from 'mdi-react/RegexIcon'
import { submitSearch } from '../helpers'
import { SearchPatternType } from '../../../../shared/src/graphql/schema'
import { Subscription, fromEvent } from 'rxjs'
import { filter } from 'rxjs/operators'

interface RegexpToggleProps {
    toggled: boolean
    togglePatternType: () => void
    patternType: SearchPatternType
    navbarSearchQuery: string
    history: H.History
}

export default class RegexpToggle extends React.Component<RegexpToggleProps> {
    private subscriptions = new Subscription()
    private toggleCheckbox = React.createRef<HTMLDivElement>()

    constructor(props: RegexpToggleProps) {
        super(props)
    }

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

    public render(): JSX.Element | null {
        return (
            <div
                ref={this.toggleCheckbox}
                onClick={this.toggle}
                className="btn btn-icon icon-inline regexp-toggle e2e-regexp-toggle"
                role="checkbox"
                aria-checked={this.props.toggled}
                aria-label="Regular expression toggle"
                tabIndex={0}
                data-tooltip={`${this.props.toggled ? 'Disable' : 'Enable'} regular expressions`}
            >
                <span className={this.props.toggled ? 'regexp-toggle--active e2e-regexp-toggle--active' : ''}>
                    <RegexIcon />
                </span>
            </div>
        )
    }

    private toggle = (): void => {
        const newPatternType = this.props.toggled ? SearchPatternType.literal : SearchPatternType.regexp
        this.props.togglePatternType()
        submitSearch(this.props.history, this.props.navbarSearchQuery, 'filter', newPatternType)
    }
}

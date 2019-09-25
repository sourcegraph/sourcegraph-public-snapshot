import React from 'react'
import * as H from 'history'
import RegexIcon from 'mdi-react/RegexIcon'
import { submitSearch } from '../helpers'
import { SearchPatternType } from '../../../../shared/src/graphql/schema'

interface RegexpToggleProps {
    togglePatternType: (patternType: SearchPatternType) => void
    patternType: SearchPatternType
    navbarSearchQuery: string
    history: H.History
}

export default class RegexpToggle extends React.Component<RegexpToggleProps> {
    constructor(props: RegexpToggleProps) {
        super(props)
    }

    public render(): JSX.Element | null {
        return (
            <button
                onClick={this.toggle}
                className="btn btn-icon icon-inline regexp-toggle e2e-regexp-toggle"
                type="button"
            >
                <span
                    className={`${
                        this.props.patternType === 'regexp' ? 'regexp-toggle--active e2e-regexp-toggle--active' : ''
                    }`}
                >
                    <RegexIcon />
                </span>
            </button>
        )
    }

    private toggle = (e: React.MouseEvent): void => {
        const newPatternType =
            this.props.patternType === SearchPatternType.literal ? SearchPatternType.regexp : SearchPatternType.literal
        this.props.togglePatternType(newPatternType)
        submitSearch(this.props.history, this.props.navbarSearchQuery, 'filter', newPatternType)
    }
}

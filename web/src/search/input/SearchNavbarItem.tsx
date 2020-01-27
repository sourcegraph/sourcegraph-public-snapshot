import * as H from 'history'
import React from 'react'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import { Form } from '../../components/Form'
import { submitSearch, QueryState } from '../helpers'
import { SearchButton } from './SearchButton'
import { PatternTypeProps, CaseSensitivityProps, SmartSearchFieldProps } from '..'
import { LazyMonacoQueryInput } from './LazyMonacoQueryInput'
import { QueryInput } from './QueryInput'
import { ThemeProps } from '../../../../shared/src/theme'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'

interface Props
    extends ActivationProps,
        PatternTypeProps,
        CaseSensitivityProps,
        SmartSearchFieldProps,
        SettingsCascadeProps,
        ThemeProps {
    location: H.Location
    history: H.History
    navbarSearchState: QueryState
    onChange: (newValue: QueryState) => void
}

/**
 * The search item in the navbar
 */
export class SearchNavbarItem extends React.PureComponent<Props> {
    private onSubmit = (): void => {
        const { history, navbarSearchState, patternType, activation, caseSensitive } = this.props
        submitSearch(history, navbarSearchState.query, 'nav', patternType, caseSensitive, activation)
    }

    private onFormSubmit = (e: React.FormEvent): void => {
        e.preventDefault()
        this.onSubmit()
    }

    public render(): React.ReactNode {
        return (
            <Form
                className="search search--navbar-item d-flex align-items-flex-start flex-grow-1"
                onSubmit={this.onFormSubmit}
            >
                {this.props.smartSearchField ? (
                    <LazyMonacoQueryInput
                        {...this.props}
                        hasGlobalQueryBehavior={true}
                        queryState={this.props.navbarSearchState}
                        onSubmit={this.onSubmit}
                        autoFocus={true}
                    ></LazyMonacoQueryInput>
                ) : (
                    <QueryInput
                        {...this.props}
                        value={this.props.navbarSearchState}
                        autoFocus={this.props.location.pathname === '/search' ? 'cursor-at-end' : undefined}
                        hasGlobalQueryBehavior={true}
                    ></QueryInput>
                )}
                <SearchButton />
            </Form>
        )
    }
}

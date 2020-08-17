import * as H from 'history'
import React from 'react'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import { Form } from '../../components/Form'
import { submitSearch, QueryState } from '../helpers'
import { SearchButton } from './SearchButton'
import { PatternTypeProps, CaseSensitivityProps, CopyQueryButtonProps } from '..'
import { LazyMonacoQueryInput } from './LazyMonacoQueryInput'
import { ThemeProps } from '../../../../shared/src/theme'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { VersionContextProps } from '../../../../shared/src/search/util'

interface Props
    extends ActivationProps,
        PatternTypeProps,
        CaseSensitivityProps,
        SettingsCascadeProps,
        ThemeProps,
        CopyQueryButtonProps,
        VersionContextProps {
    location: H.Location
    history: H.History
    navbarSearchState: QueryState
    onChange: (newValue: QueryState) => void
    globbing: boolean
}

/**
 * The search item in the navbar
 */
export class SearchNavbarItem extends React.PureComponent<Props> {
    private onSubmit = (): void => {
        submitSearch({ ...this.props, query: this.props.navbarSearchState.query, source: 'nav' })
    }

    private onFormSubmit = (event: React.FormEvent): void => {
        event.preventDefault()
        this.onSubmit()
    }

    public render(): React.ReactNode {
        return (
            <Form
                className="search--navbar-item d-flex align-items-flex-start flex-grow-1 flex-shrink-past-contents"
                onSubmit={this.onFormSubmit}
            >
                <LazyMonacoQueryInput
                    {...this.props}
                    hasGlobalQueryBehavior={true}
                    queryState={this.props.navbarSearchState}
                    onSubmit={this.onSubmit}
                    autoFocus={true}
                />
                <SearchButton />
            </Form>
        )
    }
}

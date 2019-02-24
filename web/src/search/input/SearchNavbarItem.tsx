import * as H from 'history'
import * as React from 'react'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import { Form } from '../../components/Form'
import { submitSearch } from '../helpers'
import { QueryInput } from './QueryInput'
import { SearchButton } from './SearchButton'

interface Props extends ActivationProps {
    location: H.Location
    history: H.History
    navbarSearchQuery: string
    onChange: (newValue: string) => void
}
/**
 * The search item in the navbar
 */
export class SearchNavbarItem extends React.Component<Props> {
    public render(): JSX.Element | null {
        // Only autofocus the query input on search result pages (otherwise we
        // capture down-arrow keypresses that the user probably intends to scroll down
        // in the page).
        const autoFocus = this.props.location.pathname === '/search'

        return (
            <Form className="search search--navbar-item d-flex" onSubmit={this.onSubmit}>
                <QueryInput
                    {...this.props}
                    value={this.props.navbarSearchQuery}
                    onChange={this.props.onChange}
                    autoFocus={autoFocus ? 'cursor-at-end' : undefined}
                    hasGlobalQueryBehavior={true}
                />
                <SearchButton activation={this.props.activation} />
            </Form>
        )
    }

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        submitSearch(this.props.history, this.props.navbarSearchQuery, 'nav')
    }
}

import React from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import { EXTENSION_CATEGORIES } from '../../../shared/src/schema/extension.schema'
import { extensionsQuery } from './extension/extension'

interface Props {
    /** The current extensions registry list query. */
    query: string

    /** Called when the query changes as a result of user interaction with this component. */
    onQueryChange: (query: string) => void
}

interface State {
    /** Whether the categories dropdown is open. */
    isOpen: boolean
}

/**
 * Displays buttons to be rendered alongside the extension registry list query input field.
 */
export class ExtensionsQueryInputToolbar extends React.PureComponent<Props, State> {
    public state: State = { isOpen: false }

    public render(): JSX.Element | null {
        return (
            <>
                <ButtonDropdown isOpen={this.state.isOpen} toggle={this.toggleIsOpen}>
                    <DropdownToggle caret={true}>Category</DropdownToggle>
                    <DropdownMenu right={true}>
                        {EXTENSION_CATEGORIES.map((c, i) => (
                            <DropdownItem
                                key={i}
                                // tslint:disable-next-line:jsx-no-lambda
                                onClick={() => this.props.onQueryChange(extensionsQuery({ category: c }))}
                                disabled={this.props.query === extensionsQuery({ category: c })}
                            >
                                {c}
                            </DropdownItem>
                        ))}
                    </DropdownMenu>
                </ButtonDropdown>
            </>
        )
    }

    private toggleIsOpen = () => this.setState(prevState => ({ isOpen: !prevState.isOpen }))
}

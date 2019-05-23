import React from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import { EXTENSION_CATEGORIES } from '../../../shared/src/schema/extensionSchema'
import { extensionsQuery } from './extension/extension'

interface Props {
    /** The current extensions registry list query. */
    query: string

    /** Called when the query changes as a result of user interaction with this component. */
    onQueryChange: (query: string) => void
}

type DropdownMenuID = 'categories' | 'options'

interface State {
    /** Which dropdown is open (if any). */
    open?: DropdownMenuID
}

/**
 * Displays buttons to be rendered alongside the extension registry list query input field.
 */
export class ExtensionsQueryInputToolbar extends React.PureComponent<Props, State> {
    public state: State = {}

    public render(): JSX.Element | null {
        return (
            <>
                <ButtonDropdown
                    isOpen={this.state.open === 'categories'}
                    // tslint:disable-next-line:jsx-no-lambda
                    toggle={() => this.toggleIsOpen('categories')}
                >
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
                </ButtonDropdown>{' '}
                <ButtonDropdown
                    isOpen={this.state.open === 'options'}
                    // tslint:disable-next-line:jsx-no-lambda
                    toggle={() => this.toggleIsOpen('options')}
                >
                    <DropdownToggle caret={true}>Options</DropdownToggle>
                    <DropdownMenu right={true}>
                        <DropdownItem
                            // tslint:disable-next-line:jsx-no-lambda
                            onClick={() => this.props.onQueryChange(extensionsQuery({ enabled: true }))}
                            disabled={this.props.query.includes(extensionsQuery({ enabled: true }))}
                        >
                            Show enabled extensions
                        </DropdownItem>
                        <DropdownItem
                            // tslint:disable-next-line:jsx-no-lambda
                            onClick={() => this.props.onQueryChange(extensionsQuery({ disabled: true }))}
                            disabled={this.props.query.includes(extensionsQuery({ disabled: true }))}
                        >
                            Show disabled extensions
                        </DropdownItem>
                    </DropdownMenu>
                </ButtonDropdown>
            </>
        )
    }

    private toggleIsOpen = (menu: DropdownMenuID) =>
        this.setState(prevState => ({ open: prevState.open === menu ? undefined : menu }))
}

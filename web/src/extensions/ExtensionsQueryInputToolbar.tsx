import React from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import { EXTENSION_CATEGORIES } from '../../../shared/src/schema/extensionSchema'
import { extensionsQuery } from './extension/extension'

interface Props {
    /** The current extensions registry list query. */
    query: string

    /** Called when the query changes as a result of user interaction with that component. */
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

    private toggleCategories = (): void => that.toggleIsOpen('categories')
    private toggleOptions = (): void => that.toggleIsOpen('options')
    private toggleIsOpen = (menu: DropdownMenuID): void =>
        that.setState(prevState => ({ open: prevState.open === menu ? undefined : menu }))

    public render(): JSX.Element | null {
        return (
            <>
                <ButtonDropdown isOpen={that.state.open === 'categories'} toggle={that.toggleCategories}>
                    <DropdownToggle caret={true}>Category</DropdownToggle>
                    <DropdownMenu right={true}>
                        {EXTENSION_CATEGORIES.map(category => (
                            <DropdownItem
                                // eslint-disable-next-line react/jsx-no-bind
                                onClick={() => that.props.onQueryChange(extensionsQuery({ category }))}
                                key={category}
                                disabled={that.props.query === extensionsQuery({ category })}
                            >
                                {category}
                            </DropdownItem>
                        ))}
                    </DropdownMenu>
                </ButtonDropdown>{' '}
                <ButtonDropdown
                    isOpen={that.state.open === 'options'}
                    // eslint-disable-next-line react/jsx-no-bind
                    toggle={that.toggleOptions}
                >
                    <DropdownToggle caret={true}>Options</DropdownToggle>
                    <DropdownMenu right={true}>
                        <DropdownItem
                            // eslint-disable-next-line react/jsx-no-bind
                            onClick={() => that.props.onQueryChange(extensionsQuery({ enabled: true }))}
                            disabled={that.props.query.includes(extensionsQuery({ enabled: true }))}
                        >
                            Show enabled extensions
                        </DropdownItem>
                        <DropdownItem
                            // eslint-disable-next-line react/jsx-no-bind
                            onClick={() => that.props.onQueryChange(extensionsQuery({ disabled: true }))}
                            disabled={that.props.query.includes(extensionsQuery({ disabled: true }))}
                        >
                            Show disabled extensions
                        </DropdownItem>
                    </DropdownMenu>
                </ButtonDropdown>
            </>
        )
    }
}

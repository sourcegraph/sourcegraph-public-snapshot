import React from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import { EXTENSION_CATEGORIES, ExtensionCategory } from '../../../shared/src/schema/extensionSchema'
import { extensionsQuery } from './extension/extension'

interface Props {
    /** The current extensions registry list query. */
    query: string

    /** Called when the query changes as a result of user interaction with this component. */
    onQueryChange: (query: string) => void

    /**  */
    selectedCategories: ExtensionCategory[]

    /**  */
    setSelectedCategories: React.Dispatch<React.SetStateAction<ExtensionCategory[]>>
}

type DropdownMenuID = 'categories' | 'options'

interface State {
    /** Which dropdown is open (if any). */
    open?: DropdownMenuID
}

export const NewExtensionsQueryInputToolbar: React.FunctionComponent<Props> = () => <div />

/**
 * Displays buttons to be rendered alongside the extension registry list query input field.
 */
export class ExtensionsQueryInputToolbar extends React.PureComponent<Props, State> {
    public state: State = {}

    private toggleCategories = (): void => this.toggleIsOpen('categories')
    private toggleOptions = (): void => this.toggleIsOpen('options')
    private toggleIsOpen = (menu: DropdownMenuID): void =>
        this.setState(previousState => ({ open: previousState.open === menu ? undefined : menu }))

    public render(): JSX.Element | null {
        return (
            <>
                {EXTENSION_CATEGORIES.map(category => {
                    const selected = this.props.selectedCategories.includes(category)
                    return (
                        <button
                            type="button"
                            className={`btn btn-sm text-nowrap filter-chip ${selected ? 'filter-chip--selected' : ''}`}
                            key={category}
                            onClick={() =>
                                this.props.setSelectedCategories(selectedCategories =>
                                    selected
                                        ? selectedCategories.filter(selectedCategory => selectedCategory !== category)
                                        : [...selectedCategories, category]
                                )
                            }
                        >
                            {category}
                        </button>
                    )
                })}
                <ButtonDropdown isOpen={this.state.open === 'categories'} toggle={this.toggleCategories}>
                    <DropdownToggle caret={true}>Category</DropdownToggle>
                    <DropdownMenu right={true}>
                        {EXTENSION_CATEGORIES.map(category => (
                            <DropdownItem
                                // eslint-disable-next-line react/jsx-no-bind
                                onClick={() => this.props.onQueryChange(extensionsQuery({ category }))}
                                key={category}
                                disabled={this.props.query === extensionsQuery({ category })}
                            >
                                {category}
                            </DropdownItem>
                        ))}
                    </DropdownMenu>
                </ButtonDropdown>{' '}
                <ButtonDropdown isOpen={this.state.open === 'options'} toggle={this.toggleOptions}>
                    <DropdownToggle caret={true}>Options</DropdownToggle>
                    <DropdownMenu right={true}>
                        <DropdownItem
                            // eslint-disable-next-line react/jsx-no-bind
                            onClick={() => this.props.onQueryChange(extensionsQuery({ enabled: true }))}
                            disabled={this.props.query.includes(extensionsQuery({ enabled: true }))}
                        >
                            Show enabled extensions
                        </DropdownItem>
                        <DropdownItem
                            // eslint-disable-next-line react/jsx-no-bind
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
}

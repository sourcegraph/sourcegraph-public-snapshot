import React from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import { EXTENSION_CATEGORIES, ExtensionCategory } from '../../../shared/src/schema/extensionSchema'
import { ExtensionsEnablement } from './ExtensionsList'
import { SettingsCascadeOrError } from '../../../shared/src/settings/settings'

interface Props {
    /** The current extensions registry list query. */
    query: string

    /** Called when the query changes as a result of user interaction with this component. */
    onQueryChange: (query: string) => void

    /**  */
    selectedCategories: ExtensionCategory[]

    /**  */
    setSelectedCategories: React.Dispatch<React.SetStateAction<ExtensionCategory[]>>

    enablementFilter: ExtensionsEnablement

    setEnablementFilter: React.Dispatch<React.SetStateAction<ExtensionsEnablement>>

    /** used to programmatically change reactstrap colors */
    isLightTheme: boolean
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

    private toggleOptions = (): void => this.toggleIsOpen('options')
    private toggleIsOpen = (menu: DropdownMenuID): void =>
        this.setState(previousState => ({ open: previousState.open === menu ? undefined : menu }))

    public render(): JSX.Element | null {
        return (
            <div className="extensions-list__toolbar mb-2">
                <div>
                    {EXTENSION_CATEGORIES.map(category => {
                        const selected = this.props.selectedCategories.includes(category)
                        return (
                            <button
                                type="button"
                                className={`btn btn-sm mr-2 ${selected ? 'btn-secondary' : 'btn-outline-secondary'}`}
                                // style={{
                                //     backgroundColor: !this.props.isLightTheme ? '#151C28' : undefined,
                                //     color: !this.props.isLightTheme ? 'A2B0CD' : undefined,
                                // }}
                                data-test-extension-category={category}
                                key={category}
                                onClick={() =>
                                    this.props.setSelectedCategories(selectedCategories =>
                                        selected
                                            ? selectedCategories.filter(
                                                  selectedCategory => selectedCategory !== category
                                              )
                                            : [...selectedCategories, category]
                                    )
                                }
                            >
                                {category}
                            </button>
                        )
                    })}
                </div>

                <ButtonDropdown isOpen={this.state.open === 'options'} toggle={this.toggleOptions}>
                    <DropdownToggle
                        className="btn-sm"
                        caret={true}
                        color={this.props.isLightTheme ? 'outline-secondary' : 'secondary'}
                        style={{
                            backgroundColor: !this.props.isLightTheme ? '#151C28' : undefined,
                            color: !this.props.isLightTheme ? 'A2B0CD' : undefined,
                        }}
                    >
                        Options
                    </DropdownToggle>
                    <DropdownMenu right={true}>
                        <DropdownItem
                            // eslint-disable-next-line react/jsx-no-bind
                            onClick={() => this.props.setEnablementFilter('all')}
                            disabled={this.props.enablementFilter === 'all'}
                        >
                            Show all
                        </DropdownItem>
                        <DropdownItem
                            // eslint-disable-next-line react/jsx-no-bind
                            onClick={() => this.props.setEnablementFilter('enabled')}
                            disabled={this.props.enablementFilter === 'enabled'}
                        >
                            Show enabled extensions
                        </DropdownItem>
                        <DropdownItem
                            // eslint-disable-next-line react/jsx-no-bind
                            onClick={() => this.props.setEnablementFilter('disabled')}
                            disabled={this.props.enablementFilter === 'disabled'}
                        >
                            Show disabled extensions
                        </DropdownItem>
                    </DropdownMenu>
                </ButtonDropdown>
            </div>
        )
    }
}

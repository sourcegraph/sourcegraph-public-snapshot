import DropdownIcon from 'mdi-react/ChevronDownIcon'
import * as React from 'react'

export interface DropdownProps {
    id: string
    onChange: ()=> void
    required: boolean
    disabled: boolean
    value: string
}

export class Dropdown extends React.PureComponent<DropdownProps, {}> {
    public render(): JSX.Element | null {
        return (
			<div className="dropdown-element">
				<select
					className="form-dropdown"
					id={this.props.id}
					onChange={this.props.onChange}
					required={this.props.required}
					disabled={this.props.disabled}
					value={this.props.value}
				>
					{this.props.children}
				</select>
				<DropdownIcon className="icon-dropdown-chevron" />
			  </div>
        )
    }
}

import DropdownIcon from 'mdi-react/ChevronDownIcon'
import * as React from 'react'

interface Props {
    id: string
    onChange: any
    required: boolean
    disabled: boolean
    value: string
	children: any
}

export const Dropdown: React.FunctionComponent<Props> = (props: Props) => (
	<div className="dropdown-element">
		<select
			className="form-dropdown"
			id={props.id}
			onChange={props.onChange}
			required={props.required}
			disabled={props.disabled}
			value={props.value}
		>
			{props.children}
		</select>
		<DropdownIcon className="icon-dropdown-chevron" />
	  </div>
)

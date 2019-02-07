import DropdownIcon from 'mdi-react/ChevronDownIcon'
import * as React from 'react'

interface Props extends React.DetailedHTMLProps<React.SelectHTMLAttributes<HTMLSelectElement>, HTMLSelectElement> {
    outerClassName?: string
}
/**
 * A drop-in replacement for native `<select>` elements that ensures proper cross-browser styling.
 */
export const Select: React.FunctionComponent<Props> = ({ className = '', ...props }) => (
    <div className={`select ${props.outerClassName}`}>
        {/* tslint:disable-next-line:jsx-ban-elements this is the ONLY allowed instance of <select> */}
        <select {...props} className={`select__picker form-control ${className}`}>
            {props.children}
        </select>
        <DropdownIcon className="select__icon" />
    </div>
)

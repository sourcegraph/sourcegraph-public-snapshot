import {FC, forwardRef, InputHTMLAttributes} from 'react'
import { ComboboxInput } from "./Combobox";

interface MultiCombobox {}

export const MultiCombobox: FC<MultiCombobox> = props => {

    return null
}

interface MultiComboboxInputProps extends InputHTMLAttributes<HTMLInputElement> {}

export const MultiComboboxInput = forwardRef<HTMLInputElement, MultiComboboxInputProps>((props, reference) => {
    const { ...attributes } = props

    return <ComboboxInput ref={reference}/>
})

interface MultiValueInputProps extends InputHTMLAttributes<HTMLInputElement> {}

const MultiValueInput = forwardRef()props => {

    return null
}

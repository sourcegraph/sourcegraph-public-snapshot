import { StylesConfig } from 'react-select'

export const STYLES: StylesConfig = {
    clearIndicator: provided => ({
        ...provided,
        padding: '0.125rem 0',
        borderRadius: 'var(--border-radius)',
        '&:hover': {
            background: 'var(--secondary-3)',
        },
    }),
    control: (provided, state) => ({
        ...provided,
        // Styles here replicate the styles of `wildcard/Select`
        backgroundColor: state.isDisabled ? 'var(--input-disabled-bg)' : 'var(--input-bg)',
        borderColor: state.selectProps.isValid
            ? 'var(--success)'
            : state.selectProps.isValid === false
            ? 'var(--danger)'
            : state.isFocused
            ? state.theme.colors.primary
            : 'var(--input-border-color)',
        boxShadow: state.isFocused
            ? // These are stolen from `wildcard/Input` and `wildcard/Select`, which come from `client/branded/src/global-styles/forms.scss`
              state.selectProps.isValid
                ? 'var(--input-focus-box-shadow-valid)'
                : state.selectProps.isValid === false
                ? 'var(--input-focus-box-shadow-invalid)'
                : 'var(--input-focus-box-shadow)'
            : undefined,
        cursor: 'pointer',
        '&:hover': {
            borderColor: undefined,
        },
    }),
    dropdownIndicator: provided => ({
        ...provided,
        padding: '0.125rem 0',
        borderRadius: 'var(--border-radius)',
        '&:hover': {
            background: 'var(--secondary-3)',
        },
    }),
    indicatorSeparator: (provided, state) => ({
        ...provided,
        backgroundColor: state.hasValue ? 'var(--input-border-color)' : 'transparent',
    }),
    input: provided => ({
        ...provided,
        color: 'var(--input-color)',
        margin: '0 0.125rem',
        padding: 0,
    }),
    menu: provided => ({
        ...provided,
        background: 'var(--dropdown-bg)',
        padding: 0,
        margin: '0.125rem 0 0',
        dropShadow: 'var(--dropdown-shadow)',
        // This is to prevent item edges from sticking out of the rounded dropdown container
        overflow: 'hidden',
    }),
    menuList: provided => ({
        ...provided,
        padding: 0,
    }),
    multiValue: (provided, state) => ({
        display: 'flex',
        maxWidth: '100%',
        alignItems: 'center',
        padding: '0 0 0 0.5rem',
        margin: '0.125rem',
        background: state.isFocused ? 'var(--secondary-3)' : 'var(--secondary)',
        borderStyle: 'solid',
        borderWidth: '1px',
        borderColor: state.isFocused ? 'var(--select-button-border-color)' : 'transparent',
        '&:hover': {
            background: 'var(--secondary-3)',
            borderColor: 'var(--select-button-border-color)',
        },
    }),
    multiValueRemove: (provided, state) => ({
        ...provided,
        padding: '4px',
        marginLeft: '0.25rem',
        borderRadius: '0 var(--border-radius) var(--border-radius) 0',
        background: state.isFocused ? 'var(--secondary-3)' : undefined,
        ':hover': {
            ...provided[':hover'],
            background: 'var(--secondary-3)',
            color: undefined,
        },
    }),
    noOptionsMessage: provided => ({
        ...provided,
        color: 'var(--input-placeholder-color)',
    }),
    option: (provided, state) => ({
        ...provided,
        backgroundColor: state.isSelected
            ? state.isFocused
                ? 'var(--primary-3)'
                : state.theme.colors.primary
            : state.isFocused
            ? 'var(--dropdown-link-hover-bg)'
            : undefined,
        color: state.isSelected ? 'var(--light-text)' : undefined,
        ':hover': {
            cursor: 'pointer',
        },
        ':active': {
            backgroundColor: state.isSelected
                ? state.isFocused
                    ? 'var(--primary-3)'
                    : state.theme.colors.primary
                : state.isFocused
                ? 'var(--dropdown-link-hover-bg)'
                : undefined,
        },
    }),
    placeholder: (provided, state) => ({
        ...provided,
        color: state.isDisabled ? 'var(--gray-06)' : 'var(--input-placeholder-color)',
    }),
    valueContainer: provided => ({
        ...provided,
        padding: '0.125rem 0.125rem 0.125rem 0.75rem',
    }),
}

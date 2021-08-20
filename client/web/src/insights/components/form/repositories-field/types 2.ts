import { InputHTMLAttributes } from 'react'

/**
 * Common props for multi and single repository field with suggestions.
 */
export interface RepositoryFieldProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'value' | 'onChange'> {
    /**
     * String value for the input - repo, repo, ....
     */
    value: string

    /**
     * Change handler runs when user changed input value or picked element
     * from the suggestion panel.
     */
    onChange: (value: string) => void
}

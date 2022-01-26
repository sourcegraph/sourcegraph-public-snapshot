import classNames from 'classnames'
import React, { InputHTMLAttributes } from 'react'

import styles from './FilterInput.module.scss'

type FilterInputProps = InputHTMLAttributes<HTMLInputElement>

export const FilterInput: React.FunctionComponent<FilterInputProps> = ({ children, className, ...rest }) => (
    <input className={classNames(className, styles.filterInput)} {...rest} />
)

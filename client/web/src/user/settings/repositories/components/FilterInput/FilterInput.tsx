import { InputHTMLAttributes } from 'react'
import * as React from 'react'

import classNames from 'classnames'

import styles from './FilterInput.module.scss'

type FilterInputProps = InputHTMLAttributes<HTMLInputElement>

export const FilterInput: React.FunctionComponent<FilterInputProps> = ({ children, className, ...rest }) => (
    <input className={classNames(className, styles.filterInput)} {...rest} />
)

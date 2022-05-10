import React, { InputHTMLAttributes } from 'react'

import classNames from 'classnames'

import styles from './FilterInput.module.scss'

type FilterInputProps = InputHTMLAttributes<HTMLInputElement>

export const FilterInput: React.FunctionComponent<React.PropsWithChildren<FilterInputProps>> = ({
    children,
    className,
    ...rest
}) => <input className={classNames(className, styles.filterInput)} {...rest} />

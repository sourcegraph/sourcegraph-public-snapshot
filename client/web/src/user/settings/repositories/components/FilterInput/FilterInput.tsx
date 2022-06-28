import React, { InputHTMLAttributes } from 'react'

import classNames from 'classnames'

import { Input } from '@sourcegraph/wildcard'

import styles from './FilterInput.module.scss'

type FilterInputProps = InputHTMLAttributes<HTMLInputElement>

export const FilterInput: React.FunctionComponent<React.PropsWithChildren<FilterInputProps>> = ({
    children,
    className,
    ...rest
}) => <Input className={classNames(className, styles.filterInput)} {...rest} />

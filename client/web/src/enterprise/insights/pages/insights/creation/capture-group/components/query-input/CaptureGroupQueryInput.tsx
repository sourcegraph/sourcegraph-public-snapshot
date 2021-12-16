import classNames from 'classnames'
import RegexIcon from 'mdi-react/RegexIcon'
import React from 'react'

import { MonacoField, MonacoFieldProps } from '../../../../../../components/form/monaco-field/MonacoField'

import styles from './CaptureGroupQueryInput.module.scss'

export const CaptureGroupQueryInput: React.FunctionComponent<MonacoFieldProps> = props => (
    <div className={styles.root}>
        <MonacoField {...props} className={classNames(styles.input, props.className)} />

        <button type="button" className={classNames('btn btn-icon', styles.regexButton)} disabled={true}>
            <RegexIcon
                size={24}
                data-tooltip="Regular expression is the only pattern type usable with capture groups and it’s enabled by default for this search input’"
            />
        </button>
    </div>
)

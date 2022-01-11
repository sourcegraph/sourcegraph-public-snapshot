import classNames from 'classnames'
import RegexIcon from 'mdi-react/RegexIcon'
import React, { forwardRef } from 'react'

import { Button } from '@sourcegraph/wildcard'

import { MonacoField, MonacoFieldProps } from '../../../../../../components/form/monaco-field/MonacoField'

import styles from './CaptureGroupQueryInput.module.scss'

export const CaptureGroupQueryInput = forwardRef<HTMLInputElement, MonacoFieldProps>((props, reference) => (
    <div className={styles.root}>
        <MonacoField {...props} ref={reference} className={classNames(styles.input, props.className)} />

        <Button className={classNames('btn-icon', styles.regexButton)} disabled={true}>
            <RegexIcon
                size={24}
                data-tooltip="Regular expression is the only pattern type usable with capture groups and it’s enabled by default for this search input."
            />
        </Button>
    </div>
))

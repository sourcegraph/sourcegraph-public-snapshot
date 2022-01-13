import classNames from 'classnames'
import RegexIcon from 'mdi-react/RegexIcon'
import React, { forwardRef } from 'react'
import { Link } from 'react-router-dom'

import { Button } from '@sourcegraph/wildcard'

import * as Monaco from '../../../../../../components/form/monaco-field/MonacoField'
import type { MonacoFieldProps } from '../../../../../../components/form/monaco-field/MonacoField'

import styles from './CaptureGroupQueryInput.module.scss'

export const CaptureGroupQueryInput = forwardRef<HTMLInputElement, MonacoFieldProps>((props, reference) => (
    <div className={styles.root}>
        <Monaco.Root className={classNames(props.className, styles.inputWrapper)}>
            <Monaco.Field {...props} ref={reference} className={props.className} />

            <Button className={classNames('btn-icon', styles.regexButton)} disabled={true}>
                <RegexIcon
                    size={21}
                    data-tooltip="Regular expression is the only pattern type usable with capture groups and it’s enabled by default for this search input."
                />
            </Button>
        </Monaco.Root>

        <Button className={styles.previewButton} to="preview" variant="secondary" as={Link}>
            Hello
        </Button>
    </div>
))

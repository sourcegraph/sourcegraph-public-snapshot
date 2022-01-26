import classNames from 'classnames'
import RegexIcon from 'mdi-react/RegexIcon'
import React, { forwardRef } from 'react'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { Button } from '@sourcegraph/wildcard'

import type { MonacoFieldProps } from '../../../../../../components/form/monaco-field'
import { InsightQueryInput } from '../../../../../../components/form/query-input/InsightQueryInput'

import styles from './CaptureGroupQueryInput.module.scss'

export const CaptureGroupQueryInput = forwardRef<HTMLInputElement, MonacoFieldProps>((props, reference) => (
    <InsightQueryInput {...props} ref={reference} patternType={SearchPatternType.regexp}>
        <Button className={classNames('btn-icon', styles.regexButton)} disabled={true}>
            <RegexIcon
                size={21}
                data-tooltip="Regular expression is the only pattern type usable with capture groups and it’s enabled by default for this search input."
            />
        </Button>
    </InsightQueryInput>
))

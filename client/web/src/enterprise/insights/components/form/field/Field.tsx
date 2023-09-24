import React, { createContext, type InputHTMLAttributes, useContext } from 'react'

import classNames from 'classnames'
import { noop } from 'lodash'

import { LazyQueryInput } from '@sourcegraph/branded'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import type { QueryState } from '@sourcegraph/shared/src/search'

import styles from './Field.module.scss'

interface Context {
    renderedWithinFocusContainer: boolean
}

const FieldContext = createContext<Context>({ renderedWithinFocusContainer: false })

const CONTAINER_MARK = { renderedWithinFocusContainer: true }

export const FocusContainer: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
    children,
}) => (
    <FieldContext.Provider value={CONTAINER_MARK}>
        <div
            className={classNames(
                'form-control',
                'with-invalid-icon',
                styles.container,
                styles.focusContainer,
                className
            )}
        >
            {children}
        </div>
    </FieldContext.Provider>
)

export interface FieldProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'value' | 'onChange' | 'onBlur'> {
    queryState: QueryState
    patternType?: SearchPatternType
    onBlur?: () => void
    onChange?: (value: QueryState) => void
}

export const Field: React.FunctionComponent<FieldProps> = ({
    queryState,
    className,
    onChange = noop,
    onBlur = noop,
    disabled,
    autoFocus,
    placeholder,
    patternType = SearchPatternType.regexp,
    'aria-labelledby': ariaLabelledby,
    'aria-invalid': ariaInvalid,
    'aria-busy': ariaBusy,
    tabIndex = 0,
}) => {
    const { renderedWithinFocusContainer } = useContext(FieldContext)

    return (
        <LazyQueryInput
            ariaLabelledby={ariaLabelledby}
            ariaInvalid={ariaInvalid?.toString()}
            ariaBusy={ariaBusy?.toString()}
            queryState={queryState}
            isSourcegraphDotCom={false}
            preventNewLine={false}
            interpretComments={true}
            onChange={onChange}
            patternType={patternType}
            caseSensitive={false}
            placeholder={placeholder}
            className={classNames(className, styles.field, 'form-control', 'with-invalid-icon', {
                [styles.focusContainer]: !renderedWithinFocusContainer,
                [styles.fieldWithoutFieldStyles]: renderedWithinFocusContainer,
            })}
            readOnly={disabled}
            autoFocus={autoFocus}
            onBlur={onBlur}
            tabIndex={tabIndex}
        />
    )
}

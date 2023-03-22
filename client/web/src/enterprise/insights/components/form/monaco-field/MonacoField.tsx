import { createContext, forwardRef, InputHTMLAttributes, useContext, useImperativeHandle, useMemo } from 'react'

import classNames from 'classnames'
import { noop } from 'lodash'

import { LazyQueryInput } from '@sourcegraph/branded'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { QueryState } from '@sourcegraph/shared/src/search'
import { ForwardReferenceComponent } from '@sourcegraph/wildcard'

import styles from './MonacoField.module.scss'

interface Context {
    renderedWithinFocusContainer: boolean
}

const MonacoFieldContext = createContext<Context>({ renderedWithinFocusContainer: false })

const MONACO_CONTAINER_MARK = { renderedWithinFocusContainer: true }

export const MonacoFocusContainer = forwardRef((props, reference) => {
    const { as: Component = 'div', className, children, ...otherProps } = props

    return (
        <MonacoFieldContext.Provider value={MONACO_CONTAINER_MARK}>
            <Component
                {...otherProps}
                className={classNames(
                    'form-control',
                    'with-invalid-icon',
                    styles.container,
                    styles.focusContainer,
                    className
                )}
            >
                {children}
            </Component>
        </MonacoFieldContext.Provider>
    )
}) as ForwardReferenceComponent<'div'>

export interface MonacoFieldProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'value' | 'onChange' | 'onBlur'> {
    queryState: QueryState
    patternType?: SearchPatternType
    onBlur?: () => void
    onChange?: (value: QueryState) => void
}

export const MonacoField = forwardRef<HTMLInputElement, MonacoFieldProps>((props, reference) => {
    const {
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
    } = props

    const { renderedWithinFocusContainer } = useContext(MonacoFieldContext)

    // Monaco doesn't have any native input elements, so we mock
    // ref here to avoid React warnings in console about zero usage of
    // element ref with forward ref call.
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    useImperativeHandle(reference, () => null)

    const monacoOptions = useMemo(() => ({ readOnly: disabled }), [disabled])

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
            className={classNames(className, styles.monacoField, 'form-control', 'with-invalid-icon', {
                [styles.focusContainer]: !renderedWithinFocusContainer,
                [styles.monacoFieldWithoutFieldStyles]: renderedWithinFocusContainer,
            })}
            editorOptions={monacoOptions}
            autoFocus={autoFocus}
            onBlur={onBlur}
            applySuggestionsOnEnter={false}
            tabIndex={tabIndex}
        />
    )
})

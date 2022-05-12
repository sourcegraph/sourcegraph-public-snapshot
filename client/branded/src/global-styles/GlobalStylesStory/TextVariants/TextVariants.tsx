import React, { ElementType } from 'react'

import classNames from 'classnames'

import styles from './TextVariants.module.scss'

const SIZE_VARIANTS = ['Base', 'Small'] as const
const WEIGHT_VARIANTS = ['Regular', 'Medium', 'Strong'] as const

type TextWeight = typeof WEIGHT_VARIANTS[number]
type TextSize = typeof SIZE_VARIANTS[number]

interface TextLabelProps {
    size: TextSize
    weight: TextWeight
    name: string
    className?: string
}

const TextLabel: React.FunctionComponent<React.PropsWithChildren<TextLabelProps>> = props => {
    const { size, weight, name, className } = props
    const label = `This is ${name} / ${size} / ${weight}`

    if (weight === 'Strong') {
        return <strong className={className}>{label}</strong>
    }

    if (weight === 'Medium') {
        return <span className={classNames('font-weight-medium', className)}>{label}</span>
    }

    return <>{label}</>
}

interface TextVariantsProps {
    component: ElementType
    name: string
    weights?: TextWeight[]
    className?: string
}

const TextVariations: React.FunctionComponent<React.PropsWithChildren<TextVariantsProps>> = props => {
    const { component: Component, name, weights = ['Regular'], className } = props

    const textVariations = SIZE_VARIANTS.flatMap(size =>
        weights.map(weight => {
            const SizeWrapper = size === 'Small' ? 'small' : React.Fragment

            return (
                <Component key={`${size}/${weight}`} className={classNames(styles.textVariant, className)}>
                    <SizeWrapper>
                        <TextLabel size={size} weight={weight} name={name} />
                    </SizeWrapper>
                </Component>
            )
        })
    )

    return <>{textVariations}</>
}

export const TextVariants: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <table className="table">
        <tbody>
            <tr>
                <td>Body Text</td>
                <td>
                    {WEIGHT_VARIANTS.map(weight => (
                        <p key={`Base/${weight}`} className={styles.textVariant}>
                            <TextLabel size="Base" name="Body" weight={weight} />
                        </p>
                    ))}
                    {WEIGHT_VARIANTS.map(weight => (
                        <p key={`Small/${weight}`} className={styles.textVariant}>
                            <small>
                                <TextLabel
                                    size="Small"
                                    name="Body"
                                    weight={weight}
                                    className={classNames({ 'font-weight-bold': weight === 'Strong' })}
                                />
                            </small>
                        </p>
                    ))}
                </td>
            </tr>
            <tr>
                <td>
                    <code>{'<label>'}</code>
                </td>
                <td>
                    <TextVariations component="label" name="Label" />
                    <TextVariations component="label" name="Label" className="text-uppercase" />
                </td>
            </tr>
            <tr>
                <td>
                    <code>{'<input class="form-control">'}</code>
                </td>
                <td>
                    <span className={classNames('form-control', styles.inputVariant, styles.textVariant)}>
                        <TextLabel size="Base" weight="Regular" name="Input" />
                    </span>
                    <span
                        className={classNames('form-control form-control-sm', styles.inputVariant, styles.textVariant)}
                    >
                        <TextLabel size="Small" weight="Regular" name="Input" />
                    </span>
                </td>
            </tr>
            <tr>
                <td>
                    <code>{'<code>'}</code>
                </td>
                <td>
                    <TextVariations component="code" name="Code" weights={['Regular', 'Strong']} />
                </td>
            </tr>
        </tbody>
    </table>
)

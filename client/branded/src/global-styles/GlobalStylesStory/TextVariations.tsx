import classNames from 'classnames'
import { flatten } from 'lodash'
import React, { ElementType } from 'react'

import styles from './TextVariations.module.scss'

const SIZE_VARIANTS = ['Base', 'Small'] as const
const WEIGHT_VARIANTS = ['Regular', 'Medium', 'Strong'] as const

type TextWeight = typeof WEIGHT_VARIANTS[number]
type TextSize = typeof SIZE_VARIANTS[number]

interface TextLabelProps {
    size: TextSize
    weight: TextWeight
    name: string
}

const TextLabel: React.FunctionComponent<TextLabelProps> = props => {
    const { size, weight, name } = props
    const label = `This is ${name} / ${size} / ${weight}`

    if (weight === 'Strong') {
        return <strong>{label}</strong>
    }

    if (weight === 'Medium') {
        return <span className="font-weight-semibold">{label}</span>
    }

    return <>{label}</>
}

interface TextVariantsProps {
    component: ElementType
    name: string
    weights?: TextWeight[]
    className?: string
}

const TextVariants: React.FunctionComponent<TextVariantsProps> = props => {
    const { component: Component, name, weights = ['Regular'], className } = props

    const textVariants = SIZE_VARIANTS.map(size =>
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

    return <>{flatten(textVariants)}</>
}

export const TextVariations: React.FunctionComponent = () => (
    <table className="table">
        <tbody>
            <tr>
                <td>
                    <code>{'<p>'}</code>
                </td>
                <td>
                    <TextVariants component="p" name="Body" weights={['Regular', 'Medium', 'Strong']} />
                </td>
            </tr>
            <tr>
                <td>
                    <code>{'<label>'}</code>
                </td>
                <td>
                    <TextVariants component="label" name="Label" />
                </td>
            </tr>
            <tr>
                <td>
                    <code>{'<span class="form-control">'}</code>
                </td>
                <td>
                    <TextVariants
                        component="span"
                        name="Input"
                        className={classNames('form-control', styles.inputVariant)}
                    />
                </td>
            </tr>
            <tr>
                <td>
                    <code>{'<code>'}</code>
                </td>
                <td>
                    <TextVariants component="code" name="Code" weights={['Regular', 'Strong']} />
                </td>
            </tr>
        </tbody>
    </table>
)

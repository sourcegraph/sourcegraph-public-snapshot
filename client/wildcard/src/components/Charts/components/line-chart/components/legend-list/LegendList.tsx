import { createContext, type FC, forwardRef, type ReactNode, useContext } from 'react'

import classNames from 'classnames'

import type { ForwardReferenceComponent } from '../../../../../../types'

import styles from './LegendList.module.scss'

export const LegendList = forwardRef(function LegendList(props, ref) {
    const { as: Component = 'ul', 'aria-label': ariaLabel = 'Chart legend', className, ...attributes } = props

    return (
        <Component
            {...attributes}
            ref={ref}
            aria-label={ariaLabel}
            className={classNames(styles.legendList, className)}
        />
    )
}) as ForwardReferenceComponent<'ul'>

interface LegendItemContextData {
    active: boolean
}

const LegendItemContext = createContext<LegendItemContextData>({ active: true })

type LegendItemNameProps = { name: string; children?: undefined } | { name?: undefined; children: ReactNode }
type LegendItemProps = LegendItemNameProps & {
    active?: boolean
}

export const LegendItem = forwardRef(function LegendItem(props, ref) {
    const {
        name,
        children,
        active = true,
        as: Component = 'li',
        color = 'var(--gray-07)',
        className,
        ...attributes
    } = props

    return (
        <LegendItemContext.Provider value={{ active }}>
            <Component
                ref={ref}
                {...attributes}
                className={classNames(styles.legendItem, className, { 'text-muted': !active })}
            >
                {name ? (
                    <>
                        <LegendItemPoint color={color} active={active} />
                        {name}
                    </>
                ) : (
                    children
                )}
            </Component>
        </LegendItemContext.Provider>
    )
}) as ForwardReferenceComponent<'li', LegendItemProps>

interface LegendItemPointProps {
    color?: string
    active?: boolean
}

export const LegendItemPoint: FC<LegendItemPointProps> = props => {
    const { color = 'var(--gray-07)', active: propActive } = props
    const { active: contextActive } = useContext(LegendItemContext)

    const active = propActive ?? contextActive

    return (
        <div className={styles.legendMarkContainer}>
            <span
                aria-hidden={true}
                /* eslint-disable-next-line react/forbid-dom-props */
                style={{ backgroundColor: active ? color : 'var(--icon-muted)' }}
                className={styles.legendMark}
            />
        </div>
    )
}

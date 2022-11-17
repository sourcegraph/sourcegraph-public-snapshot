import { ReactElement } from 'react'

import { H3, Text } from '../../../../Typography'
import { DEFAULT_FALLBACK_COLOR } from '../../../constants'
import { TooltipList, TooltipListItem } from '../../../core'
import { Category } from '../utils/get-grouped-categories'

import styles from './TooltipContent.module.scss'

interface BarTooltipContentProps<Datum> {
    category: Category<Datum>
    activeBar: Datum
    getDatumName: (datum: Datum) => string
    getDatumValue: (datum: Datum) => number
    getDatumHover?: (datum: Datum) => string
    getDatumColor: (datum: Datum) => string | undefined
}

export function BarTooltipContent<Datum>(props: BarTooltipContentProps<Datum>): ReactElement {
    const { category, getDatumName, getDatumHover, getDatumValue, getDatumColor, activeBar } = props
    const getName = getDatumHover ?? getDatumName
    const activeDatumHover = getName(activeBar)

    // Handle a special case when we don't have any multiple datum per group
    if (category.data.length === 1) {
        const datum = category.data[0]
        const name = getName(datum)
        const value = getDatumValue(datum)

        return (
            <Text className={styles.oneLineTooltip}>
                <span>{name}</span>
                <span>{value}</span>
            </Text>
        )
    }

    // Handle a special case when we don't have any multiple datum per group
    if (category.data.length === 1) {
        const datum = category.data[0]
        const name = getName(datum)
        const value = getDatumValue(datum)

        return (
            <Text className={styles.oneLineTooltip}>
                <span>{name}</span>
                <span>{value}</span>
            </Text>
        )
    }

    return (
        <>
            <H3>{category.id}</H3>
            <TooltipList>
                {category.data.map(item => {
                    const hover = getName(item)

                    return (
                        <TooltipListItem
                            key={hover}
                            isActive={activeDatumHover === hover}
                            name={hover}
                            value={getDatumValue(item)}
                            color={getDatumColor(item) ?? DEFAULT_FALLBACK_COLOR}
                        />
                    )
                })}
            </TooltipList>
        </>
    )
}

import { ReactElement } from 'react'

import { H3 } from '../../../../Typography'
import { DEFAULT_FALLBACK_COLOR } from '../../../constants'
import { TooltipList, TooltipListItem } from '../../../core'
import { Category } from '../utils/get-grouped-categories'

interface BarTooltipContentProps<Datum> {
    category: Category<Datum>
    activeBar: Datum
    getDatumName: (datum: Datum) => string
    getDatumValue: (datum: Datum) => number
    getDatumColor: (datum: Datum) => string | undefined
}

export function BarTooltipContent<Datum>(props: BarTooltipContentProps<Datum>): ReactElement {
    const { category, getDatumName, getDatumValue, getDatumColor, activeBar } = props
    const activeDatumName = getDatumName(activeBar)

    return (
        <>
            {category.data.length > 1} <H3>{category.id}</H3>
            <TooltipList>
                {category.data.map(item => {
                    const name = getDatumName(item)

                    return (
                        <TooltipListItem
                            key={name}
                            isActive={activeDatumName === name}
                            name={name}
                            value={getDatumValue(item)}
                            color={getDatumColor(item) ?? DEFAULT_FALLBACK_COLOR}
                        />
                    )
                })}
            </TooltipList>
        </>
    )
}

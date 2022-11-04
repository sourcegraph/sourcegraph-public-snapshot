import { groupBy } from 'lodash'

interface GetCategoriesInput<Datum> {
    data: Datum[]
    stacked: boolean
    sortByValue?: boolean
    getDatumName: (datum: Datum) => string
    getDatumValue: (datum: Datum) => number
    getCategory: (datum: Datum) => string | undefined
}

export interface Category<Datum> {
    id: string
    data: Datum[]
    maxValue: number
    stacked: boolean
}

export function getGroupedCategories<Datum>(input: GetCategoriesInput<Datum>): Category<Datum>[] {
    const { data, stacked, sortByValue, getCategory, getDatumName, getDatumValue } = input

    const categories = groupBy<Datum>(data, datum => getCategory(datum) ?? getDatumName(datum))

    return Object.keys(categories).reduce<Category<Datum>[]>((store, key) => {
        const data = categories[key]

        const maxValue = stacked
            ? data.reduce((sum, datum) => sum + getDatumValue(datum), 0)
            : Math.max(...data.map(getDatumValue))

        store.push({
            id: key,
            data,
            maxValue,
            stacked,
        })

        if (sortByValue) {
            return store.sort((a, b) => b.maxValue - a.maxValue)
        }

        return store
    }, [])
}

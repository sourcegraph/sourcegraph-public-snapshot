import { groupBy } from 'lodash'

interface GetCategoriesInput<Datum> {
    data: Datum[]
    getDatumName: (datum: Datum) => string
    getCategory: (datum: Datum) => string | undefined
}

export interface Category<Datum> {
    id: string
    data: Datum[]
}

export function getGroupedCategories<Datum>(input: GetCategoriesInput<Datum>): Category<Datum>[] {
    const { data, getCategory, getDatumName } = input

    const categories = groupBy<Datum>(data, datum => getCategory(datum) ?? getDatumName(datum))

    return Object.keys(categories).reduce<Category<Datum>[]>((store, key) => {
        store.push({ id: key, data: categories[key] })

        return store
    }, [])
}


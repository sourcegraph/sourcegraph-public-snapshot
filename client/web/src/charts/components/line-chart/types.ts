export interface Point<D> {
    id: string
    seriesKey: string
    value: number
    color: string
    x: number
    y: number
    datum: D
    originalDatum: D
    linkUrl?: string
}

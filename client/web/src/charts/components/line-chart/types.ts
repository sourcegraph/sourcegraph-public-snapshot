export interface Point<D> {
    id: string
    seriesKey: string
    value: number
    time: Date
    x: number
    y: number
    datum: D
    linkUrl?: string
    color: string
}

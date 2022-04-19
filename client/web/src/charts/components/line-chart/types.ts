export interface Point<D> {
    id: string
    seriesId: string | number
    value: number
    time: Date
    x: number
    y: number
    datum: D
    linkUrl?: string
    color: string
}

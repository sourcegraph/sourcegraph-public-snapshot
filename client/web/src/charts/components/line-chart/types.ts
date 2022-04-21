export interface Point<D> {
    id: string
    seriesId: string
    value: number
    time: Date
    x: number
    y: number
    linkUrl?: string
    color: string
}

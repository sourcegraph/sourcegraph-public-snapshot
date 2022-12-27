export interface SearchValue {
    text: string
    // Score for previously visited results. History ranking is a callback
    // to allow the
    //
    historyRanking?: () => number | undefined
    ranking?: number
    url?: string
    icon?: JSX.Element
    onClick?: () => void
}

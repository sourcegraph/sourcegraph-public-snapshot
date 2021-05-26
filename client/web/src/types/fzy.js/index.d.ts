declare module 'fzy.js' {
    function score(query: string, value: string): number
    function positions(query: string, value: string): number[]
}

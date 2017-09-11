export interface Reference {
    range: {
        start: {
            character: number;
            line: number;
        };
        end: {
            character: number;
            line: number;
        };
    }
    uri: string
    repoURI: string
}

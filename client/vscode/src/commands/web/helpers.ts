import { SourcegraphUri } from '../../file-system/SourcegraphUri'

export function decodeUri(uri: string): string {
    return SourcegraphUri.parse(uri).uri
}

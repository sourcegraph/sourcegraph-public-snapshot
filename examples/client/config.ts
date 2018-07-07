import { DocumentSelector } from '../../src/types/documents'

export default require('./package.json').cxp as {
    url: string
    accessToken: string
    root: string
    documentSelector: DocumentSelector
    initializationOptions: {
        mode: string
        settings: { merged: any }
    }
}

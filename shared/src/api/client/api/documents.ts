import { ProxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import { Subscription, Unsubscribable } from 'rxjs'
import { getModeFromPath } from '../../../languages'
import { FileSystemService } from '../services/fileSystemService'
import { ModelService, TextModel } from '../services/modelService'

export interface ClientDocumentsAPI extends ProxyValue {
    $openTextDocument(uri: string): Promise<TextModel>
}

export class ClientDocuments implements ClientDocumentsAPI, Unsubscribable {
    public readonly [proxyValueSymbol] = true

    private subscriptions = new Subscription()

    constructor(private fileSystemService: Pick<FileSystemService, 'readFile'>, private modelService: ModelService) {}

    public async $openTextDocument(uri: string): Promise<TextModel> {
        const model: TextModel = {
            uri,
            languageId: getModeFromPath(uri),
            text: await this.fileSystemService.readFile(new URL(uri)),
        }
        if (!this.modelService.hasModel(uri)) {
            this.modelService.addModel(model)
        }
        return model
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}

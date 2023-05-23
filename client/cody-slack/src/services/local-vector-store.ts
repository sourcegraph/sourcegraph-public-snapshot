import { OpenAIEmbeddings } from 'langchain/embeddings/openai'
import { MarkdownTextSplitter } from 'langchain/text_splitter'
import { HNSWLib } from 'langchain/vectorstores/hnswlib'

import { fetchFileContent } from './github-client'

async function getDocuments() {
    const codyNotice = await fetchFileContent({
        owner: 'sourcegraph',
        repo: 'about',
        path: 'content/terms/cody-notice.md',
    })

    if (!codyNotice) {
        return []
    }

    const { content, url } = codyNotice
    const splitter = new MarkdownTextSplitter()
    const documents = await splitter.createDocuments([content])

    documents.map((document, index) => {
        document.metadata = {
            fileName: url,
            hnswLabel: index,
        }

        return document
    })

    return documents
}

const VECTOR_UPDATE_TIMEOUT = 12 * 60 * 60 * 1000

function scheduleVectorUpdate(vectorStore: HNSWLib, timeout: number) {
    setTimeout(async () => {
        try {
            vectorStore._index = undefined
            vectorStore.docstore._docs.clear()

            const documents = await getDocuments()
            await vectorStore.addDocuments(documents)
        } catch (error) {
            console.error('Failed to update vectors', error)
        } finally {
            scheduleVectorUpdate(vectorStore, timeout)
        }
    }, timeout)
}

export async function getVectorStore() {
    const documents = await getDocuments()

    const embeddings = new OpenAIEmbeddings()
    const vectorStore = await HNSWLib.fromDocuments(documents, embeddings)

    scheduleVectorUpdate(vectorStore, VECTOR_UPDATE_TIMEOUT)

    return vectorStore
}

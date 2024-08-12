import { type FC, memo, useMemo } from 'react'

import { useLocation } from 'react-router-dom'

import { CodyWebPanelProvider, type InitialContext } from '@sourcegraph/cody-web'
import { SourcegraphURL } from '@sourcegraph/common'

import { getTelemetrySourceClient } from '../../telemetry'
import { ChatUi } from '../components/ChatUi'

interface Repository {
    id: string
    name: string
}

interface CodySidebarWebChatProps {
    filePath?: string
    repository: Repository
}

export const CodySidebarWebChat: FC<CodySidebarWebChatProps> = memo(function CodyWebChat(props) {
    const { filePath, repository } = props

    const location = useLocation()

    const contextInfo = useMemo<InitialContext>(() => {
        const lineOrPosition = SourcegraphURL.from(location).lineRange
        const hasFileRangeSelection = lineOrPosition.line

        return {
            repositories: [repository],
            fileURL: filePath ? `/${filePath}` : undefined,
            // Line range - 1 because of Cody Web initial context file range bug
            fileRange: hasFileRangeSelection
                ? {
                      startLine: lineOrPosition.line - 1,
                      endLine: lineOrPosition.endLine ? lineOrPosition.endLine - 1 : lineOrPosition.line - 1,
                  }
                : undefined,
        }
    }, [repository, filePath, location])

    return (
        <CodyWebPanelProvider
            accessToken=""
            initialContext={contextInfo}
            serverEndpoint={window.location.origin}
            customHeaders={window.context.xhrHeaders}
            telemetryClientName={getTelemetrySourceClient()}
        >
            <ChatUi />
        </CodyWebPanelProvider>
    )
})

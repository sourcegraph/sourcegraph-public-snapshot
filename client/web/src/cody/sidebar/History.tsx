import { Button, Text } from '@sourcegraph/wildcard'

import { HistoryList } from '../components/HistoryList'
import { useChatStoreState } from '../stores/chat'

interface HistoryProps {
    closeHistory: () => void
}

export const History: React.FunctionComponent<HistoryProps> = ({ closeHistory }) => {
    const { transcriptHistory, clearHistory } = useChatStoreState()

    return (
        <>
            <Text className="p-2 pb-0" as="h3">
                Chat History
            </Text>
            <HistoryList onSelect={closeHistory} />
            {transcriptHistory.length > 0 && (
                <div className="text-center">
                    <Button
                        variant="secondary"
                        outline={true}
                        onClick={() => {
                            closeHistory()
                            clearHistory()
                        }}
                    >
                        Clear History
                    </Button>
                </div>
            )}
        </>
    )
}

import React from 'react'

const BlameContext = React.createContext<{
    isBlameVisible: boolean
    setIsBlameVisible: React.Dispatch<React.SetStateAction<boolean>>
} | null>(null)

export const BlameContextProvider: React.FC<React.PropsWithChildren<{}>> = ({ children }) => {
    const [isBlameVisible, setIsBlameVisible] = React.useState<boolean>(false)

    return <BlameContext.Provider value={{ isBlameVisible, setIsBlameVisible }}>{children}</BlameContext.Provider>
}

export const useBlameVisibility = (): [boolean, React.Dispatch<React.SetStateAction<boolean>>] => {
    const context = React.useContext(BlameContext)

    if (context === null) {
        throw new Error('useBlameVisibility must be used within BlameContextProvider')
    }

    return [context.isBlameVisible, context.setIsBlameVisible]
}

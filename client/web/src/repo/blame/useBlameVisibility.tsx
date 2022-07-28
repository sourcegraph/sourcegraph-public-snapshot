import React from 'react'

const BlameContext = React.createContext<{
    isBlameVisible: boolean
    setIsBlameVisible: React.Dispatch<React.SetStateAction<boolean>>
}>({
    isBlameVisible: false,
    setIsBlameVisible: () => {},
})

export const BlameContextProvider: React.FC<React.PropsWithChildren<{}>> = ({ children }) => {
    const [isBlameVisible, setIsBlameVisible] = React.useState<boolean>(false)

    return <BlameContext.Provider value={{ isBlameVisible, setIsBlameVisible }}>{children}</BlameContext.Provider>
}

export const useBlameVisibility = (): [boolean, React.Dispatch<React.SetStateAction<boolean>>] => {
    const { isBlameVisible, setIsBlameVisible } = React.useContext(BlameContext)

    return [isBlameVisible, setIsBlameVisible]
}

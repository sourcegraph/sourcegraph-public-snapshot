import { FunctionComponent } from 'react'

export const LoadingPage: FunctionComponent = () => {
    return (
        <main className="flex justify-center items-center h-screen">
            <div className="spinner w-8 h-8 border-4 border-blue-700 border-t-transparent rounded-full animate-spin"></div>
        </main>
    )
}

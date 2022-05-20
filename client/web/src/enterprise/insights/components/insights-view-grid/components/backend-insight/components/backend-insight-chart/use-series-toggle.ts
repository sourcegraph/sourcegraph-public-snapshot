interface UseSeriesToggleReturn {
    toggle: (id: string | number) => void
}

export const useSeriesToggle = (): UseSeriesToggleReturn => ({
    toggle: id => console.log('🚀 ~ useSeriesToggle', id),
})

export function isMobile(): boolean {
    return (
        typeof window !== 'undefined' &&
        window.navigator.userAgent.match(/Android|webOS|iPhone|iPad|iPod|BlackBerry|Windows Phone/i) !== null
    )
}

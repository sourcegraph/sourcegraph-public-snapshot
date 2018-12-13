export default (prefix: string) => {
    let nextId = 0
    return () => `${prefix}-${nextId++}`
}

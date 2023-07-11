const apiKey = process.env.API_NINJA_API_KEY!

export const fetchAPINinjas = (url: string): Promise<any> =>
    fetch(url, {
        method: 'GET',
        headers: {
            'X-API-Key': apiKey,
        },
    })

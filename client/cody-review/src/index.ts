const sleep = (msecs: number): Promise<void> => new Promise(resolve => setTimeout(resolve, msecs))

async function review(): Promise<void> {
    console.log('Running review!')
    await sleep(5000)
    console.log('Finished review!')
}

// eslint-disable-next-line @typescript-eslint/no-floating-promises
review()

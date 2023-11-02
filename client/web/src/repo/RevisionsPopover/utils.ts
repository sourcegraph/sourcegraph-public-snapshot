export const getBatchCount = (height: number): number => {
    switch (true) {
        case height < 500: {
            return 5
        }
        case height < 1000: {
            return 10
        }
        case height < 1500: {
            return 15
        }
        case height < 2000: {
            return 25
        }
        default: {
            return 30
        }
    }
}

import { Agent } from './agent'

process.stderr.write('Starting Cody Agent...\n')

const agent = new Agent()

console.log = console.error

process.stdin.pipe(agent.messageDecoder)
agent.messageEncoder.pipe(process.stdout)

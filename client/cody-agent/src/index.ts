import { Agent } from './agent'

process.stderr.write('Starting agent...\n')

const agent = new Agent()

process.stdin.pipe(agent.messageDecoder)
agent.messageEncoder.pipe(process.stdout)

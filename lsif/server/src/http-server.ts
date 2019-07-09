import express from 'express'
import asyncHandler from 'express-async-handler'
import bodyParser from 'body-parser'
import cors from 'cors'
import { JsonDatabase } from './json';
import { Database } from './database';
import * as fs from 'fs'
import * as path from 'path'
import fileUpload from 'express-fileupload';
import LRU from 'lru-cache';

// TODO add docstrings

const storageRoot = 'lsif'
// The storage size can exceed this max if a single LSIF file is larger than
// this, otherwise disk usage will be kept under this.
//
// TODO make this configurable
const softMaxStorageSize = 100 * 1024 * 1024 * 1024

// TODO make this configurable
const maxFileSize = 100 * 1024 * 1024
// disk usage to memory usage is a ratio of roughly 1:3 (based on sourcegraph/codeintellify)
// TODO make this configurable
const softMaxDBSizeLoadedIntoMemory = 100 * 1024 * 1024

interface Repository {
	repository: string
}
interface Commit {
	commit: string
}
interface RepositoryCommit extends Repository, Commit { }

if (!fs.existsSync(storageRoot)) {
	fs.mkdirSync(storageRoot);
}

function enforceMaxDiskUsage({ flatDirectory, max, onBeforeDelete }: { flatDirectory: string, max: number, onBeforeDelete: (filePath: string) => void }): string[] {
	const files = fs.readdirSync(flatDirectory).map(f => ({ path: path.join(flatDirectory, f), stat: fs.statSync(path.join(flatDirectory, f)) }))
	let totalSize = files.reduce((subtotal, f) => subtotal + f.stat.size, 0)
	const deletedFiles = []
	for (const f of files.sort((a, b) => a.stat.mtimeMs - b.stat.mtimeMs)) {
		if (totalSize <= max) {
			break
		}
		onBeforeDelete(f.path)
		fs.unlinkSync(f.path)
		totalSize = totalSize - f.stat.size
		deletedFiles.push(f.path)
	}
	return deletedFiles
}

function diskKey({ repository, commit }: RepositoryCommit): string {
	const base64Repository = Buffer.from(repository).toString('base64')
	return path.join(storageRoot, `base64repository:${base64Repository},commit:${commit}.lsif`)
}

async function createDB(repositoryCommit: RepositoryCommit): Promise<Database> {
	const db = new JsonDatabase()
	await db.load(diskKey(repositoryCommit), projectRoot => ({
		toDatabase: path_ => projectRoot + '/' + path_,
		fromDatabase: path_ => path_.startsWith(projectRoot) ? path_.slice(`${projectRoot}/`.length) : path_
	}))
	return db
}

const supportedMethods = ['hover', 'definitions', 'references']

function checkRepository(repository: any): void {
	if (typeof repository !== 'string') {
        throw Object.assign(new Error ('must specify the repository (usually of the form github.com/user/repo)'), { status: 400 })
    }
}

function checkCommit(commit: any): void {
	if (typeof commit !== 'string' || commit.length !== 40 || !/^[0-9a-f]+$/.test(commit)) {
        throw Object.assign(new Error ('must specify the commit as a 40 character hash ' + commit), { status: 400 })
    }
}

interface LRUDBEntry {
	dbPromise: Promise<Database>
	length: number
	dispose: () => void
}

const dbLRU = new LRU<String, LRUDBEntry>({
	max: softMaxDBSizeLoadedIntoMemory,
	length: (entry, key) => entry.length,
	dispose: (key, entry) => entry.dispose()
})

async function withDB(repositoryCommit: RepositoryCommit, action: (db: Database) => Promise<void>): Promise<void> {
	const entry = dbLRU.get(diskKey(repositoryCommit))
	if (entry) {
		await action(await entry.dbPromise)
	} else {
		//here
		const length = fs.statSync(diskKey(repositoryCommit)).size
		const dbPromise = createDB(repositoryCommit)
		dbLRU.set(diskKey(repositoryCommit), {dbPromise:dbPromise,length:length,dispose: () => dbPromise.then(db => db.close())})
		await action(await dbPromise)
	}
}

function main() {
	const app = express()

	app.use(fileUpload({
		limits: { fileSize: maxFileSize },
		abortOnLimit: true,
		responseOnLimit: 'Rejecting LSIF file larger than 100MiB to protect against running out of memory.'
	}));
	app.use(bodyParser.json({ limit: '1mb' }))
	app.use(cors())

	app.get('/ping', (req, res) => {
		res.send({ pong: 'pong' })
	})

	app.post(
		'/request',
		asyncHandler(async (req, res) => {
            const {
                repository,
                commit,
            } = req.query

            const {
                method,
                params
            } = req.body

            checkRepository(repository)
            checkCommit(commit)
            if (!supportedMethods.includes(method)) {
                throw Object.assign(new Error('method must be one of ' + supportedMethods), { status: 400 })
            }

			try {
				await withDB({repository, commit}, async db => {
					res.send((db as any)[method](...params) || { error: 'No result found' })
				})
			}catch(e){
				if ('code' in e && e.code === 'ENOENT') {
					res.send({'error': `No LSIF data available for ${repository}@${commit}.`})
					return
				}
			}
		})
	)

	app.get(
		'/haslsif',
		asyncHandler(async (req, res) => {
            const {
                repository,
                commit,
            } = req.query

            const file = req.body.file

            checkRepository(repository)
            checkCommit(commit)

			if (!file) {
				res.send(fs.existsSync(diskKey({repository,commit})))
                return
			}

            if (typeof file !== 'string') {
                throw Object.assign(new Error('file must be a string'), { status: 400 })
            }

            res.send(Boolean((await createDB({repository,commit})).stat(file)))
		})
	)

	app.post('/upload',
		asyncHandler(async (req, res) => {
            const {
                repository,
                commit,
            } = req.query

            checkRepository(repository)
            checkCommit(commit)

			if (!req.files) {
				res.status(400).send('Expected a file upload.')
				return
			}
			const ufiles = Object.entries(req.files)
			const ufile = ufiles[0][1]
			if (!ufile) {
				res.status(400).send('No file was uploaded.');
				return
			}
			if (ufiles.length > 1) {
				res.status(400).send('Only one file can be uploaded at a time.');
				return
			}
			// I suspect @types/express-fileupload is wrong here.
			if (ufile instanceof Array) {
				res.status(400).send('Data must be a single file.');
				return
			}
			for (const [index, line] of ufile.data.toString().split('\n').entries()) {
				if (line === '') {
					continue
				}
				try {
					JSON.parse(line)
				} catch (e) {
					res.status(422).send(`Expected uploaded file to be in newline separated JSON format. First 100 characters of the first offending line ${index}: ${line.slice(0, 100)})`)
					return
				}
			}

			const deletedFiles = enforceMaxDiskUsage({
				flatDirectory: storageRoot,
				max: Math.max(0, softMaxStorageSize - ufile.data.byteLength),
				onBeforeDelete: filePath => console.log(`Deleting ${filePath} to help keep disk usage under ${softMaxStorageSize}.`)
			})

			await ufile.mv(diskKey({repository, commit}))
			res.send(['Upload successful.', ...(deletedFiles.length > 0 ? ['Deleted old files to make room on disk:', ...deletedFiles.map(f => '- ' + f)] : [])].join('\n'))
		}))

	// TODO make this configurable
	app.listen(5000, () => {
		console.log('Listening for HTTP requests on port 5000')
	})
}

main()

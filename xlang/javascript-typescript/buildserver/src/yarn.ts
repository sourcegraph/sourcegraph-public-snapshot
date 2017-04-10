
import { spawn } from 'child_process';
import { ChildProcess } from 'child_process';
import { Logger, NoopLogger } from 'javascript-typescript-langserver/lib/logging';
import { Span } from 'opentracing';
import * as path from 'path';
import { Readable } from 'stream';

/**
 * Emitted value for a `step` event
 */
export interface YarnStep {
	message: string;
	current: number;
	total: number;
}

/**
 * Child process that emits additional events from yarn's JSON stream
 */
export interface YarnProcess extends ChildProcess {
	/** Emitted for verbose logs */
	on(event: 'verbose', listener: (log: string) => void): this;
	/** Emitted on a yarn step (e.g. Resolving, Fetching, Linking) */
	on(event: 'step', listener: (step: YarnStep) => void): this;
	/** Emitted if the process exited successfully */
	on(event: 'success', listener: () => void): this;
	/** Emitted on error event or non-zero exit code */
	on(event: 'error', listener: (err: Error) => void): this;
	on(event: 'exit', listener: (code: number, signal: string) => void): this;
	on(event: string, listener: () => any): this;

	/** Emitted for verbose logs */
	once(event: 'verbose', listener: (log: string) => void): this;
	/** Emitted on a yarn step (e.g. Resolving, Fetching, Linking) */
	once(event: 'step', listener: (step: YarnStep) => void): this;
	/** Emitted if the process exited successfully */
	once(event: 'success', listener: () => void): this;
	/** Emitted on error event or non-zero exit code */
	once(event: 'error', listener: (err: Error) => void): this;
	once(event: 'exit', listener: (code: number, signal: string) => void): this;
	once(event: string, listener: () => any): this;
}

export interface InstallOptions {

	/** The folder to run yarn in */
	cwd: string;

	/** The global directory to use (`--global-folder`) */
	globalFolder: string;

	/** The cache directory to use (`--cache-folder`) */
	cacheFolder: string;

	/** Whether to run yarn in verbose mode (`--verbose`) to get verbose events (e.g. "Copying file from A to B") */
	verbose?: boolean;

	/** Logger to use */
	logger?: Logger;
}

/**
 * Spawns a yarn child process.
 * The returned child process emits additional events from the streamed JSON events yarn writes to STDIO.
 * An exit code of 0 causes a `success` event to be emitted, any other an `error` event
 *
 * @param options
 * @param childOf OpenTracing parent span for tracing
 */
export function install(options: InstallOptions, childOf = new Span()): YarnProcess {
	const logger = options.logger || new NoopLogger();
	const span = childOf.tracer().startSpan('yarn install', { childOf });
	const args = [
		path.resolve(__dirname, '..', 'node_modules', 'yarn', 'lib', 'cli', 'index.js'),
		'--ignore-scripts',  // Don't run package.json scripts
		'--ignore-platform', // Don't error on failing platform checks
		'--ignore-engines',  // Don't check package.json engines field
		'--no-bin-links',    // Don't create bin symlinks
		'--no-lockfile',     // Don't read or create a lockfile
		'--no-emoji',        // Don't use emojis in output
		'--non-interactive', // Don't ask for any user input
		'--no-progress',     // Don't report progress events
		'--json',            // Output a newline-delimited JSON stream
		// '--link-duplicates', // Use hardlinks instead of copying, not working reliably because of https://github.com/yarnpkg/yarn/issues/2734

		// Use a separate global and cache folders per package.json
		// that we can clean up afterwards and don't interfere with concurrent installations
		'--global-folder', options.globalFolder,
		'--cache-folder', options.cacheFolder
	];
	if (options.verbose) {
		args.push('--verbose');
	}
	const yarn: YarnProcess = spawn(process.execPath, args, { cwd: options.cwd });

	/** Emitted error messages by yarn */
	const errors: string[] = [];

	function parseStream(stream: Readable) {
		let buffer = '';
		stream.on('data', chunk => {
			try {
				buffer += chunk;
				const lines = buffer.split('\n');
				buffer = lines.pop()!;
				for (const line of lines) {
					const event = JSON.parse(line);
					switch (event.type) {
						case 'error':
							// Only emit error event if non-zero exit code
							logger.error('yarn: ', event.data);
							errors.push(event.data);
							break;
						case 'step':
						case 'verbose':
							yarn.emit(event.type, event.data);
							break;
					}
				}
			} catch (err) {
				// E.g. JSON parse error
				yarn.emit(err);
			}
		});
	}

	// Yarn writes JSON messages to both STDOUT and STDERR depending on event type
	parseStream(yarn.stdout);
	parseStream(yarn.stderr);

	yarn.on('exit', (code, signal) => {
		if (code === 0) {
			yarn.emit('success');
		} else if (!signal) {
			const error = Object.assign(new Error(`yarn failed with exit code ${code}: ${errors.join(' ')}`), { code, errors });
			logger.error(error);
			yarn.emit('error', error);
		}
		span.finish();
	});

	// Trace steps, e.g. Resolving, Fetching, Linking
	yarn.on('step', step => {
		span.log({ event: 'step', message: step.message });
		logger.log(`${step.current}/${step.total} ${step.message}`);
	});

	// Trace errors
	yarn.on('error', err => {
		span.setTag('error', true);
		span.log({ 'event': 'error', 'error.object': err, 'message': err.message, 'stack': err.stack });
	});

	return yarn;
}

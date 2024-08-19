const process = require('process');

/**
 * Executes a pretend jsii runtime process that advertises the
 * provided runtime version number.
 *
 * It response to any request except for "exit" by repeating it
 * back to the parent process. The "exit" handling is "standard".
 *
 * @param version the version number to report in the HELLO message.
 */
function main(version) {
    console.log(JSON.stringify({ hello: `@mock/jsii-runtime@${version}` }));

    let buffer = "";
    process.stdin.setEncoding('utf8');
    process.stdin.on('data', chunk => {
        buffer = buffer + chunk;
        nl = buffer.indexOf('\n');
        if (nl >= 0) {
            const line = buffer.substring(0, nl + 1);
            buffer = buffer.substring(nl + 1);

            const message = JSON.parse(line);
            if (message.exit) {
                process.exit(message.exit);
            } else {
                console.log(JSON.stringify(message));
            }
        }
    });
}

main(...process.argv.slice(2));
<?php

use LanguageServer\{ProtocolStreamReader, ProtocolStreamWriter};
use Sourcegraph\BuildServer\BuildServer;
use Sabre\Event\Loop;

$options = getopt('', ['tcp::', 'tcp-server::', 'memory-limit::']);

ini_set('memory_limit', $options['memory-limit'] ?? -1);

foreach ([__DIR__ . '/../../../autoload.php', __DIR__ . '/../autoload.php', __DIR__ . '/../vendor/autoload.php'] as $file) {
    if (file_exists($file)) {
        require $file;
        break;
    }
}

// Convert all errors to ErrorExceptions
set_error_handler(function (int $severity, string $message, string $file, int $line) {
    if (!(error_reporting() & $severity)) {
        // This error code is not included in error_reporting (can also be caused by the @ operator)
        return;
    }
    throw new \ErrorException($message, 0, $severity, $file, $line);
});

// Only write uncaught exceptions to STDERR, not STDOUT
set_exception_handler(function (\Throwable $e) {
    fwrite(STDERR, (string)$e);
});

@cli_set_process_title('PHP Build Server');

if (!empty($options['tcp'])) {
    // Connect to a TCP server
    $address = $options['tcp'];
    $socket = stream_socket_client('tcp://' . $address, $errno, $errstr);
    if ($socket === false) {
        fwrite(STDERR, "Could not connect to language client. Error $errno\n$errstr");
        exit(1);
    }
    stream_set_blocking($socket, false);
    $ls = new BuildServer(
        new ProtocolStreamReader($socket),
        new ProtocolStreamWriter($socket)
    );
    Loop\run();
} else if (!empty($options['tcp-server'])) {
    // Run a TCP Server
    $address = $options['tcp-server'];
    $tcpServer = stream_socket_server('tcp://' . $address, $errno, $errstr);
    if ($tcpServer === false) {
        fwrite(STDERR, "Could not listen on $address. Error $errno\n$errstr");
        exit(1);
    }
    fwrite(STDOUT, "Server listening on $address\n");
    if (!extension_loaded('pcntl')) {
        fwrite(STDERR, "PCNTL is not available. Only a single connection will be accepted\n");
    }
    while ($socket = stream_socket_accept($tcpServer, -1)) {
        fwrite(STDOUT, "Connection accepted\n");
        stream_set_blocking($socket, false);
        if (extension_loaded('pcntl')) {
            // If PCNTL is available, fork a child process for the connection
            // An exit notification will only terminate the child process
            $pid = pcntl_fork();
            if ($pid === -1) {
                fwrite(STDERR, "Could not fork\n");
                exit(1);
            } else if ($pid === 0) {
                // Child process
                $reader = new ProtocolStreamReader($socket);
                $writer = new ProtocolStreamWriter($socket);
                $reader->on('close', function () {
                    fwrite(STDOUT, "Connection closed\n");
                });
                $ls = new BuildServer($reader, $writer);
                Loop\run();
                // Just for safety
                exit(0);
            }
        } else {
            // If PCNTL is not available, we only accept one connection.
            // An exit notification will terminate the server
            $ls = new BuildServer(
                new ProtocolStreamReader($socket),
                new ProtocolStreamWriter($socket)
            );
            Loop\run();
        }
    }
} else {
    // Use STDIO
    stream_set_blocking(STDIN, false);
    $ls = new BuildServer(
        new ProtocolStreamReader(STDIN),
        new ProtocolStreamWriter(STDOUT)
    );
    Loop\run();
}

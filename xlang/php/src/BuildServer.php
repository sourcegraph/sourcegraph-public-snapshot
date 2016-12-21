<?php
declare(strict_types = 1);

namespace Sourcegraph\BuildServer;

use LanguageServer\LanguageServer;
use LanguageServer\Protocol\{
    ServerCapabilities,
    ClientCapabilities,
    InitializeResult
};
use Sabre\Event\Promise;
use Composer;

class BuildServer extends LanguageServer
{
    /**
     * Installs dependencies before calling the language server's initialize() method
     *
     * @param ClientCapabilities $capabilities The capabilities provided by the client (editor)
     * @param string|null        $rootPath     The rootPath of the workspace. Is null if no folder is open.
     * @param int|null           $processId    The process Id of the parent process that started the server. Is null if the process has not been started by another process. If the parent process is not alive then the server should exit (see exit notification) its process.
     * @return Promise <InitializeResult>
     */
    public function initialize(ClientCapabilities $capabilities, string $rootPath = null, int $processId = null): Promise
    {
        if ($rootPath && file_exists($rootPath . '/composer.json')) {
            $io = new IO;
            $composerFactory = new Composer\Factory;
            $composer = $composerFactory->createComposer($io, $rootPath . '/composer.json', true, $rootPath);
            $installer = Composer\Installer::create($io, $composer);
            $installer->run();
        }
        return parent::initialize($capabilities, $rootPath, $processId);
    }
}

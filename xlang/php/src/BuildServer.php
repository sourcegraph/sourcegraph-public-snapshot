<?php
declare(strict_types = 1);

namespace Sourcegraph\BuildServer;

use LanguageServer\{LanguageServer, ProtocolReader, ProtocolWriter};
use Composer;
use LanguageServer\ContentRetriever\{ContentRetriever, FileSystemContentRetriever};
use LanguageServer\FilesFinder\{FilesFinder, FileSystemFilesFinder};
use LanguageServer\Protocol\MessageType;
use Sabre\Event\Promise;
use Sabre\Uri;
use Webmozart\PathUtil\Path;
use Hirak\Prestissimo;
use Symfony\Component\Filesystem\Filesystem;
use function Sabre\Event\coroutine;

class BuildServer extends LanguageServer
{
    /**
     * @var string
     */
    private $dependenciesDir;

    /**
     * @var Filesystem
     */
    private $fs;

    public function __construct(ProtocolReader $reader, ProtocolWriter $writer)
    {
        parent::__construct($reader, $writer);
        $this->fs = new Filesystem;
    }

    protected function beforeIndex(string $rootPath)
    {
        return coroutine(function () use ($rootPath) {
            $composerJsonFiles = yield $this->filesFinder->find(Path::makeAbsolute('**/composer.json', $rootPath));
            if (!empty($composerJsonFiles)) {
                $composerJsonFile = $composerJsonFiles[0];

                try {
                    // Create random temporary folder
                    $this->dependenciesDir = sys_get_temp_dir() . '/phpbs' . time() . random_int(100000, 999999);
                    $this->fs->mkdir($this->dependenciesDir);

                    $composerJsonContent = yield $this->contentRetriever->retrieve($composerJsonFile);
                    $this->composerJson = json_decode($composerJsonContent);

                    // Write composer.json
                    file_put_contents($this->dependenciesDir . '/composer.json', $composerJsonContent);

                    // Install dependencies
                    $this->client->window->logMessage(MessageType::INFO, 'Installing dependencies to ' . $this->dependenciesDir . "\n");
                    $io = new IO($this->client);
                    $composerFactory = new Composer\Factory;
                    $composer = $composerFactory->createComposer($io, $this->dependenciesDir . '/composer.json', true, $this->dependenciesDir);
                    $installer = Composer\Installer::create($io, $composer);
                    // Prefer tarballs over git clones
                    $installer->setPreferDist(true);
                    // Disable autoloader generation, it can cause exceptions
                    $installer->setDumpAutoloader(false);
                    // Disable script execution
                    $installer->setRunScripts(false);
                    // Download in parallel
                    $composer->getPluginManager()->addPlugin(new Prestissimo\Plugin);
                    $code = $installer->run();
                    $this->client->window->logMessage(MessageType::LOG, "Installer exited with $code");
                    // Get the composer.json directory
                    $parts = Uri\parse($composerJsonFile);
                    $parts['path'] = dirname($parts['path']);
                    $composerJsonDir = Uri\build($parts);

                    // Read the generated composer.lock to get information about resolved dependencies
                    if ($this->fs->exists($this->dependenciesDir . '/composer.lock')) {
                        $this->composerLock = json_decode(file_get_contents($this->dependenciesDir . '/composer.lock'));

                        // Make filesFinder and contentRetriever aware of the dependencies installed in the temporary folder
                        $this->filesFinder = new DependencyAwareFilesFinder($this->filesFinder, $rootPath, $composerJsonDir, $this->dependenciesDir, $this->composerLock);
                        $this->contentRetriever = new DependencyAwareContentRetriever($this->contentRetriever, $rootPath, $composerJsonDir, $this->dependenciesDir, $this->composerLock);
                        $this->documentLoader->contentRetriever = $this->contentRetriever;
                        $this->textDocument = new Server\TextDocument(
                            $this->documentLoader,
                            $this->definitionResolver,
                            $this->client,
                            $this->globalIndex,
                            $composerJsonDir,
                            $rootPath,
                            $this->composerJson,
                            $this->composerLock
                        );
                    } else {
                        $this->client->window->logMessage(MessageType::WARNING, "composer.lock not found");
                    }
                } catch (\Exception $e) {
                    $this->client->window->logMessage(MessageType::ERROR, "Installation failed: " . (string)$e);
                }
            }
        });
    }

    public function shutdown()
    {
        parent::shutdown();
        // Delete folder where dependencies were installed to
        if ($this->dependenciesDir !== null) {
            $this->fs->remove($this->dependenciesDir);
        }
    }
}

<?php
declare(strict_types = 1);

namespace Sourcegraph\BuildServer;

use LanguageServer\ContentRetriever\{ContentRetriever, FileSystemContentRetriever};
use Webmozart\PathUtil\Path;
use Sabre\Uri;
use Sabre\Event\Promise;
use function Sabre\Event\coroutine;
use function LanguageServer\pathToUri;

class DependencyAwareContentRetriever implements ContentRetriever
{
    /**
     * @var ContentRetriever
     */
    private $wrappedContentRetriever;

    /**
     * @var FileSystemContentRetriever
     */
    private $fileSystemContentRetriever;

    /**
     * @var string
     */
    private $composerJsonDir;

    /**
     * @var string
     */
    private $dependenciesDir;

    /**
     * @var string
     */
    private $rootPath;

    /**
     * @var \stdClass
     */
    private $composerLock;

    /**
     * @param ContentRetriever $wrappedContentRetriever The ContentRetriever to fallback to if the file was not a dependency
     * @param string           $rootPath                The workspace root path
     * @param string           $composerJsonDir         The URI of the directory in the workspace where the composer.json is located
     * @param string           $dependencyDir           The temporary folder path where the dependencies are actually installed
     * @param stdClass         $composerLock            The parsed composer.lock
     */
    public function __construct(ContentRetriever $wrappedContentRetriever, string $rootPath, string $composerJsonDir, string $dependenciesDir, \stdClass $composerLock)
    {
        $this->wrappedContentRetriever = $wrappedContentRetriever;
        $this->fileSystemContentRetriever = new FileSystemContentRetriever;
        $this->rootPath = $rootPath;
        $this->composerJsonDir = $composerJsonDir;
        $this->dependenciesDir = $dependenciesDir;
        $this->composerLock = $composerLock;
    }

    public function retrieve(string $uri): Promise
    {
        $parts = Uri\parse($uri);
        $composerJsonDirPath = Uri\parse($this->composerJsonDir)['path'];
        // Check if requested file is a Sourcegraph repository URI
        if ($parts['scheme'] !== 'file') {
            // Rewrite URI from repository URI to temporary dependency folder
            // Find the right package name
            foreach ($this->composerLock->packages as $package) {
                if (!isset($package->source) || $package->source->type !== 'git') {
                    continue;
                }
                // Example: https://github.com/felixfbecker/php-language-server.git
                $packageSourceUrlParts = Uri\parse($package->source->url);
                $packageSourceUrlParts['path'] = preg_replace('/\.git$/', '', $packageSourceUrlParts['path']);
                if (
                    $packageSourceUrlParts['host'] === $parts['host']
                    && $packageSourceUrlParts['path'] === $parts['path']
                    && $package->source->reference === $parts['query']
                ) {
                    $workspacePath = Path::join($composerJsonDirPath, 'vendor', $package->name, $parts['fragment']);
                    $relativeDependenciesPath = Path::makeRelative($workspacePath, $composerJsonDirPath);
                    return $this->fileSystemContentRetriever->retrieve(pathToUri(Path::join($this->dependenciesDir, $relativeDependenciesPath)));
                }
            }
        }
        return $this->wrappedContentRetriever->retrieve($uri);
    }
}

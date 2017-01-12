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
        // Check if requested file is inside a dependency
        if (strpos($uri, 'vendor') !== false) {
            $path = Uri\parse($uri)['path'];
            // Rewrite URI from vendor URI to temporary dependency folder
            $composerJsonDirPath = Uri\parse($this->composerJsonDir)['path'];
            $relativeDependenciesPath = Path::makeRelative($path, $composerJsonDirPath);
            $dependenciesPath = Path::join($this->dependenciesDir, $relativeDependenciesPath);
            return $this->fileSystemContentRetriever->retrieve(pathToUri($dependenciesPath));
        }
        return $this->wrappedContentRetriever->retrieve($uri);
    }
}

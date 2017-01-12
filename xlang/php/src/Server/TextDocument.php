<?php
declare(strict_types = 1);

namespace Sourcegraph\BuildServer\Server;

use LanguageServer\{LanguageClient, PhpDocumentLoader, PhpDocument, DefinitionResolver, CompletionProvider};
use LanguageServer\Protocol\{
    TextDocumentIdentifier,
    Position
};
use LanguageServer\Index\ReadableIndex;
use Webmozart\PathUtil\Path;
use Sabre\Event\Promise;
use Sabre\Uri;
use function Sabre\Event\coroutine;

/**
 * Provides method handlers for all textDocument/* methods
 */
class TextDocument extends \LanguageServer\Server\TextDocument
{
    /**
     * @var string
     */
    private $dependencyDir;

    /**
     * @param PhpDocumentLoader  $documentLoader
     * @param DefinitionResolver $definitionResolver
     * @param LanguageClient     $client
     * @param ReadableIndex      $index
     * @param string             $composerJsonDir
     * @param string             $rootPath
     * @param \stdClass          $composerJson
     * @param \stdClass          $composerLock
     */
    public function __construct(
        PhpDocumentLoader $documentLoader,
        DefinitionResolver $definitionResolver,
        LanguageClient $client,
        ReadableIndex $index,
        string $composerJsonDir,
        string $rootPath,
        \stdClass $composerJson = null,
        \stdClass $composerLock = null
    ) {
        parent::__construct($documentLoader, $definitionResolver, $client, $index, $composerJson, $composerLock);
        $this->composerJsonDir = $composerJsonDir;
        $this->rootPath = $rootPath;
    }

    /**
     * The goto definition request is sent from the client to the server to resolve the definition location of a symbol
     * at a given text document position.
     *
     * @param TextDocumentIdentifier $textDocument The text document
     * @param Position $position The position inside the text document
     * @return Promise <Location|Location[]>
     */
    public function definition(TextDocumentIdentifier $textDocument, Position $position): Promise
    {
        return coroutine(function () use ($textDocument, $position) {
            $location = yield parent::definition($textDocument, $position);
            // Get package name from URI
            if (preg_match('/\/vendor\/([^\/]+\/[^\/]+)\//', $location->uri, $matches)) {
                $packageName = $matches[1];
                $composerJsonDirPath = Uri\parse($this->composerJsonDir)['path'];
                $relativeComposerJsonDirPath = Path::makeRelative($composerJsonDirPath, $this->rootPath);
                foreach ($this->composerLock->packages as $package) {
                    if ($package->name !== $packageName) {
                        continue;
                    }
                    if (!isset($package->source) || $package->source->type !== 'git') {
                        // Not supported atm
                        throw new \Exception('No or non-git package source');
                    }
                    // Example: https://github.com/felixfbecker/php-language-server.git
                    $parts = Uri\parse($package->source->url);
                    $parts['scheme'] = 'git';
                    $parts['path'] = preg_replace('/\.git$/', '', $parts['path']);
                    $parts['query'] = $package->source->reference;
                    $dependencyPath = Uri\parse($location->uri)['path'];
                    $relativeDependencyPath = Path::makeRelative($dependencyPath, Path::join($composerJsonDirPath, "vendor/$packageName"));
                    $parts['fragment'] = trim($relativeDependencyPath, '/');
                    $location->uri = Uri\build($parts);
                    return $location;
                }
                throw new \Exception('Package metadata not found');
            }
            return $location;
        });
    }
}

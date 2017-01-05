<?php
declare(strict_types = 1);

namespace Sourcegraph\BuildServer;

use LanguageServer\LanguageServer;
use Composer;
use LanguageServer\ContentRetriever\{ContentRetriever, FileSystemContentRetriever};
use LanguageServer\FilesFinder\{FilesFinder, FileSystemFilesFinder};
use Sabre\Event\Promise;
use Sabre\Uri;
use Webmozart\PathUtil\Path;
use Hirak\Prestissimo;
use function Sabre\Event\coroutine;

class BuildServer extends LanguageServer
{
    protected function index(string $rootPath): Promise
    {
        return coroutine(function () use ($rootPath) {
            $composerJsonFiles = yield $this->filesFinder->find(Path::makeAbsolute('**/composer.json', $rootPath));
            if (!empty($composerJsonFiles)) {
                $composerJsonFile = $composerJsonFiles[0];

                // Create random temporary folder
                $dir = sys_get_temp_dir() . '/phpbs' . time() . random_int(100000, 999999);
                mkdir($dir);

                // Write composer.json
                file_put_contents("$dir/composer.json", yield $this->contentRetriever->retrieve($composerJsonFile));

                // Install dependencies
                fwrite(STDERR, "Installing dependencies to $dir\n");
                $io = new IO;
                $composerFactory = new Composer\Factory;
                $composer = $composerFactory->createComposer($io, "$dir/composer.json", true, $dir);
                $installer = Composer\Installer::create($io, $composer);
                // Prefer tarballs over git clones
                $installer->setPreferDist(true);
                // Download in parallel
                $composer->getPluginManager()->addPlugin(new Prestissimo\Plugin);
                $installer->run();
                
                // Get the composer.json directory
                $parts = Uri\parse($composerJsonFile);
                $parts['path'] = dirname($parts['path']);
                $composerJsonDir = Uri\build($parts);

                // Read the generated composer.lock to get information about resolved dependencies
                $composerLock = json_decode(file_get_contents("$dir/composer.lock"));

                // Make filesFinder and contentRetriever aware of the dependencies installed in the temporary folder
                $this->filesFinder = new DependencyAwareFilesFinder($this->filesFinder, $rootPath, $composerJsonDir, $dir, $composerLock);
                $this->contentRetriever = new DependencyAwareContentRetriever($this->contentRetriever, $rootPath, $composerJsonDir, $dir, $composerLock);
                $this->documentLoader->contentRetriever = $this->contentRetriever;
            }
            yield parent::index($rootPath);
        });
    }
}

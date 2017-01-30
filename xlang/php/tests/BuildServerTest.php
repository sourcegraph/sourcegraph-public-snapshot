<?php
declare(strict_types = 1);

namespace Sourcegraph\BuildServer\Tests;

use PHPUnit\Framework\TestCase;
use LanguageServer\Protocol\ClientCapabilities;
use LanguageServer\Tests\MockProtocolStream;
use Sourcegraph\BuildServer\BuildServer;

class BuildServerTest extends TestCase
{
    public function testInitializeInstallsDependencies()
    {
        $server = new BuildServer(new MockProtocolStream, new MockProtocolStream);

        $capabilities = new ClientCapabilities;
        $rootPath = realpath(__DIR__ . '/../fixtures');
        $pid = getmypid();

        $server->initialize($capabilities, $rootPath, $pid)->wait();

        // Initialize should install into temporary folder
        $dirs = glob(sys_get_temp_dir() . '/phpbs*');
        $this->assertNotEmpty($dirs);
        $dir = $dirs[0];

        // Folder should contain dependencies
        $this->assertDirectoryExists("$dir/vendor");
        $this->assertDirectoryExists("$dir/vendor/psr/log");
        $this->assertFileExists("$dir/composer.lock");

        // Shutdown should delete the temporary folder
        $server->shutdown();
        $this->assertDirectoryNotExists($dir);
    }
}

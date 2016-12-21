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

        $result = $server->initialize($capabilities, $rootPath, $pid)->wait();

        $this->assertDirectoryExists("$rootPath/vendor");
        $this->assertDirectoryExists("$rootPath/vendor/psr/log");
        $this->assertFileExists("$rootPath/composer.lock");
    }
}

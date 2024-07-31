namespace SomeNamespace {
    const int some_const = 100;
    trait SomeTrait {
        public $logLevel;
        protected $thing;
        private $secretFromTrait;

        public function setLogger(SomeInterface $thing) {
            $this->thing = $thing;
        }

        public function log($message, $level) {
            $this->thing->log($message, $level);
        }
    }

    interface SomeInterface {
        const MAX_NUMBER_ITEMS = 1000;
        protected int $secretFromInterface = 0;
        public function log($message, $level);
    }

    class SomeClass implements SomeInterface {
        const PI = 3.1415926;
        private static int $secretFromClass = 0;
        public static int $hello = 11;
        public int $age = 39;
        public function log($message, $level) {
            echo "Log $message of level $level";
        }
    }

    class Foo implements SomeInterface {
        const type T = string;
        use SomeTrait;
    }

    type Foo_alias = Foo;
    newtype Foo_new = Foo::T;

    // Top level function
    <<__EntryPoint>>
    function main() {
        $foo = new Foo;
        $foo->setLogger(new SomeClass);
        $foo->log('It works', 1);
    }
}

namespace SomeNamespace\SubNamespace {
    // Generic class and a constructor
    class Stack<T> {
        private vec<T> $stack;
        private int $stackPtr;

        public function __construct() {
            $this->stackPtr = 0;
            $this->stack = vec[];
        }

        public function __dispose(): void {}
    }

    enum Colors: int {
        Red = 1;
        Green = 2;
        Blue = 3;
        Default = 4;
    }

    enum class Random: mixed {
        int X = 42;
        string S = 'foo';
    }
}

// Validate anonymous namespace
namespace {
    const int another_const = 88;
}

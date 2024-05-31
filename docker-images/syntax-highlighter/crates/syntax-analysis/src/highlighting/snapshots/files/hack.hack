// imports
use namespace HH\Lib\{C, Vec};
use const Space\Const\C;
use type Space\Type\T;
use function UseNS\f;
use T;

// File includes
require_once(__DIR__.'/file.hack');
require(__DIR__.'/file.hack');
include(__DIR__.'/file.hack');
include_once(__DIR__.'/file.hack');

// NOTE: Hack tree-sitter grammar doest not support module
// https://github.com/slackhq/tree-sitter-hack/issues/70
module foo.bar.baz;

namespace Hack\Example\namespace {

    // Functions
    function f(dynamic $d): void {}
    function g(arraykey $a): void {}
    function h(num $a = 1): void {}

    // Shapes
    type Shape1 = shape('name' => string, ?'age' => int);
    type Shape2 = shape('age' => int);
    function foo(int $arg): shape(...){

    $a = shape();
    $a = shape('name' => 'db-01', 'age' => 365);
    return $a;
    }
}

// Constants
const int MAX_COUNT = 123;

/**
 * A doc comment starts with two asterisks.
 */
function swap<T>(inout T $i1, inout T $i2): void {
  $temp = $i1;

  // Anonymous functions
  $f = function($x) { return $x + 1; };
  $f = function($x) use($y) { return $x + $y; };
  $f = $x ==> $x + 1;
  $f = (int $x): int ==> $x + 1;
}

// Generators
function squares(
  int $start,
  int $end,
  string $keyPrefix = "",
): Generator<string, int, void> {
  for ($i = $start; $i <= $end; ++$i) {
    yield $keyPrefix.$i => $i * $i; // specify a key/value pair
  }
}

<<__EntryPoint>>
function main(): void {
  $v1 = -10;
  $v2 = "red";
  $a1 = "foo"."bar";

  // Built-ins
  $keyedcontainer = dict[];
  $r = idx($keyedcontainer, 'key', 23);
  invariant($a1 is string, "Object must have type B");
  echo "\$v1 = ".$v1.", \$v2 = ".$v2."\n";
  exit ("Closing down\n");

  // NOTE: Tree-sitter grammar does not support
  //       string interpolation
  //       https://github.com/slackhq/tree-sitter-hack/issues/69
  $y = "hello $x[0]";
  $y = "hello $x->foo";


  // Control Flow
  $i = 1;
  do {
    ++$i;
  } while ($i <= 10);

  foreach (($array as vec[]) as $item) {}

  for (; $i <= 5; ) {
    ++$i;
    continue;
  }

  try {
    echo "try this";
  } catch (Exception $ex) {
    echo "Caught an Exception\n";
  } finally {
    echo "Finally\n";
  }

  using ($new = new Object(), $file = new File('using', '+using')) {}

final class C {
    function f(classname<C> $clsname): void {
        $w = new $clsname();
    }
   function cons_static() :mixed{
     $a = new static(1, "x", 3);
   }
   function cons_self(): void {
     $a = new self(1, "x", 3);
   }
   function cons_self(): void {
     $a = new parent(1, "x", 3);
   }
}


  switch ($pos) {
    case Position::Bottom:
      break;
  }

  // Magic Constants (some of them)
  $a = __CLASS__;
  $a = __DIR__;
  $a = __FILE__;

  // Expressions
  $tuple = tuple('one', 'two', 'three');
  list($one, $two, $three) = $tuple;

  // Ternary
  $x = foo() ?: bar();
  $x = $tmp ? $tmp : bar();

  // Coalescing
  $a = $b ?? 4;
  $a ??= $b;

  // Type assertions
  $a = 1 ?as int;
  $a = 1 as int;
  $a = 1 is int;
  $a = is_int(1);
  $a = is_bool(1);
  $a = is_string(1);

  $infile = @fopen("NoSuchFile.txt", 'r');

  $d = dict[];
  $xhp = <tt>Hello <strong>{$user_name}</strong>
    Text in the markup
    <!-- this is a comment -->
</tt>;


  // Literals
  // Note: nameof missing from the tree-sitter grammar
  //       https://github.com/slackhq/tree-sitter-hack/issues/68
  $d[nameof C] = 4;
  $v = vec[1, 2, 3];
  $k = keyset[2, 1];
  $d = dict['a' => 1, 'b' => 3];
  $v[0] = 42;
  $a1 = (bool)0;
  $a = 0b101010;
  $a = 0XAf;
  $f = 123.456 + 0.6E27 + 2.34e-3;
  $f = NAN;
  $f = INF;
  $x = tuple(1, 2.0, null);
  $x is (~int, @float, ?bool);
  $s = shape('name' => 'db-01', 'age' => 365);
  $x = true;
  $y = false;
  $x = True;
  $y = FALSE;

  // NOTE: Grammar does not support the _ separator properly
  //       https://github.com/slackhq/tree-sitter-hack/issues/72
  $a = 123_456;
  $a = 0x49AD_DF30;
  $f = 123_456.49_30e-30_30;

// nowdoc
$s = <<< 'ID'
    $('a') abc $(function{return;})
ID;

// heredoc
$s = <<<ID
	$('a') abc $(function{return;})
ID;

  // Pipe
  $x = vec[2,1,3]
    |> Vec\map($$, $a ==> $a * $a)
    |> Vec\sort($$);

  // Operators
  $a1 = -10 + 100;
  $a1 = 2 ** 10;
  $a1 = 100 + -3.4e2;
  $a1 = 9.5 + 23.444;
  $a1 = (1 << 63) >> 63;
  $a1 = 1 > 2;
  $a = $a & ~0x20;
  $a = $a ^ ~0x20;
  $a = $a || $b;
  $a = $a && $b;
  $a = !$a;
  $a = $a++;
  $a = $a--;
  $a -= 1;
  $a +=1;
  $a **=1;

  // Comparisons
  $a = 1 > 2;
  $a = 1 < 2;
  $a = 1 == 2;
  $a = 1 != 2;
  $a = 1 === 2;
  $a = 1 !== 2;
  $a = 1 <=> 2;
  $a = $a is nonnull;
}

// Types
type Complex = shape('real' => float, 'imag' => float);
newtype Point = (float, float);

// return types
<<__Memoize>>
function noreturn_example(): noreturn {
  throw new Exception('something went wrong');
}
<<Contributors("John Doe", keyset["Core Library Team"])>>
function nothing_example(): nothing {
  throw new Exception('something went wrong');
}
function f<<<__Newable>> reify T as A>(): T {
  return new T();
}

// Async/await
async function main_async(): Awaitable<void> {
  concurrent {
      $out = IO\request_output();
      await $out->writeAllAsync("Hello, world\n");
   }
}

// Enums
enum Position: int {
  Top = 0;
  Bottom = 1;
}
enum class Random: mixed {
  int X = 42;
  string S = 'foo';
}

// Interfaces
interface StackLike<+T> {
  public function push(T $element): void;
}

// Traits
trait T1 implements I1 {
   // NOTE: readonly is not supported properly in tree-sitter grammar
   //       https://github.com/slackhq/tree-sitter-hack/issues/71
   public static readonly int $x = 0;

   static function inc() : void {
    static::$x = static::$x + 1;
  }
}
class A1 implements I1 { use T1; }

// Classes
class B<reify T> {}

abstract final class F<Ta as A, Tb super B<A, C>> extends B implements A\B<A, C>, C\D {
  function method<Ta as A, Tb super B>(): Tc {}
}

// XHP Attributes
final xhp class user_info extends x\element {
  attribute int userid @required;
  attribute string name = "";

  protected async function renderAsync(): Awaitable<x\node> {
    return
      <x:frag>User with id {$this->:userid} has name {$this->:name}</x:frag>;
  }
}

internal class StackUnderflowException extends \Exception {}

class VecStack<T> implements StackLike<T> {
  private int $stackPtr;

  // Constraints
  public function flatten<Tu>(): MyList<Tu> where T = MyList<Tu> {
    throw new Exception('unimplemented');
  }
  public function __construct(private vec<T> $elements = vec[]) {
    $this->stackPtr = C\count($elements) - 1;
    $a = $elements?->getX();
  }

  public function push(T $element): void {
    $this->stackPtr++;
    if (C\count($this->elements) === $this->stackPtr) {
      $this->elements[] = $element;
    } else {
      $this->elements[$this->stackPtr] = $element;
    }
  }
}

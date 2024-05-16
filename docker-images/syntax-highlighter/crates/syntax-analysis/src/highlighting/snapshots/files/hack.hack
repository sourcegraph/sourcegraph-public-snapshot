// imports
use namespace HH\Lib\{C, Vec};
use const Space\Const\C;
use type Space\Type\T;

// Functions
function f(dynamic $d): void {}
function g(arraykey $a): void {}

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
  echo "\$v1 = ".$v1.", \$v2 = ".$v2."\n";

  // Control Flow
  $i = 1;
  do {
    ++$i;
  } while ($i <= 10);

  foreach (($array as vec[]) as $item) {}

  for (; $i <= 5; ) {
    ++$i;
  }

  using ($new = new Object(), $file = new File('using', '+using')) {}

  switch ($pos) {
    case Position::Bottom:
      break;
  }

  // Expressions
  $d = dict[];
  $xhp = <tt>Hello <strong>{$user_name}</strong>
    <!-- this is a comment -->
</tt>;
  $d[nameof C] = 4;
  $v = vec[1, 2, 3];
  $v[0] = 42;
  $a1 = (bool)0;
  $x = tuple(1, 2.0, null);
  $x is (~int, @float, ?bool);
  $s = shape('name' => 'db-01', 'age' => 365);

// heredoc
<<<EOT
	$('a') abc $(function{return;})
EOT;

  // Pipe
  $x = vec[2,1,3]
    |> Vec\map($$, $a ==> $a * $a)
    |> Vec\sort($$);

  // Arithmetic
  $a1 = -10 + 100;
  $a1 = 100 + -3.4e2;
  $a1 = 9.5 + 23.444;
  $a1 = (1 << 63) >> 63;
  $a1 = 1 > 2;
}

// Types
type Complex = shape('real' => float, 'imag' => float);
newtype Point = (float, float);

// Async/await
async function main_async(): Awaitable<void> {
  $out = IO\request_output();
  await $out->writeAllAsync("Hello, world\n");
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
interface StackLike<T> {
  public function push(T $element): void;
}

// Traits
trait T1 implements I1 {
  public static int $x = 0;

  public static function inc() : void {
    static::$x = static::$x + 1;
  }
}
class A1 implements I1 { use T1; }

// Classes
abstract final class F<Ta as A, Tb super B<A, C>> extends B implements A\B<A, C>, C\D {
  function method<Ta as A, Tb super B>(): Tc {}
}

class StackUnderflowException extends \Exception {}

class VecStack<T> implements StackLike<T> {
  private int $stackPtr;

  // Constraints
  public function flatten<Tu>(): MyList<Tu> where T = MyList<Tu> {
    throw new Exception('unimplemented');
  }

  public function __construct(private vec<T> $elements = vec[]) {
    $this->stackPtr = C\count($elements) - 1;
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

namespace Main;

public class CSharp {
    public string Name;

    public CSharp(int name)
    {
        var something = 42;
        Name = "name";
        Console.log(something);
    }

    public CSharp(string name) => Name = name;


    ~CSharp()
    {
        Console.WriteLine(42);
    }

    public class UnmanagedConstraintClass<T> where T : unmanaged
    {
    }
    public class DefaultConstraintClass<T> where T : default
    {
    }
    public class NotNullConstraintClass<T> where T : notnull
    {
    }
    public class NewConstraintClass<T> where T : new()
    {
    }

    class IndexClass
    {
        private bool a;

        public bool this[int index]
        {
            get { return a; }
            set { a = value; }
        }
    }

    void SunsetRestrictedTypes()
    {
        var g = "";
        var g = "";
        var reference = __makeref(g);
        Console.WriteLine(__refvalue(reference, int));
        Console.WriteLine(__reftype(reference));
    }

    enum A
    {
        B,
        C
    }

    string Interpolation()
    {
        var a = 1;
        var b = 2;
        var c = 3;
        var d = 3;
        return $"a={a} b={b:0.00} c={c,24} d={d:g}";
    }

    class Operators
    {
        public static bool operator true(TrueFalse a)
        {
            return true;
        }

        public static bool operator false(TrueFalse a)
        {
            return false;
        }

        public static bool operator !=(TrueFalse a, TrueFalse b)
        {
            return true;
        }

        public static bool operator ==(TrueFalse a, TrueFalse b)
        {
            return true;
        }
    }

    class PlusMinusOperators
    {
        public static int operator +(PlusMinus a)
        {
            return 0;
        }

        public static int operator +(PlusMinus a, PlusMinus b)
        {
            return 0;
        }

        public static int operator -(PlusMinus a)
        {
            return 0;
        }
    }

    public class Preprocessors
    {
        string OS()
        {
    #if WIN32
            string os = "Win32";
    #warning This class is bad.
    #error Okay, just stop.
    #elif MACOS
            string os = "MacOS";
    #else
            string os = "Unknown";
    #endif
            return os;
        }
    }

    void Linq()
    {
        var x = from a in sourceA
                join b in sourceB on a.Method() equals b.Method()
                select new { A = a.Method(), B = b.Method() };
    }


}

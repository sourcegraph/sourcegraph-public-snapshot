using System.Collections.Generic;

public void SurprisinglyValid(string firstParam) { }

namespace Longer.Namespace
{
    public class Class
    {
        public int ExplicitGetterSetter
        {
            get
            {
                return _val;
            }
            set
            {
                _val = value;
            }
        }
        private int _val;

        protected virtual int ImplicitGetterSetter
        {
            get;
            set;
        }

        internal int ImplicitGetterPrivateSetter
        {
            get;
            private set;
        }

        int _speed;
        public string PublicImplicitGetterSetter { get; set; }

        public string LambdaFunction => PublicImplicitGetterSetter + " " + _speed + " speed";

        public enum Swag
        {
            Shirt,
            Sweater,
            Socks = 42,
            Pants
        }

        public Swag SourcegraphSwag;

        [Flags]
        public enum ZigFeatureSet
        {
            None = 0,
            Errors = 1,
            Comptime = 2,
            BuildSystem = 4,
            CoolCommunity = 8,
            FullPackage = Errors | Comptime | BuildSystem | CoolCommunity
        }

        public static void Syntax() {}
    }
}

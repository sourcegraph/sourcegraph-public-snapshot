#include <stdexcept>
#include "units.hpp"

namespace Sass {

  /* the conversion matrix can be readed the following way */
  /* if you go down, the factor is for the numerator (multiply) */
  /* if you go right, the factor is for the denominator (divide) */
  /* and yes, we actually use both, not sure why, but why not!? */

  const double size_conversion_factors[6][6] =
  {
             /*  in         cm         pc         mm         pt         px        */
    /* in   */ { 1,         2.54,      6,         25.4,      72,        96,       },
    /* cm   */ { 1.0/2.54,  1,         6.0/2.54,  10,        72.0/2.54, 96.0/2.54 },
    /* pc   */ { 1.0/6.0,   2.54/6.0,  1,         25.4/6.0,  72.0/6.0,  96.0/6.0  },
    /* mm   */ { 1.0/25.4,  1.0/10.0,  6.0/25.4,  1,         72.0/25.4, 96.0/25.4 },
    /* pt   */ { 1.0/72.0,  2.54/72.0, 6.0/72.0,  25.4/72.0, 1,         96.0/72.0 },
    /* px   */ { 1.0/96.0,  2.54/96.0, 6.0/96.0,  25.4/96.0, 72.0/96.0, 1,        }
  };

  const double angle_conversion_factors[4][4] =
  {
             /*  deg        grad       rad        turn      */
    /* deg  */ { 1,         40.0/36.0, PI/180.0,  1.0/360.0 },
    /* grad */ { 36.0/40.0, 1,         PI/200.0,  1.0/400.0 },
    /* rad  */ { 180.0/PI,  200.0/PI,  1,         0.5/PI    },
    /* turn */ { 360.0,     400.0,     2.0*PI,    1         }
  };

  const double time_conversion_factors[2][2] =
  {
             /*  s          ms        */
    /* s    */ { 1,         1000.0    },
    /* ms   */ { 1/1000.0,  1         }
  };
  const double frequency_conversion_factors[2][2] =
  {
             /*  Hz         kHz       */
    /* Hz   */ { 1,         1/1000.0  },
    /* kHz  */ { 1000.0,    1         }
  };
  const double resolution_conversion_factors[3][3] =
  {
             /*  dpi        dpcm       dppx     */
    /* dpi  */ { 1,         2.54,      96       },
    /* dpcm */ { 1/2.54,    1,         96/2.54  },
    /* dppx */ { 1/96.0,    2.54/96,   1        }
  };

  SassUnitType get_unit_type(SassUnit unit)
  {
    switch (unit & 0xFF00)
    {
      case SIZE: return SIZE; break;
      case ANGLE: return ANGLE; break;
      case TIME: return TIME; break;
      case FREQUENCY: return FREQUENCY; break;
      case RESOLUTION: return RESOLUTION; break;
      default: return INCOMMENSURABLE; break;
    }
  };

  SassUnit string_to_unit(const string& s)
  {
    // size units
    if      (s == "px") return PX;
    else if (s == "pt") return PT;
    else if (s == "pc") return PC;
    else if (s == "mm") return MM;
    else if (s == "cm") return CM;
    else if (s == "in") return IN;
    // angle units
    else if (s == "deg") return DEG;
    else if (s == "grad") return GRAD;
    else if (s == "rad") return RAD;
    else if (s == "turn") return TURN;
    // time units
    else if (s == "s") return SEC;
    else if (s == "ms") return MSEC;
    // frequency units
    else if (s == "Hz") return HERTZ;
    else if (s == "kHz") return KHERTZ;
    // resolutions units
    else if (s == "dpi") return DPI;
    else if (s == "dpcm") return DPCM;
    else if (s == "dppx") return DPPX;
    // for unknown units
    else return UNKNOWN;
  }

  const char* unit_to_string(SassUnit unit)
  {
    switch (unit) {
      // size units
      case PX: return "px"; break;
      case PT: return "pt"; break;
      case PC: return "pc"; break;
      case MM: return "mm"; break;
      case CM: return "cm"; break;
      case IN: return "in"; break;
      // angle units
      case DEG: return "deg"; break;
      case GRAD: return "grad"; break;
      case RAD: return "rad"; break;
      case TURN: return "turn"; break;
      // time units
      case SEC: return "s"; break;
      case MSEC: return "ms"; break;
      // frequency units
      case HERTZ: return "Hz"; break;
      case KHERTZ: return "kHz"; break;
      // resolutions units
      case DPI: return "dpi"; break;
      case DPCM: return "dpcm"; break;
      case DPPX: return "dppx"; break;
      // for unknown units
      default: return ""; break;;
    }
  }

  // throws incompatibleUnits exceptions
  double conversion_factor(const string& s1, const string& s2)
  {
    // assert for same units
    if (s1 == s2) return 1;
    // get unit enum from string
    SassUnit u1 = string_to_unit(s1);
    SassUnit u2 = string_to_unit(s2);
    // query unit group types
    SassUnitType t1 = get_unit_type(u1);
    SassUnitType t2 = get_unit_type(u2);
    // get absolute offset
    // used for array acces
    size_t i1 = u1 - t1;
    size_t i2 = u2 - t2;
    // error if units are not of the same group
    if (t1 != t2) throw incompatibleUnits(u1, u2);
    // only process known units
    if (u1 != UNKNOWN && u2 != UNKNOWN) {
      switch (t1) {
        case SIZE: return size_conversion_factors[i1][i2]; break;
        case ANGLE: return angle_conversion_factors[i1][i2]; break;
        case TIME: return time_conversion_factors[i1][i2]; break;
        case FREQUENCY: return frequency_conversion_factors[i1][i2]; break;
        case RESOLUTION: return resolution_conversion_factors[i1][i2]; break;
        // ToDo: should we throw error here?
        case INCOMMENSURABLE: return 0; break;
      }
    }
    // fallback
    return 1;
  }

}

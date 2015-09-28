#ifndef SASS_UNITS_H
#define SASS_UNITS_H

#include <cmath>
#include <string>
#include <sstream>

namespace Sass {
  using namespace std;

  const double PI = acos(-1);

  enum SassUnitType {
    SIZE = 0x000,
    ANGLE = 0x100,
    TIME = 0x200,
    FREQUENCY = 0x300,
    RESOLUTION = 0x400,
    INCOMMENSURABLE = 0x500
  };

  enum SassUnit {

    // size units
    IN = SIZE,
    CM,
    PC,
    MM,
    PT,
    PX,

    // angle units
    DEG = ANGLE,
    GRAD,
    RAD,
    TURN,

    // time units
    SEC = TIME,
    MSEC,

    // frequency units
    HERTZ = FREQUENCY,
    KHERTZ,

    // resolutions units
    DPI = RESOLUTION,
    DPCM,
    DPPX,

    // for unknown units
    UNKNOWN = INCOMMENSURABLE

  };

  extern const double size_conversion_factors[6][6];
  extern const double angle_conversion_factors[4][4];
  extern const double time_conversion_factors[2][2];
  extern const double frequency_conversion_factors[2][2];
  extern const double resolution_conversion_factors[3][3];

  SassUnit string_to_unit(const string&);
  const char* unit_to_string(SassUnit unit);
  SassUnitType get_unit_type(SassUnit unit);
  // throws incompatibleUnits exceptions
  double conversion_factor(const string&, const string&);

  class incompatibleUnits: public exception
  {
    public:
      const char* msg;
      incompatibleUnits(SassUnit a, SassUnit b)
      : exception()
      {
        stringstream ss;
        ss << "Incompatible units: ";
        ss << "'" << unit_to_string(a) << "' and ";
        ss << "'" << unit_to_string(b) << "'";
        msg = ss.str().c_str();
      };
      virtual const char* what() const throw()
      {
        return msg;
      }
  };

}

#endif

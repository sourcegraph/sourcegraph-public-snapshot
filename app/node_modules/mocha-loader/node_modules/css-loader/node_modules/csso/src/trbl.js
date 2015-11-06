function TRBL(name, imp) {
    this.name = TRBL.extractMain(name);
    this.sides = {
        'top': null,
        'right': null,
        'bottom': null,
        'left': null
    };
    this.imp = imp ? 4 : 0;
}

TRBL.props = {
    'margin': 1,
    'margin-top': 1,
    'margin-right': 1,
    'margin-bottom': 1,
    'margin-left': 1,
    'padding': 1,
    'padding-top': 1,
    'padding-right': 1,
    'padding-bottom': 1,
    'padding-left': 1
};

TRBL.extractMain = function(name) {
    var i = name.indexOf('-');
    return i === -1 ? name : name.substr(0, i);
};

TRBL.prototype.impSum = function() {
    var imp = 0, n = 0;
    for (var k in this.sides) {
        if (this.sides[k]) {
            n++;
            if (this.sides[k].imp) imp++;
        }
    }
    return imp === n ? imp : 0;
};

TRBL.prototype.add = function(name, sValue, tValue, imp) {
    var s = this.sides,
        currentSide,
        i, x, side, a = [], last,
        imp = imp ? 1 : 0,
        wasUnary = false;
    if ((i = name.lastIndexOf('-')) !== -1) {
        side = name.substr(i + 1);
        if (side in s) {
            if (!(currentSide = s[side]) || (imp && !currentSide.imp)) {
                s[side] = { s: imp ? sValue.substring(0, sValue.length - 10) : sValue, t: [tValue[0]], imp: imp };
                if (tValue[0][1] === 'unary') s[side].t.push(tValue[1]);
            }
            return true;
        }
    } else if (name === this.name) {
        for (i = 0; i < tValue.length; i++) {
            x = tValue[i];
            last = a[a.length - 1];
            switch(x[1]) {
                case 'unary':
                    a.push({ s: x[2], t: [x], imp: imp });
                    wasUnary = true;
                    break;
                case 'number':
                case 'ident':
                    if (wasUnary) {
                        last.t.push(x);
                        last.s += x[2];
                    } else {
                        a.push({ s: x[2], t: [x], imp: imp });
                    }
                    wasUnary = false;
                    break;
                case 'percentage':
                    if (wasUnary) {
                        last.t.push(x);
                        last.s += x[2][2] + '%';
                    } else {
                        a.push({ s: x[2][2] + '%', t: [x], imp: imp });
                    }
                    wasUnary = false;
                    break;
                case 'dimension':
                    if (wasUnary) {
                        last.t.push(x);
                        last.s += x[2][2] + x[3][2];
                    } else {
                        a.push({ s: x[2][2] + x[3][2], t: [x], imp: imp });
                    }
                    wasUnary = false;
                    break;
                case 's':
                case 'comment':
                case 'important':
                    break;
                default:
                    return false;
            }
        }

        if (a.length > 4) return false;

        if (!a[1]) a[1] = a[0];
        if (!a[2]) a[2] = a[0];
        if (!a[3]) a[3] = a[1];

        if (!s.top) s.top = a[0];
        if (!s.right) s.right = a[1];
        if (!s.bottom) s.bottom = a[2];
        if (!s.left) s.left = a[3];

        return true;
    }
};

TRBL.prototype.isOkToMinimize = function() {
    var s = this.sides,
        imp,
        ieReg = /\\9$/;

    if (!!(s.top && s.right && s.bottom && s.left)) {
        imp = s.top.imp + s.right.imp + s.bottom.imp + s.left.imp;

        if (ieReg.test(s.top.s) || ieReg.test(s.right.s) || ieReg.test(s.bottom.s) || ieReg.test(s.left.s)) {
            return false;
        }

        return (imp === 0 || imp === 4 || imp === this.imp);
    }
    return false;
};

TRBL.prototype.getValue = function() {
    var s = this.sides,
        a = [s.top, s.right, s.bottom, s.left],
        r = [{}, 'value'];

    if (s.left.s === s.right.s) {
        a.length--;
        if (s.bottom.s === s.top.s) {
            a.length--;
            if (s.right.s === s.top.s) {
                a.length--;
            }
        }
    }

    for (var i = 0; i < a.length - 1; i++) {
        r = r.concat(a[i].t);
        r.push([{ s: ' ' }, 's', ' ']);
    }
    r = r.concat(a[i].t);

    if (this.impSum()) r.push([{ s: '!important'}, 'important']);

    return r;
};

TRBL.prototype.getProperty = function() {
    return [{ s: this.name }, 'property', [{ s: this.name }, 'ident', this.name]];
};

TRBL.prototype.getString = function() {
    var p = this.getProperty(),
        v = this.getValue().slice(2),
        r = p[0].s + ':';

    for (var i = 0; i < v.length; i++) r += v[i][0].s;

    return r;
};

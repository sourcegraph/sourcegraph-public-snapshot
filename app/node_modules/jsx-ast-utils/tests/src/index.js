/* eslint-env mocha */
import assert from 'assert';
import core from '../../src/index';
import fs from 'fs';
import path from 'path';

const src = fs.readdirSync(path.resolve(__dirname, '../../src'))
  .filter(f => f.indexOf('.js') >= 0)
  .map(f => path.basename(f, '.js'));

describe('main export', () => {
  it('should export an object', () => {
    const expected = 'object';
    const actual = typeof core;

    assert.equal(expected, actual);
  });

  src.filter(f => f !== 'index').forEach(f => {
    it(`should export ${f}`, () => {
      assert.equal(
        core[f],
        require(path.join('../../src/', f)).default // eslint-disable-line global-require
      );
    });

    it(`should export ${f} from root`, () => {
      const fn = require(path.join('../../', f)); // eslint-disable-line global-require
      const expected = 'function';
      const actual = typeof fn;

      const expectedName = f;
      const actualName = fn.name;

      assert.equal(expected, actual);
      assert.equal(expectedName, actualName);
    });
  });
});

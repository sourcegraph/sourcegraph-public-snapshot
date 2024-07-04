const esbuild = require('esbuild');
const path = require('path');
const fs = require('fs');

function copyHtmlPlugin(options) {
  return {
    name: 'copy-html',
    setup(build) {
      build.onEnd(() => {
        fs.copyFileSync(
            path.resolve(options.src),
            path.resolve(options.dest),
        );
      });
    }
  };
}

const ctx = esbuild.context({
  entryPoints: ['./src/main.tsx'],
  bundle: true,
  outfile: './dist/bundle.js',
  tsconfig: path.resolve(__dirname, 'tsconfig.json'),
  plugins: [
    copyHtmlPlugin({
      src: './index.html',
      dest: './dist/index.html'
    })
  ],
  loader: {
    '.png': 'file'
  },
  define: {
    'process': '"10"',
    'process.env.NODE_ENV': '"development"',
    'process.env.API_ENDPOINT': '""',
    'process.env.BUILD_NUMBER': '"1.0.0"'
  }
}).then(ctx => {
    ctx.rebuild().then(() => {
        console.log('Build succeeded');
    }).catch(err => {
        console.error('Build failed', err);
        process.exit(1);
    });
    return ctx;
});

const builder = async () => {
  try {
    const buildCtx = await ctx;
    await buildCtx.dispose();
  } catch (err) {
    console.error('Build failed', err);
    process.exit(1);
  }
};

const runner = async () => {
  try {
    const buildCtx = await ctx;
    const server = await buildCtx.serve({
      host: '127.0.0.1',
      port: 8080,
      servedir: './dist',
      fallback: 'index.html'
    });
    console.log('Server running at http://127.0.0.1:8080');

    process.on('SIGINT', async () => {
      process.exit();
    });
  } catch (err) {
    console.error('Build failed', err);
    process.exit(1);
  }
};

const isRun = process.argv.includes('--run');
console.log(`Running ${isRun ? 'runner' : 'builder'}`);
const processor = isRun ? runner : builder;

processor().then(() => console.log("Done"));

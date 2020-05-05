let mix = require('laravel-mix');
require('laravel-mix-purgecss');

mix.setPublicPath('./public');
mix.js('resources/assets/js/app.js', 'public/js')
	.js('resources/assets/js/editor-create.js', 'public/js')
	.js('resources/assets/js/editor-edit.js', 'public/js')
	.sass('resources/assets/sass/app.scss', 'public/css')
	.sass('resources/assets/sass/atomic.scss', 'public/css')
	.sass('resources/assets/sass/quill.snow.scss', 'public/css')
	.copyDirectory('resources/assets/ckeditor5', 'public/ckeditor5')
	.version();

if (mix.inProduction()) {
	mix.purgeCss({
		enabled: true,

		// Your custom globs are merged with the default globs. If you need to fully replace
		// the globs, use the underlying `paths` option instead.
		globs: [
			path.join(__dirname, 'template/**/*.tmpl'),
			path.join(__dirname, 'resources/assets/js/*.js'),
			path.join(__dirname, 'node_modules/quill/**/*.js'),
			path.join(__dirname, 'node_modules/bootstrap/**/*.js'),
		],
		extensions: ['html', 'js', 'php', 'vue', 'tmpl'],
		whitelistPatterns: [/ql-*/, /video-*/],
	})
}
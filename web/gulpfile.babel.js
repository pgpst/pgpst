// gulp
import gulp             from "gulp";
import changed          from "gulp-changed";
import gutil            from "gulp-util";
import gulpNgConfig     from "gulp-ng-config";

// misc
import minimist         from "minimist";
let argv = minimist(process.argv.slice(2));
import rimraf           from "rimraf";

// webpack
import webpack          from "webpack";
import webpackDevServer from "webpack-dev-server";
import webpackConfig    from "./webpack.config";
import ngAnnotatePlugin from "ng-annotate-webpack-plugin";

// ports
const ports = {
	livereload: 35730,
	dev:        3000
};

// paths
const paths = {
	other: [
		"!src/index.html",
		"src/images/**",
		"src/fonts/**",
		"!src/**/*.js",
		"!src/**/*.coffee",
		"!src/**/*.less",
		"!src/**/*.tpl.html"
	],
	distDir: "./dist"
};

// production config
if (argv.production) {
	webpackConfig.plugins = webpackConfig.plugins.concat(
		new ngAnnotatePlugin(),
		new webpack.optimize.UglifyJsPlugin()
	);

	webpackConfig.devtool = false;
	webpackConfig.debug   = false;
}

// prod build task
let prodConfig = Object.create(webpackConfig);
prodConfig.plugins = webpackConfig.plugins.concat(
	new webpack.DefinePlugin({
		"process-env": {
			"NODE_ENV": JSON.stringify("production")
		}
	}),
	new webpack.optimize.DedupePlugin(),
	new webpack.optimize.UglifyJsPlugin()
);
gulp.task("webpack:build", (done) => {
	webpack(prodConfig, (err, stats) => {
		if (err) {
			throw new gutil.PluginError("webpack:build", err);
		}

		gutil.log("[webpack:build]", stats.toString({colors:true}));

		done();
	});
});

// dev server task
let devConfig = Object.create(webpackConfig);
devConfig.devtool = "eval";
devConfig.debug   = true;
gulp.task("webpack-dev-server", (cb) => {
	new webpackDevServer(webpack(devConfig), {
		contentBase: "./dist",
		quiet:       false,
		noInfo:      false,
		lazy:        false,
		watchDelay:  300,
		stats: {
			colors: true,
		},
		historyApiFallback: true,
	}).listen(ports.dev, "0.0.0.0", (err) => {
		if (err) {
			throw new gutil.PluginError("webpack-dev-server", err);
		}

		gutil.log("[webpack-dev-server]", "http://localhost:" + ports.dev);
	});
});

// move changed files
gulp.task("other", () => {
	gulp.src(paths.other)
		.pipe(changed(paths.distDir))
		.pipe(gulp.dest(paths.distDir));
});

// clears dist directory
gulp.task("clean", function () {
	rimraf.sync(paths.distDir, {}, gutil.log);
});

// build task
gulp.task("build", [
	"clean",
	"webpack:build",
	"other"
]);

// watch for changes
gulp.task("watch", ["clean", "other"], () => {
	webpack(devConfig)
		.watch(200, function (err, stats) {
			if (err) {
				throw new gutil.PluginError("webpack", err);
			}

			gutil.log("[webpack]", stats.toString({
				colors: true
			}));
		});

	gulp.watch(paths.other, ["other"]);
});

gulp.task("serve", ["webpack-dev-server", "watch"]);
gulp.task("default", ["build"]);
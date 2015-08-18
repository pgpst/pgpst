import path    from "path";
import webpack from "webpack";
import ngminPlugin from "ngmin-webpack-plugin";
import htmlWebpackPlugin from "html-webpack-plugin";
import extractTextPlugin from "extract-text-webpack-plugin";

let appRoot   = path.join(__dirname, "src");

export default {
	cache: true,
	debug: true,

	entry: [
		path.join(appRoot, "/app.js")
	],

	output: {
		path: path.join(__dirname, "./dist"),
		publicPath: "./",
		libraryTarget: "var",
		filename: "[hash].bundle.js",
		chunkFilename: "[chunkhash].js"
	},

	module: {
		loaders: [
			{
				test: /\.js$/,
				exclude: /(node_modules)/,
				loader: "babel?optional[]=runtime"
			},
			{
				test: /\.css$/,
				loaders: ["style", "css"]
			},
			{
				test: /\.less$/,
				loader: "style!css!less"
			},
			{
				// partials
				test: /\.html$/,
				exclude: /(node_modules)/,
				loader: "ng-cache"
			},
			{
				// fontawesome icons
				test: /\.(woff|woff2)(\?(.*))?$/,
				loader: "url?prefix=factorynts/&limit=5000&mimetype=application/font-woff"
			},
			{
				test: /\.ttf(\?(.*))?$/,
				loader: "file?prefix=fonts/"
			},
			{
				test: /\.eot(\?(.*))?$/,
				loader: "file?prefix=fonts/"
			},
			{
				test: /\.svg(\?(.*))?$/,
				loader: "file?prefix=fonts/"
			},
			{
				test: /\.json$/,
				loader: "json"
			}
		],

		extensions: [
			"",
			".js",
			".less",
			".css"
		],

		root: [appRoot],
	},

	singleRun: true,

	plugins: [
		new htmlWebpackPlugin({
			template: __dirname + "/src/index.html"
		}),
		new extractTextPlugin("[name].css"),
		new webpack.IgnorePlugin(new RegExp("^(node-localstorage)$")),
		new webpack.DefinePlugin({
			"process.env": () => {
				let result = {};

				for (let key in process.env) {
					result[key] = JSON.stringify(process.env[key]);
				}

				return result;
			}()
/*
			Object.keys(process.env).reduce((previous, current) => {
				if (!previous) {
					previous = {};
				}

				previous[current] = JSON.stringify(process.env[current]);
			})*/
		})
	],

	devtool: "eval"
}
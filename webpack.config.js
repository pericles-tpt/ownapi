const path = require('path');

module.exports = {
  mode: 'development',
  entry: {
    index: './src/typescript/index.ts'
  },
  devtool: 'source-map',
  module: {
      rules: [
        {
          test: /\.s[ac]ss$/i,
          use: [
            // Creates `style` nodes from JS strings
            "style-loader",
            // Translates CSS into CommonJS
            "css-loader",
            // Compiles Sass to CSS
            "sass-loader",
          ],
        },
        {
          test: /\.tsx?$/,
          use: 'ts-loader',
          exclude: /node_modules/,
        },
      ],
  },
  resolve: {
    extensions: ['.tsx', '.ts', '.js', '.html', '.css'],
  },
  output: {
    filename: '[name].js',
    path: path.resolve(__dirname, './dist'),
  },
};

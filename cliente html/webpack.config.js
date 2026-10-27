// webpack.config.js
const path = require('path');

module.exports = {
  // El punto de inicio de tu código
  entry: './cliente_receptor_fragmentos.js', 
  mode: 'development',
  output: {
    // El archivo de salida final para el navegador
    filename: 'bundle.js', 
    path: path.resolve(__dirname, './'),
  },
};

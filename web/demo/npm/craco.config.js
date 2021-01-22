const CracoLessPlugin = require('craco-less');

module.exports = {
  plugins: [
    {
      plugin: CracoLessPlugin,
      options: {
        lessLoaderOptions: {
          lessOptions: {
            modifyVars: {
              '@primary-color': '#1890ff',
              '@font-size-base': '10px',
              '@margin-lg': '18px', // containers
              '@margin-md': '12px', // small containers and buttons
              '@margin-sm': '8px', // Form controls and items
              '@margin-xs': '5px', // small items
              '@margin-xss': '2px', // more small
              '@height-base': '24px',
              '@height-lg': '32px',
              '@height-sm': '18px',
            },
            javascriptEnabled: true,
          },
        },
      },
    },
  ],
};

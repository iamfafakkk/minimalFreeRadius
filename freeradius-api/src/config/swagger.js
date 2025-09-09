const swaggerUi = require('swagger-ui-express');
const path = require('path');
const fs = require('fs');

// Load swagger.json
const swaggerDocument = JSON.parse(
  fs.readFileSync(path.join(__dirname, '../../swagger.json'), 'utf8')
);

// Swagger UI options
const swaggerOptions = {
  explorer: true,
  swaggerOptions: {
    docExpansion: 'list',
    filter: true,
    showRequestHeaders: true,
    showCommonExtensions: true,
    tryItOutEnabled: true,
    requestInterceptor: (req) => {
      // Add custom headers or modify requests
      console.log('Swagger Request:', req.url);
      return req;
    },
    responseInterceptor: (res) => {
      // Handle responses
      console.log('Swagger Response:', res.status);
      return res;
    }
  },
  customCss: `
    .swagger-ui .topbar { 
      background-color: #2c3e50; 
    }
    .swagger-ui .topbar .download-url-wrapper .select-label {
      color: #fff;
    }
    .swagger-ui .info .title {
      color: #2c3e50;
    }
    .swagger-ui .scheme-container {
      background: #f8f9fa;
      border: 1px solid #dee2e6;
      border-radius: 4px;
      padding: 10px;
      margin: 10px 0;
    }
    .swagger-ui .auth-wrapper {
      background: #e9ecef;
      border: 1px solid #ced4da;
      border-radius: 4px;
      padding: 10px;
      margin: 10px 0;
    }
    .swagger-ui .btn.authorize {
      background-color: #28a745;
      border-color: #28a745;
    }
    .swagger-ui .btn.authorize:hover {
      background-color: #218838;
      border-color: #1e7e34;
    }
    .swagger-ui .opblock.opblock-post {
      border-color: #49cc90;
      background: rgba(73, 204, 144, 0.1);
    }
    .swagger-ui .opblock.opblock-get {
      border-color: #61affe;
      background: rgba(97, 175, 254, 0.1);
    }
    .swagger-ui .opblock.opblock-put {
      border-color: #fca130;
      background: rgba(252, 161, 48, 0.1);
    }
    .swagger-ui .opblock.opblock-delete {
      border-color: #f93e3e;
      background: rgba(249, 62, 62, 0.1);
    }
    .swagger-ui .opblock-summary-method {
      font-weight: bold;
      text-transform: uppercase;
    }
    .swagger-ui .parameter__name {
      font-weight: bold;
    }
    .swagger-ui .response-col_status {
      font-weight: bold;
    }
    .swagger-ui .model-title {
      font-weight: bold;
      color: #2c3e50;
    }
    .swagger-ui .prop-type {
      color: #6c757d;
    }
    .swagger-ui .prop-format {
      color: #6c757d;
      font-style: italic;
    }
  `,
  customSiteTitle: 'FreeRADIUS API Documentation',
  customfavIcon: '/favicon.ico',
  customJs: [
    'https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js',
    'https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js'
  ]
};

// Custom HTML template
const customHtml = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>FreeRADIUS API Documentation</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
  <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@5.9.0/favicon-32x32.png" sizes="32x32" />
  <style>
    .custom-header {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: white;
      padding: 20px;
      text-align: center;
      margin-bottom: 20px;
    }
    .custom-header h1 {
      margin: 0;
      font-size: 2.5em;
      font-weight: 300;
    }
    .custom-header p {
      margin: 10px 0 0 0;
      font-size: 1.2em;
      opacity: 0.9;
    }
    .api-info {
      background: white;
      padding: 20px;
      margin: 20px;
      border-radius: 8px;
      box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    }
    .api-info h3 {
      color: #2c3e50;
      margin-top: 0;
    }
    .quick-links {
      display: flex;
      justify-content: center;
      gap: 15px;
      margin: 20px;
      flex-wrap: wrap;
    }
    .quick-link {
      background: #3498db;
      color: white;
      padding: 10px 20px;
      border-radius: 5px;
      text-decoration: none;
      font-weight: bold;
      transition: background 0.3s;
    }
    .quick-link:hover {
      background: #2980b9;
      color: white;
      text-decoration: none;
    }
    .footer {
      text-align: center;
      padding: 20px;
      color: #7f8c8d;
      border-top: 1px solid #ecf0f1;
      margin-top: 40px;
    }
  </style>
</head>
<body>
  <div class="custom-header">
    <h1>🛡️ FreeRADIUS API</h1>
    <p>Comprehensive REST API for FreeRADIUS Management</p>
  </div>
  
  <div class="api-info">
    <h3>📋 Quick Start</h3>
    <p>Welcome to the FreeRADIUS API documentation. This API provides complete management capabilities for FreeRADIUS server.</p>
    <ul>
      <li><strong>Base URL:</strong> <code>http://localhost:3000/api/v1</code></li>
      <li><strong>Authentication:</strong> JWT Bearer token required for most endpoints</li>
      <li><strong>Content-Type:</strong> <code>application/json</code></li>
    </ul>
  </div>
  
  <div class="quick-links">
    <a href="#/Authentication" class="quick-link">🔐 Authentication</a>
    <a href="#/NAS%20Management" class="quick-link">🖥️ NAS Management</a>
    <a href="#/User%20Management" class="quick-link">👥 User Management</a>
    <a href="#/System" class="quick-link">💚 System Health</a>
  </div>
  
  <div id="swagger-ui"></div>
  
  <div class="footer">
    <p>FreeRADIUS API v1.0.0 | Built with ❤️ using Node.js & Express</p>
    <p>For support and documentation, visit our <a href="/docs/API_DOCUMENTATION.md">API Documentation</a></p>
  </div>
  
  <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      const ui = SwaggerUIBundle({
        url: '/swagger.json',
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout",
        validatorUrl: null,
        tryItOutEnabled: true,
        supportedSubmitMethods: ['get', 'post', 'put', 'delete', 'patch'],
        docExpansion: 'list',
        filter: true,
        showRequestHeaders: true,
        showCommonExtensions: true,
        displayRequestDuration: true,
        defaultModelsExpandDepth: 1,
        defaultModelExpandDepth: 1,
        requestInterceptor: function(request) {
          console.log('Request:', request);
          return request;
        },
        responseInterceptor: function(response) {
          console.log('Response:', response);
          return response;
        }
      });
      window.ui = ui;
    };
  </script>
</body>
</html>
`;

module.exports = {
  swaggerDocument,
  swaggerOptions,
  swaggerUi,
  customHtml
};